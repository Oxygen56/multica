// Package notifications implements the cross-squad collaboration notification
// system (Feature #4596). It provides:
//
//   - A rule engine (R1–R5) that fires targeted notifications across issue
//     boundaries based on metadata patterns, not hardcoded logic.
//   - Anomaly detectors (D1–D6) that run as a side effect of event processing.
//   - Event persistence for reliability and gap detection.
//
// Architecture: the rule engine layers on top of events.Bus. It subscribes
// to domain events (status_changed, comment_added, metadata_changed, etc.),
// evaluates rules by matching metadata patterns, and publishes synthesized
// events (child:terminal, stage:closed, notification:delivered, anomaly:detected)
// back onto the bus. Downstream listeners (inbox, WS broadcast) consume these
// synthesized events the same way they consume raw domain events.
//
// Priority-based suppression: when two rules target the same (source, target)
// pair, the higher-priority rule (lower numeric value) suppresses the
// lower-priority rule's comment action. Metadata updates from both rules
// still apply. Cross-target rules never suppress each other.
package notifications

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/multica-ai/multica/server/internal/events"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

// RulePriority assigns each rule a numeric priority. Lower number = higher
// priority. When two rules target the same (source, target) pair for a
// comment action, the higher-priority rule suppresses the lower one.
// Metadata updates from both rules always apply.
const (
	PriorityR5 = 10 // dispatch contract callback
	PriorityR2 = 20 // review completion
	PriorityR4 = 30 // stage completion
	PriorityR3 = 40 // blocking removal
	PriorityR1 = 50 // child task completion
)

// Rule represents a single notification rule. Each rule matches on metadata
// patterns (not hardcoded posture values) so future posture additions don't
// require rule engine changes.
type Rule struct {
	ID       string // "R1", "R2", etc.
	Priority int    // lower = higher priority
	// Match evaluates whether this rule should fire for the given trigger context.
	// Returns the target issue ID and a payload for the notification, or nil if
	// the rule does not match.
	Match func(ctx *TriggerContext) *MatchResult
	// Action executes the rule's side effects (creating inbox items, posting
	// system comments, etc.).
	Action func(ctx context.Context, deps RuleDeps, result *MatchResult) error
}

// TriggerContext carries all the information a rule needs to evaluate
// whether it should fire — the triggering event, the affected issue, and
// its current metadata.
type TriggerContext struct {
	Event    events.Event
	Issue    db.Issue
	Metadata map[string]string // current metadata KV for the issue
}

// MatchResult captures what a rule matched and what it should do about it.
type MatchResult struct {
	RuleID     string
	Priority   int
	TargetType string // "agent", "squad", "member"
	TargetID   string
	IssueID    string // the issue the notification points to
	NotifType  string
	Title      string
	Body       string
	Details    map[string]any
	// SuppressibleBy is the set of rule IDs that can suppress this match's
	// comment action. Populated by the engine based on priority ordering.
	SuppressibleBy []string
}

// RuleDeps provides the rule engine's action phase with access to shared
// infrastructure: database queries, the event bus, and the handler layer.
type RuleDeps struct {
	Queries *db.Queries
	Bus     *events.Bus
	// NotifyFn is a function that creates an inbox notification for a
	// target. Set by the caller (cmd/server) to wire into the existing
	// notifyDirect / notifySubscribers infrastructure.
	NotifyFn func(ctx context.Context, result *MatchResult) error
	// RecordNotificationFn persists a notification_records row.
	RecordNotificationFn func(ctx context.Context, result *MatchResult, status string) error
}

// Engine is the notification rule engine. It subscribes to the event bus,
// evaluates registered rules on each relevant event, and executes matched
// rules with priority-based suppression.
type Engine struct {
	mu       sync.RWMutex
	rules    []Rule            // ordered by priority (ascending)
	bus      *events.Bus
	queries  *db.Queries
	deps     RuleDeps
	persist  *EventPersister
	anomaly  *AnomalyDetector
}

// NewEngine creates a new notification rule engine.
func NewEngine(bus *events.Bus, queries *db.Queries, deps RuleDeps) *Engine {
	e := &Engine{
		rules:   make([]Rule, 0, 5),
		bus:     bus,
		queries: queries,
		deps:    deps,
	}
	e.persist = NewEventPersister(queries)
	e.anomaly = NewAnomalyDetector(bus, queries)
	return e
}

// RegisterRule adds a rule to the engine. Rules are evaluated in priority
// order (lowest number first). Call before Start.
func (e *Engine) RegisterRule(r Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.rules = append(e.rules, r)
}

