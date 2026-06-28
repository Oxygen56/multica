package events

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/multica-ai/multica/server/pkg/protocol"
)

// NotificationRule defines a single notification rule in the engine.
// Rules are evaluated in priority order (lower Priority = higher precedence).
type NotificationRule struct {
	// Name is a stable identifier (e.g. "R5_contract_callback").
	Name string
	// Priority — lower runs first. When two rules target the same issue,
	// the higher-priority rule's comment suppresses lower-priority ones
	// unless the lower rule has additional notify targets.
	Priority int
	// TriggerEventTypes is the set of event types that activate this rule.
	TriggerEventTypes []string
	// Evaluate returns (shouldFire bool, notifyTargets []NotifyTarget, message string).
	// The rule implementation queries metadata / DB as needed.
	Evaluate func(e Event) (bool, []NotifyTarget, string)
}

// NotifyTarget describes who should receive a notification.
type NotifyTarget struct {
	TargetType string // "agent", "squad", "member"
	TargetID   string
	IssueID    string // which issue to notify on

	// Embeddable fields for building the notification message.
	IssueTitle      string
	IssueIdentifier string
}

// NotificationRecorder persists notification deliveries to the
// notification_records table for idempotency and pull-mode queries.
type NotificationRecorder interface {
	CreateNotificationRecord(targetType, targetID, sourceIssueID, notifType, message string) (string, error)
	FindDuplicateNotification(targetType, targetID, sourceIssueID, notifType string) (bool, error)
}

// noopRecorder is used when the recorder dependency is not wired in yet.
type noopRecorder struct{}

func (n *noopRecorder) CreateNotificationRecord(_, _, _, _, _ string) (string, error) {
	return "", nil
}
func (n *noopRecorder) FindDuplicateNotification(_, _, _, _ string) (bool, error) {
	return false, nil
}

// NotificationBus is a pure event-driven notification rule engine.
// It sits on top of the existing events.Bus and listens for issue
// status/completion events, then evaluates registered rules and
// fires notification actions.
type NotificationBus struct {
	bus      *Bus
	rules    []NotificationRule
	recorder NotificationRecorder
	onNotify func(Event, NotifyTarget, string) // post-notification callback
}

// NewNotificationBus creates a new notification bus.
func NewNotificationBus(bus *Bus) *NotificationBus {
	return &NotificationBus{
		bus:      bus,
		recorder: &noopRecorder{},
	}
}

// SetRecorder sets the notification recorder for persistence.
func (nb *NotificationBus) SetRecorder(r NotificationRecorder) {
	nb.recorder = r
}

// SetOnNotify sets the callback invoked after each notification target
// is determined, so the caller can post comments, dispatch mentions, etc.
func (nb *NotificationBus) SetOnNotify(fn func(Event, NotifyTarget, string)) {
	nb.onNotify = fn
}

// AddRule registers a notification rule. Rules are sorted by priority
// before evaluation.
func (nb *NotificationBus) AddRule(rule NotificationRule) {
	nb.rules = append(nb.rules, rule)
}

// Start wires the notification bus into the underlying events.Bus.
// It subscribes to the event types needed to fire the registered rules.
func (nb *NotificationBus) Start() {
	// Collect all trigger event types from registered rules.
	eventTypes := map[string]bool{}
	for _, r := range nb.rules {
		for _, t := range r.TriggerEventTypes {
			eventTypes[t] = true
		}
	}

	for et := range eventTypes {
		et := et
		nb.bus.Subscribe(et, func(e Event) {
			nb.evaluate(e)
		})
	}
	slog.Info("notification bus: started", "rules", len(nb.rules), "event_types", len(eventTypes))
}

