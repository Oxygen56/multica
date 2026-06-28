package notifications

import (
	"context"
	"log/slog"
	"time"

	"github.com/multica-ai/multica/server/internal/events"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// RepairLevel defines the severity of an anomaly repair action.
type RepairLevel int

const (
	L1SilentMetadata RepairLevel = iota + 1 // silent metadata update
	L2SystemComment                          // post a system comment as FYI
	L3NotifyDirect                           // notify assignee directly
	L4EscalateWatcher                        // notify watchers/admins
	L5CreateEscalation                       // create escalation issue
)

// Anomaly represents a detected anomaly in the collaboration graph.
type Anomaly struct {
	DetectorID  string      // "D1"–"D6"
	Severity    string      // "warning", "critical"
	Description string      // human-readable
	RepairLevel RepairLevel // recommended action level
	IssueID     string      // affected issue
	Details     map[string]any
}

// AnomalyDetector evaluates collaboration graph state on every relevant event
// (NOT on timers/cron — per the pure event-driven design in #4596).
type AnomalyDetector struct {
	bus     *events.Bus
	queries *db.Queries
}

// NewAnomalyDetector creates a new anomaly detector.
func NewAnomalyDetector(bus *events.Bus, queries *db.Queries) *AnomalyDetector {
	return &AnomalyDetector{bus: bus, queries: queries}
}

// ============================================================================
// D1: Parent-child deadlock detection
// ============================================================================
// When a child becomes terminal, check if the parent is blocked AND all
// other children are already terminal — if so, the parent won't progress
// without intervention.

func (d *AnomalyDetector) OnChildTerminal(ctx context.Context, ev events.Event) {
	payload, ok := ev.Payload.(map[string]any)
	if !ok {
		return
	}
	parentID, _ := payload["parent_id"].(string)
	if parentID == "" {
		return
	}

	parent, err := d.queries.GetIssue(ctx, util.MustParseUUID(parentID))
	if err != nil {
		return
	}
	if parent.Status != "blocked" {
		return
	}

	children, err := d.queries.ListChildIssues(ctx, util.MustParseUUID(parentID))
	if err != nil {
		return
	}

	allTerminal := true
	for _, c := range children {
		if !isTerminalChildStatus(c.Status) {
			allTerminal = false
			break
		}
	}

	if allTerminal {
		d.emit(ctx, Anomaly{
			DetectorID:  "D1",
			Severity:    "warning",
			Description: "Parent issue is blocked but all children are terminal — possible deadlock",
			RepairLevel: L3NotifyDirect,
			IssueID:     parentID,
			Details: map[string]any{
				"parent_id":  parentID,
				"child_count": len(children),
			},
		}, ev.WorkspaceID)
	}
}

// ============================================================================
// D2: Contract timeout detection
// ============================================================================
// When a contract has been pending for too long (no hard timer — checked
// "顺便" when any related event fires).

func (d *AnomalyDetector) checkContractTimeout(ctx context.Context, wsID string) {
	staleContracts, err := d.queries.ListStaleDispatchContracts(ctx, wsID)
	if err != nil || len(staleContracts) == 0 {
		return
	}
	for _, contract := range staleContracts {
		d.emit(ctx, Anomaly{
			DetectorID:  "D2",
			Severity:    "warning",
			Description: "Dispatch contract has been pending for over 24 hours without trigger",
			RepairLevel: L2SystemComment,
			IssueID:     util.UUIDToString(contract.FromIssueID),
			Details: map[string]any{
				"contract_id": util.UUIDToString(contract.ID),
				"created_at":  contract.CreatedAt.Time.Format(time.RFC3339),
			},
		}, wsID)
	}
}

// ============================================================================
// D3: Orphaned review detection
// ============================================================================
// When interaction_posture is "submitted_for_review" and the review_target
// issue has been done/cancelled — the review will never complete.

func (d *AnomalyDetector) checkOrphanedReview(ctx context.Context, tctx *TriggerContext) {
	posture, ok := tctx.Metadata["interaction_posture"]
	if !ok || posture != "submitted_for_review" {
		return
	}
	reviewTarget, ok := tctx.Metadata["review_target"]
	if !ok || reviewTarget == "" {
		return
	}

	target, err := d.queries.GetIssue(ctx, util.MustParseUUID(reviewTarget))
	if err != nil {
		return
	}
	if target.Status == "done" || target.Status == "cancelled" {
		d.emit(ctx, Anomaly{
			DetectorID:  "D3",
			Severity:    "warning",
			Description: "Review target issue is already closed — review will never be picked up",
			RepairLevel: L2SystemComment,
			IssueID:     util.UUIDToString(tctx.Issue.ID),
			Details: map[string]any{
				"review_target": reviewTarget,
				"target_status": target.Status,
			},
		}, util.UUIDToString(tctx.Issue.WorkspaceID))
	}
}

// ============================================================================
// D4: Stage stall detection
// ============================================================================
// When a stage is closed but the parent has no assignee, the parent stalls.

func (d *AnomalyDetector) OnStageClosed(ctx context.Context, ev events.Event) {
	payload, ok := ev.Payload.(map[string]any)
	if !ok {
		return
	}
	parentID, _ := payload["parent_id"].(string)
	if parentID == "" {
		return
	}

	parent, err := d.queries.GetIssue(ctx, util.MustParseUUID(parentID))
	if err != nil {
		return
	}
	if parent.AssigneeType.Valid {
		return // has assignee — can progress
	}

	d.emit(ctx, Anomaly{
		DetectorID:  "D4",
		Severity:    "critical",
		Description: "Stage completed but parent issue has no assignee — work is stalled",
		RepairLevel: L4EscalateWatcher,
		IssueID:     parentID,
		Details: map[string]any{
			"parent_id": parentID,
		},
	}, ev.WorkspaceID)
}

// ============================================================================
// D5: Cross-squad silence detection
// ============================================================================
// When a dispatch contract's to_issue hasn't had activity for >48h while
// the from_issue has been active — possible cross-squad communication gap.
// Checked "顺便" when any notification:delivered event fires.

func (d *AnomalyDetector) checkCrossSquadSilence(ctx context.Context, wsID string) {
	silent, err := d.queries.ListCrossSquadSilentContracts(ctx, wsID)
	if err != nil || len(silent) == 0 {
		return
	}
	for _, contract := range silent {
		d.emit(ctx, Anomaly{
			DetectorID:  "D5",
			Severity:    "warning",
			Description: "Cross-squad contract target has been silent — possible communication gap",
			RepairLevel: L3NotifyDirect,
			IssueID:     util.UUIDToString(contract.ToIssueID),
			Details: map[string]any{
				"contract_id":   util.UUIDToString(contract.ID),
				"from_issue_id": util.UUIDToString(contract.FromIssueID),
			},
		}, wsID)
	}
}

// ============================================================================
// D6: Metadata inconsistency detection
// ============================================================================
// When interaction_posture metadata contradicts issue status, flag it.
// e.g. status=done but posture=submitted_for_review.

func (d *AnomalyDetector) checkMetadataInconsistency(ctx context.Context, tctx *TriggerContext) {
	posture, ok := tctx.Metadata["interaction_posture"]
	if !ok {
		return
	}
	issue := tctx.Issue

	// submitted_for_review is only valid when status is in_review.
	invalid := false
	switch posture {
	case "submitted_for_review":
		if issue.Status != "in_review" {
			invalid = true
		}
	case "review_completed":
		if issue.Status != "in_review" && issue.Status != "done" {
			invalid = true
		}
	}

	if invalid {
		d.emit(ctx, Anomaly{
			DetectorID:  "D6",
			Severity:    "warning",
			Description: "Metadata posture contradicts issue status — possible stale state",
			RepairLevel: L1SilentMetadata,
			IssueID:     util.UUIDToString(issue.ID),
			Details: map[string]any{
				"posture": posture,
				"status":  issue.Status,
			},
		}, util.UUIDToString(issue.WorkspaceID))
	}
}

// ============================================================================
// Evaluate runs all anomaly detectors against the current trigger context.
// Called by the Engine after rule evaluation (the "顺便" pattern — detectors
// run as a side effect of event processing, not on timers).
// ============================================================================

func (d *AnomalyDetector) Evaluate(ctx context.Context, tctx *TriggerContext, matches []*MatchResult) {
	wsID := util.UUIDToString(tctx.Issue.WorkspaceID)

	// D3, D6 are per-event checks — they look at the current issue state.
	d.checkOrphanedReview(ctx, tctx)
	d.checkMetadataInconsistency(ctx, tctx)

	// D2, D5 are workspace-wide scans — they check stale contracts across
	// the workspace. These are gated: only run when a notification:delivered
	// or contract:triggered event fires, to bound scan frequency.
	if tctx.Event.Type == "notification:delivered" || tctx.Event.Type == "contract:triggered" {
		d.checkContractTimeout(ctx, wsID)
		d.checkCrossSquadSilence(ctx, wsID)
	}
}

// emit publishes an anomaly event on the bus and logs it.
func (d *AnomalyDetector) emit(ctx context.Context, a Anomaly, workspaceID string) {
	slog.Warn("anomaly detected",
		"detector", a.DetectorID,
		"severity", a.Severity,
		"description", a.Description,
		"issue_id", a.IssueID,
	)

	d.bus.Publish(events.Event{
		Type:        "anomaly:detected",
		WorkspaceID: workspaceID,
		ActorType:   "system_rule",
		Payload: map[string]any{
			"anomaly": a,
		},
	})
}