// Start wires the engine into the event bus. It subscribes to the domain
// events that can trigger notification rules and anomaly detectors.
func (e *Engine) Start() {
	ctx := context.Background()

	// issue:updated with status change → evaluate R1, R3, R4
	e.bus.Subscribe(protocol.EventIssueUpdated, func(ev events.Event) {
		payload, ok := ev.Payload.(map[string]any)
		if !ok {
			return
		}
		statusChanged, _ := payload["status_changed"].(bool)
		if !statusChanged {
			return
		}
		issue, ok := payload["issue"].(db.Issue)
		if !ok {
			// Try to extract from IssueResponse
			if issueResp, ok2 := payload["issue"].(map[string]any); ok2 {
				issueID, _ := issueResp["id"].(string)
				if issueID == "" {
					return
				}
				dbIssue, err := e.queries.GetIssue(ctx, util.MustParseUUID(issueID))
				if err != nil {
					return
				}
				issue = dbIssue
			} else {
				return
			}
		}
		e.evaluate(ctx, ev, issue)
	})

	// issue_metadata:changed → re-evaluate rules that key on metadata
	e.bus.Subscribe(protocol.EventIssueMetadataChanged, func(ev events.Event) {
		payload, ok := ev.Payload.(map[string]any)
		if !ok {
			return
		}
		issueID, _ := payload["issue_id"].(string)
		if issueID == "" {
			return
		}
		issue, err := e.queries.GetIssue(ctx, util.MustParseUUID(issueID))
		if err != nil {
			return
		}
		e.evaluate(ctx, ev, issue)
	})

	// child:terminal → drive anomaly detection D1 (parent-child deadlock)
	e.bus.Subscribe(protocol.EventChildTerminal, func(ev events.Event) {
		e.anomaly.OnChildTerminal(ctx, ev)
	})

	// stage:closed → drive anomaly detection D4 (stage stall)
	e.bus.Subscribe(protocol.EventStageClosed, func(ev events.Event) {
		e.anomaly.OnStageClosed(ctx, ev)
	})

	slog.Info("notification rule engine started", "rules", len(e.rules))
}

// evaluate runs all registered rules against a trigger context, collects
// matches, applies priority-based suppression, and executes actions.
func (e *Engine) evaluate(ctx context.Context, ev events.Event, issue db.Issue) {
	// Parse metadata from the issue's JSONB column.
	metaMap := parseMetadataJSON(issue.Metadata)

	tctx := &TriggerContext{
		Event:    ev,
		Issue:    issue,
		Metadata: metaMap,
	}

	// Persist the triggering event for reliability (Module E).
	e.persist.Record(ctx, ev.WorkspaceID, ev.Type, ev.Payload)

	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()

	var matches []*MatchResult
	for _, rule := range rules {
		m := rule.Match(tctx)
		if m != nil {
			m.RuleID = rule.ID
			m.Priority = rule.Priority
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		return
	}

	// Apply priority-based suppression across matches that target the same
	// (source, target) pair.
	applyPrioritySuppression(matches)

	// Execute actions for non-suppressed matches.
	for _, m := range matches {
		// Find the rule for this match.
		var rule *Rule
		for i := range rules {
			if rules[i].ID == m.RuleID {
				rule = &rules[i]
				break
			}
		}
		if rule == nil {
			continue
		}
		actionErr := rule.Action(ctx, e.deps, m)
		if actionErr != nil {
			slog.Error("notification rule action failed",
				"rule", m.RuleID, "target", m.TargetID, "error", actionErr)
		}

		// Record notification delivery.
		status := "delivered"
		if actionErr != nil {
			status = "failed"
		}
		if e.deps.RecordNotificationFn != nil {
			if err := e.deps.RecordNotificationFn(ctx, m, status); err != nil {
				slog.Warn("notification engine: failed to record notification",
					"rule", m.RuleID, "error", err)
			}
		}
	}

	// Run anomaly detectors as a side effect (Module D).
	e.anomaly.Evaluate(ctx, tctx, matches)
}

// suppressionKey returns a key that groups matches by their (source, target) pair.
func suppressionKey(m *MatchResult) string {
	return m.IssueID + "::" + m.TargetID
}

// applyPrioritySuppression marks lower-priority matches as suppressed when
// a higher-priority match targets the same (source, target) pair.
func applyPrioritySuppression(matches []*MatchResult) {
	// Group matches by (source, target) key.
	groups := map[string][]*MatchResult{}
	for _, m := range matches {
		key := suppressionKey(m)
		groups[key] = append(groups[key], m)
	}

	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}
		// Find the highest-priority match (lowest number).
		best := group[0]
		for _, m := range group[1:] {
			if m.Priority < best.Priority {
				best = m
			}
		}
		// Suppress lower-priority matches.
		for _, m := range group {
			if m != best {
				m.SuppressibleBy = append(m.SuppressibleBy, best.RuleID)
			}
		}
	}
}

// IssueCheckFn is a callback used by anomaly detectors to look up related
// issues without depending on the handler layer directly.
type IssueCheckFn func(ctx context.Context, issueID string) (db.Issue, error)

// CommentPostFn is a callback used by repair actions to post system comments.
type CommentPostFn func(ctx context.Context, workspaceID, issueID, content string) error

// parseMetadataJSON parses the issue's JSONB metadata into a string map.
// The issue table stores metadata as JSONB ([]byte). Non-string values
// are skipped — metadata rule matching only operates on string keys.
func parseMetadataJSON(raw []byte) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		slog.Warn("notification engine: failed to parse metadata JSON", "error", err)
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}