// evaluate runs all rules against the event in priority order, with
// comment suppression for lower-priority rules targeting the same issue.
func (nb *NotificationBus) evaluate(e Event) {
	if len(nb.rules) == 0 {
		return
	}

	// Sort rules by priority (stable — insertion order for ties).
	// Rules are stored in insertion order; we scan in priority order.
	type ruleResult struct {
		rule    NotificationRule
		targets []NotifyTarget
		message string
	}

	var matched []ruleResult
	for i := range nb.rules {
		r := &nb.rules[i]
		// Check if this rule's trigger types match the event.
		triggered := false
		for _, t := range r.TriggerEventTypes {
			if t == e.Type {
				triggered = true
				break
			}
		}
		if !triggered {
			continue
		}

		shouldFire, targets, message := r.Evaluate(e)
		if !shouldFire || len(targets) == 0 {
			continue
		}
		matched = append(matched, ruleResult{rule: *r, targets: targets, message: message})
	}

	// Sort matched: priority ascending; insertion order for ties.
	for i := 0; i < len(matched); i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[j].rule.Priority < matched[i].rule.Priority {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}

	// Suppress lower-priority rule comments on the same target issue.
	// Track which (targetIssue, targetType, targetID) tuples have been notified.
	notified := map[string]bool{}

	for _, mr := range matched {
		for _, t := range mr.targets {
			key := fmt.Sprintf("%s:%s:%s", t.IssueID, t.TargetType, t.TargetID)

			// Check idempotency via recorder.
			if dup, _ := nb.recorder.FindDuplicateNotification(t.TargetType, t.TargetID, t.IssueID, t.IssueIdentifier); dup {
				continue
			}

			if notified[key] {
				continue
			}
			notified[key] = true

			// Record the notification.
			if _, err := nb.recorder.CreateNotificationRecord(t.TargetType, t.TargetID, t.IssueID, mr.rule.Name, mr.message); err != nil {
				slog.Warn("notification bus: failed to record notification",
					"rule", mr.rule.Name,
					"target_type", t.TargetType,
					"target_id", t.TargetID,
					"error", err,
				)
			}

			// Fire the notification callback.
			if nb.onNotify != nil {
				nb.onNotify(e, t, mr.message)
			}
		}
	}
}

// ── Built-in rule factories ──────────────────────────────────────────────

// MetadataReader is the interface the rules need to read issue metadata.
type MetadataReader interface {
	GetIssueMetadata(issueID string) (map[string]any, error)
	GetIssue(issueID string) (issueStatus, parentID, assigneeType, assigneeID string, err error)
	GetChildIssues(parentID string) ([]ChildInfo, error)
}

// ChildInfo is the minimal child issue information needed by rules.
type ChildInfo struct {
	ID     string
	Status string
	Stage  int32
}

// issueInfo is an in-memory representation for rules.
type issueInfo struct {
	Status       string
	ParentID     string
	AssigneeType string
	AssigneeID   string
}

// ── R1: Child Terminal ────────────────────────────────────────────────────
// Priority 100 — fires when a child issue transitions to done/cancelled
// and has a parent. Posts a system comment on the parent issue.

func MakeR1ChildTerminalRule(reader MetadataReader) NotificationRule {
	return NotificationRule{
		Name:     "R1_child_terminal",
		Priority: 100,
		TriggerEventTypes: []string{
			protocol.EventIssueChildTerminal,
			protocol.EventIssueUpdated, // broad catch; evaluation narrows
		},
		Evaluate: func(e Event) (bool, []NotifyTarget, string) {
			// Extract issue info from event payload.
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return false, nil, ""
			}

			// Check if this is a status change to terminal.
			statusChanged, _ := payload["status_changed"].(bool)
			if !statusChanged {
				return false, nil, ""
			}

			issue, _ := payload["issue"].(map[string]any)
			parentID, _ := issue["parent_issue_id"].(string)
			if parentID == "" {
				return false, nil, ""
			}

			newStatus, _ := issue["status"].(string)
			if newStatus != "done" && newStatus != "cancelled" {
				return false, nil, ""
			}

			// R1: post a system comment on the parent.
			issueTitle, _ := issue["title"].(string)
			issueID, _ := issue["id"].(string)
			assigneeType, _ := issue["assignee_type"].(string)
			assigneeID, _ := issue["assignee_id"].(string)

			var targets []NotifyTarget
			if assigneeType != "" && assigneeID != "" {
				targets = append(targets, NotifyTarget{
					TargetType:      assigneeType,
					TargetID:        assigneeID,
					IssueID:         parentID,
					IssueTitle:      issueTitle,
					IssueIdentifier: issueID,
				})
			}

			msg := fmt.Sprintf("Child issue completed: %s", issueTitle)
			return len(targets) > 0, targets, msg
		},
	}
}

// ── R2: Review Completed ──────────────────────────────────────────────────
// Priority 20 — fires when metadata review_target is set and the reviewing
// issue enters in_review status. Notifies the reviewed issue's assignee.

func MakeR2ReviewCompletedRule(reader MetadataReader) NotificationRule {
	return NotificationRule{
		Name:     "R2_review_completed",
		Priority: 20,
		TriggerEventTypes: []string{
			protocol.EventIssueUpdated,
			protocol.EventIssueMetadataChanged,
		},
		Evaluate: func(e Event) (bool, []NotifyTarget, string) {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return false, nil, ""
			}

			// Check if this is a status change into in_review.
			issue, _ := payload["issue"].(map[string]any)
			if issue == nil {
				return false, nil, ""
			}

			newStatus, _ := issue["status"].(string)
			if newStatus != "in_review" {
				return false, nil, ""
			}

			// Check metadata for review_target.
			metadata, _ := payload["metadata"].(map[string]any)
			if metadata == nil {
				// Try reading from issue metadata field.
				if metaField, ok := issue["metadata"].(map[string]any); ok {
					metadata = metaField
				}
			}

			reviewTarget, ok := metadata["review_target"].(string)
			if !ok || reviewTarget == "" {
				return false, nil, ""
			}

			// We have a review_target — notify the target issue's assignee.
			issueID, _ := issue["id"].(string)
			issueTitle, _ := issue["title"].(string)

			target := NotifyTarget{
				TargetType:      "agent",
				TargetID:        "", // will be resolved by caller
				IssueID:         reviewTarget,
				IssueTitle:      issueTitle,
				IssueIdentifier: issueID,
			}

			msg := fmt.Sprintf("Review completed for [%s](mention://issue/%s). The review results are on the review issue.",
				issueTitle, issueID)

			return true, []NotifyTarget{target}, msg
		},
	}
}

// ── R3: Blocking Resolved ──────────────────────────────────────────────────
// Priority 80 — fires when an issue transitions FROM blocked TO a non-blocked
// status. Looks up who is waiting on this issue and notifies them.

func MakeR3BlockingResolvedRule(reader MetadataReader) NotificationRule {
	return NotificationRule{
		Name:     "R3_blocking_resolved",
		Priority: 80,
		TriggerEventTypes: []string{
			protocol.EventIssueUnblocked,
			protocol.EventIssueUpdated,
		},
		Evaluate: func(e Event) (bool, []NotifyTarget, string) {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return false, nil, ""
			}

			statusChanged, _ := payload["status_changed"].(bool)
			if !statusChanged {
				return false, nil, ""
			}

			prevStatus, _ := payload["prev_status"].(string)
			if prevStatus != "blocked" {
				return false, nil, ""
			}

			issue, _ := payload["issue"].(map[string]any)
			newStatus, _ := issue["status"].(string)
			if newStatus == "blocked" || newStatus == "cancelled" || newStatus == "done" {
				return false, nil, ""
			}

			// This issue was unblocked. Look up who's waiting on it.
			// In the full implementation this queries the metadata system
			// for issues with waiting_on_issue containing this issue's ID.
			// For now, return a placeholder — the caller hooks into the
			// full DB-query path.
			issueTitle, _ := issue["title"].(string)

			msg := fmt.Sprintf("Blocking issue resolved: %s", issueTitle)
			return false, []NotifyTarget{}, msg // caller resolves waiters
		},
	}
}

// ── R4: Stage Completed ────────────────────────────────────────────────────
// Priority 50 — fires when all children in a stage are terminal.
// This is the enhanced version of notifyParentOfChildDone.

func MakeR4StageCompletedRule(reader MetadataReader) NotificationRule {
	return NotificationRule{
		Name:     "R4_stage_completed",
		Priority: 50,
		TriggerEventTypes: []string{
			protocol.EventIssueStageCompleted,
		},
		Evaluate: func(e Event) (bool, []NotifyTarget, string) {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return false, nil, ""
			}

			parentID, _ := payload["parent_id"].(string)
			stage, _ := payload["stage"].(int32)
			if parentID == "" {
				return false, nil, ""
			}

			// Notify the parent assignee.
			assigneeType, _ := payload["parent_assignee_type"].(string)
			assigneeID, _ := payload["parent_assignee_id"].(string)

			if assigneeType == "" || assigneeID == "" {
				return false, nil, ""
			}

			msg := fmt.Sprintf("Stage %d completed on parent issue.", stage)
			target := NotifyTarget{
				TargetType: assigneeType,
				TargetID:   assigneeID,
				IssueID:    parentID,
			}

			return true, []NotifyTarget{target}, msg
		},
	}
}

// ── R5: Contract Callback ──────────────────────────────────────────────────
// Priority 10 — fires when a dispatch contract's to_issue reaches terminal
// status. Executes the callback per the contract configuration.

func MakeR5ContractCallbackRule(reader MetadataReader) NotificationRule {
	return NotificationRule{
		Name:     "R5_contract_callback",
		Priority: 10,
		TriggerEventTypes: []string{
			protocol.EventIssueChildTerminal,
			protocol.EventIssueUpdated,
		},
		Evaluate: func(e Event) (bool, []NotifyTarget, string) {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return false, nil, ""
			}

			statusChanged, _ := payload["status_changed"].(bool)
			if !statusChanged {
				return false, nil, ""
			}

			issue, _ := payload["issue"].(map[string]any)
			newStatus, _ := issue["status"].(string)
			if newStatus != "done" && newStatus != "cancelled" {
				return false, nil, ""
			}

			// Check metadata for dispatch_contract_id.
			metadata, _ := issue["metadata"].(map[string]any)
			contractID, _ := metadata["dispatch_contract_id"].(string)
			if contractID == "" {
				return false, nil, ""
			}

			callbackTarget, _ := metadata["callback_target"].(string)
			if callbackTarget == "" {
				return false, nil, ""
			}

			issueID, _ := issue["id"].(string)
			issueTitle, _ := issue["title"].(string)

			msg := fmt.Sprintf("Dispatch contract %s fulfilled: child [%s](mention://issue/%s) completed.",
				contractID, issueTitle, issueID)

			target := NotifyTarget{
				TargetType: "agent",
				IssueID:    callbackTarget,
			}

			return true, []NotifyTarget{target}, msg
		},
	}
}

// ── Anomaly Detectors ──────────────────────────────────────────────────────

// AnomalyDetector is a pure event-driven check. It fires when specific
// events arrive and evaluates whether a detectable anomaly exists.
type AnomalyDetector struct {
	Name       string
	EventTypes []string
	Evaluate   func(e Event) []AnomalyFinding
}

// AnomalyFinding describes a detected anomaly.
type AnomalyFinding struct {
	Severity    string // "L1", "L2", "L3", "L4", "L5"
	IssueID     string
	Title       string
	Description string
	Action      string // "metadata_update", "comment", "mention_assignee", "mention_leader", "escalate"
}

// AnomalyDetectionBus evaluates anomaly detectors on each incoming event.
type AnomalyDetectionBus struct {
	detectors []AnomalyDetector
	onFinding func(Event, AnomalyFinding)
}

// NewAnomalyDetectionBus creates a new anomaly detection bus.
func NewAnomalyDetectionBus() *AnomalyDetectionBus {
	return &AnomalyDetectionBus{}
}

// AddDetector registers a new anomaly detector.
func (adb *AnomalyDetectionBus) AddDetector(d AnomalyDetector) {
	adb.detectors = append(adb.detectors, d)
}

// SetOnFinding sets the callback for when an anomaly is detected.
func (adb *AnomalyDetectionBus) SetOnFinding(fn func(Event, AnomalyFinding)) {
	adb.onFinding = fn
}

// Evaluate runs all detectors against an event.
func (adb *AnomalyDetectionBus) Evaluate(e Event) []AnomalyFinding {
	var findings []AnomalyFinding
	for _, d := range adb.detectors {
		for _, et := range d.EventTypes {
			if et == e.Type {
				findings = append(findings, d.Evaluate(e)...)
				break
			}
		}
	}
	return findings
}

// Start wires anomaly detectors into the event bus.
func (adb *AnomalyDetectionBus) Start(bus *Bus) {
	eventTypes := map[string]bool{}
	for _, d := range adb.detectors {
		for _, et := range d.EventTypes {
			eventTypes[et] = true
		}
	}

	for et := range eventTypes {
		et := et
		bus.Subscribe(et, func(e Event) {
			findings := adb.Evaluate(e)
			for _, f := range findings {
				if adb.onFinding != nil {
					adb.onFinding(e, f)
				}
			}
		})
	}
	slog.Info("anomaly detection bus: started", "detectors", len(adb.detectors), "event_types", len(eventTypes))
}

// ── D1: Parent-Child Deadlock Detector ─────────────────────────────────────
// Detects when all children are terminal but parent is not.

func MakeD1ParentChildDeadlockDetector(reader MetadataReader) AnomalyDetector {
	return AnomalyDetector{
		Name:       "D1_parent_child_deadlock",
		EventTypes: []string{protocol.EventIssueChildTerminal, protocol.EventIssueUpdated},
		Evaluate: func(e Event) []AnomalyFinding {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return nil
			}

			issue, _ := payload["issue"].(map[string]any)
			if issue == nil {
				return nil
			}

			parentID, _ := issue["parent_issue_id"].(string)
			if parentID == "" {
				return nil
			}

			// Check if this child becoming terminal means all children are terminal.
			children, err := reader.GetChildIssues(parentID)
			if err != nil || len(children) == 0 {
				return nil
			}

			allTerminal := true
			for _, c := range children {
				if c.Status != "done" && c.Status != "cancelled" {
					allTerminal = false
					break
				}
			}

			if !allTerminal {
				return nil
			}

			// Parent exists but is not terminal — deadlock detected.
			parentStatus, _, _, _, err := reader.GetIssue(parentID)
			if err != nil {
				return nil
			}

			if parentStatus == "done" || parentStatus == "cancelled" {
				return nil
			}

			return []AnomalyFinding{{
				Severity:    "L3",
				IssueID:     parentID,
				Title:       "Parent-child deadlock detected",
				Description: fmt.Sprintf("All %d children are terminal but parent status is '%s'.", len(children), parentStatus),
				Action:      "comment",
			}}
		},
	}
}

// ── D2: Contract Timeout Detector ──────────────────────────────────────────
// Detects when a dispatch contract's to_issue is terminal but contract
// remains pending past a reasonable window.

func MakeD2ContractTimeoutDetector(reader MetadataReader) AnomalyDetector {
	return AnomalyDetector{
		Name:       "D2_contract_timeout",
		EventTypes: []string{protocol.EventIssueChildTerminal},
		Evaluate: func(e Event) []AnomalyFinding {
			payload, ok := e.Payload.(map[string]any)
			if !ok {
				return nil
			}

			issue, _ := payload["issue"].(map[string]any)
			if issue == nil {
				return nil
			}

			metadata, _ := issue["metadata"].(map[string]any)
			contractID, _ := metadata["dispatch_contract_id"].(string)
			if contractID == "" {
				return nil
			}

			// The issue is terminal and has a dispatch contract — if the
			// contract is still pending at this point, it may have been missed.
			issueID, _ := issue["id"].(string)

			return []AnomalyFinding{{
				Severity:    "L3",
				IssueID:     issueID,
				Title:       "Dispatch contract may need fulfillment check",
				Description: fmt.Sprintf("Issue %s with dispatch contract %s reached terminal status.", issueID, contractID),
				Action:      "mention_assignee",
			}}
		},
	}
}

// ── D3: Stale in_review Detector ────────────────────────────────────────────
// Detects issues stuck in in_review status.

func MakeD3StaleReviewDetector(reader MetadataReader) AnomalyDetector {
	return AnomalyDetector{
		Name:       "D3_stale_review",
		EventTypes: []string{protocol.EventIssueUpdated, protocol.EventIssueMetadataChanged},
		Evaluate: func(e Event) []AnomalyFinding {
			// This detector is activated by any status change event.
			// In a full implementation it checks all in_review issues for staleness.
			// The "event train" pattern: uses unrelated events as hooks
			// to perform lightweight scans.
			return nil // full implementation scans cross-issue state
		},
	}
}

// ── Helper: Event Type Derivation ──────────────────────────────────────────

// DeriveSynthesizedEvents listens for base events and emits synthesized
// higher-level events (child_terminal, stage_completed, blocked, unblocked).
func DeriveSynthesizedEvents(bus *Bus) {
	bus.Subscribe(protocol.EventIssueUpdated, func(e Event) {
		payload, ok := e.Payload.(map[string]any)
		if !ok {
			return
		}

		statusChanged, _ := payload["status_changed"].(bool)
		if !statusChanged {
			return
		}

		issue, _ := payload["issue"].(map[string]any)
		prevStatus, _ := payload["prev_status"].(string)
		newStatus, _ := issue["status"].(string)
		issueID, _ := issue["id"].(string)
		workspaceID, _ := issue["workspace_id"].(string)

		// Derive child_terminal event.
		parentID, _ := issue["parent_issue_id"].(string)
		if parentID != "" && (newStatus == "done" || newStatus == "cancelled") {
			if prevStatus != "done" && prevStatus != "cancelled" {
				bus.Publish(Event{
					Type:        protocol.EventIssueChildTerminal,
					WorkspaceID: workspaceID,
					ActorType:   e.ActorType,
					ActorID:     e.ActorID,
					Payload: map[string]any{
						"issue_id":        issueID,
						"parent_issue_id": parentID,
						"status":          newStatus,
						"issue":           issue,
					},
				})
			}
		}

		// Derive blocked / unblocked events.
		if newStatus == "blocked" && prevStatus != "blocked" {
			bus.Publish(Event{
				Type:        protocol.EventIssueBlocked,
				WorkspaceID: workspaceID,
				ActorType:   e.ActorType,
				ActorID:     e.ActorID,
				Payload:     payload,
			})
		}
		if prevStatus == "blocked" && newStatus != "blocked" {
			bus.Publish(Event{
				Type:        protocol.EventIssueUnblocked,
				WorkspaceID: workspaceID,
				ActorType:   e.ActorType,
				ActorID:     e.ActorID,
				Payload:     payload,
			})
		}
	})

	slog.Info("notification bus: synthesized event derivation registered")
}

// ── Helper: Metadata Validation ─────────────────────────────────────────────

// ValidateInReviewMetadata checks that an issue entering in_review has
// the required interaction_posture metadata key set.
func ValidateInReviewMetadata(metadata map[string]any) (valid bool, warnings []string) {
	posture, hasPosture := metadata["interaction_posture"]
	if !hasPosture {
		warnings = append(warnings,
			"Issue entered in_review without interaction_posture metadata. "+
				"Set to 'submitted_for_review' or 'review_completed' to clarify intent.")
		return true, warnings
	}

	postureStr, ok := posture.(string)
	if !ok {
		warnings = append(warnings,
			"interaction_posture must be a string. Value will be cleared.")
		return false, warnings
	}

	switch strings.ToLower(postureStr) {
	case "submitted_for_review", "review_completed":
		return true, nil
	default:
		warnings = append(warnings,
			fmt.Sprintf("Invalid interaction_posture '%s'. Must be 'submitted_for_review' or 'review_completed'.", postureStr))
		return false, warnings
	}
}

// ── Event Log Persistence ──────────────────────────────────────────────────

// EventLogWriter writes key events to structured JSON Lines log for
// reliability and replay. This is the first layer of the 4-layer event
// reliability strategy.
type EventLogWriter struct {
	logger *slog.Logger
}

// NewEventLogWriter creates a new event log writer.
func NewEventLogWriter(logger *slog.Logger) *EventLogWriter {
	return &EventLogWriter{logger: logger}
}

// Write logs an event as structured JSON for reliability.
func (w *EventLogWriter) Write(e Event) {
	if w.logger == nil {
		return
	}
	payloadJSON, _ := json.Marshal(e.Payload)
	w.logger.Info("notification_event",
		"type", e.Type,
		"workspace_id", e.WorkspaceID,
		"actor_type", e.ActorType,
		"actor_id", e.ActorID,
		"payload", string(payloadJSON),
	)
}
