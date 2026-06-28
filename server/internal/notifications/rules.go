package notifications

import (
	"context"
	"log/slog"
	"strings"

	"github.com/multica-ai/multica/server/internal/events"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// ============================================================================
// R1: Child task completion notification
// ============================================================================
// When a child issue transitions to terminal (done/cancelled), synthesize
// child:terminal event. The actual notification to parent assignee is handled
// by the existing notifyParentOfChildDone in the handler layer.
// Priority: 50 (lowest — most generic rule).
// ============================================================================

// BuildR1 returns the R1 rule.
func BuildR1(queries *db.Queries) Rule {
	return Rule{
		ID:       "R1",
		Priority: PriorityR1,
		Match: func(tctx *TriggerContext) *MatchResult {
			issue := tctx.Issue
			if !issue.ParentIssueID.Valid {
				return nil
			}
			payload, ok := tctx.Event.Payload.(map[string]any)
			if !ok {
				return nil
			}
			prevStatus, _ := payload["prev_status"].(string)
			if isTerminalChildStatus(prevStatus) || !isTerminalChildStatus(issue.Status) {
				return nil
			}

			parentID := util.UUIDToString(issue.ParentIssueID)
			return &MatchResult{
				TargetType: "issue",
				TargetID:   parentID,
				IssueID:    util.UUIDToString(issue.ID),
				NotifType:  "child_terminal",
				Title:      "Sub-issue completed: " + issue.Title,
				Body:       "A sub-issue of your parent issue has been completed.",
				Details: map[string]any{
					"parent_id": parentID,
					"child_id":  util.UUIDToString(issue.ID),
					"status":    issue.Status,
				},
			}
		},
		Action: func(ctx context.Context, deps RuleDeps, result *MatchResult) error {
			parentID, _ := result.Details["parent_id"].(string)
			childID, _ := result.Details["child_id"].(string)
			if parentID == "" || childID == "" {
				return nil
			}
			deps.Bus.Publish(events.Event{
				Type:        "child:terminal",
				WorkspaceID: "", // set by caller
				ActorType:   "system_rule",
				Payload: map[string]any{
					"parent_id": parentID,
					"child_id":  childID,
					"status":    result.Details["status"],
				},
			})
			return nil
		},
	}
}

// ============================================================================
// R2: Review completion notification
// ============================================================================
// When interaction_posture is set to "review_completed", notify the issue
// specified in review_target metadata.
// Priority: 20.
// ============================================================================

// BuildR2 returns the R2 rule.
func BuildR2(queries *db.Queries) Rule {
	return Rule{
		ID:       "R2",
		Priority: PriorityR2,
		Match: func(tctx *TriggerContext) *MatchResult {
			posture, hasPosture := tctx.Metadata["interaction_posture"]
			if !hasPosture || posture != "review_completed" {
				return nil
			}
			reviewTarget, hasTarget := tctx.Metadata["review_target"]
			if !hasTarget || reviewTarget == "" {
				return nil
			}
			reviewScore, _ := tctx.Metadata["review_score"]
			reviewFindings, _ := tctx.Metadata["review_findings"]

			return &MatchResult{
				TargetType: "issue",
				TargetID:   reviewTarget,
				IssueID:    util.UUIDToString(tctx.Issue.ID),
				NotifType:  "review_completed",
				Title:      "Review completed for " + tctx.Issue.Title,
				Body:       buildReviewBody(reviewScore, reviewFindings),
				Details: map[string]any{
					"review_target":    reviewTarget,
					"review_score":     reviewScore,
					"review_findings":  reviewFindings,
					"reviewer_issue_id": util.UUIDToString(tctx.Issue.ID),
				},
			}
		},
		Action: func(ctx context.Context, deps RuleDeps, result *MatchResult) error {
			if deps.NotifyFn != nil {
				return deps.NotifyFn(ctx, result)
			}
			slog.Info("R2: review completed",
				"review_target", result.TargetID,
				"reviewer", result.IssueID)
			return nil
		},
	}
}

// ============================================================================
// R3: Blocking removal notification
// ============================================================================
// When waiting_on_issue metadata is cleared, notify subscribers of the
// previously-blocked issue.
// Priority: 40.
// ============================================================================

// BuildR3 returns the R3 rule.
func BuildR3(queries *db.Queries) Rule {
	return Rule{
		ID:       "R3",
		Priority: PriorityR3,
		Match: func(tctx *TriggerContext) *MatchResult {
			payload, ok := tctx.Event.Payload.(map[string]any)
			if !ok {
				return nil
			}
			key, _ := payload["key"].(string)
			if key != "waiting_on_issue" {
				return nil
			}
			prevValue, _ := payload["prev_value"].(string)
			newValue, _ := payload["value"].(string)
			// Cleared (deleted or set to empty).
			if prevValue == "" || newValue != "" {
				return nil
			}

			return &MatchResult{
				TargetType: "issue",
				TargetID:   prevValue,
				IssueID:    util.UUIDToString(tctx.Issue.ID),
				NotifType:  "blocking_removed",
				Title:      "Blocking issue resolved: " + tctx.Issue.Title,
				Body:       "The issue that was blocking your task has been resolved.",
				Details: map[string]any{
					"blocking_issue": prevValue,
					"issue_id":       util.UUIDToString(tctx.Issue.ID),
				},
			}
		},
		Action: func(ctx context.Context, deps RuleDeps, result *MatchResult) error {
			if deps.NotifyFn != nil {
				return deps.NotifyFn(ctx, result)
			}
			return nil
		},
	}
}

// ============================================================================
// R4: Stage completion notification
// ============================================================================
// When all children in the lowest unfinished stage of a parent are terminal,
// synthesize stage:closed event.
// Priority: 30.
// ============================================================================

// BuildR4 returns the R4 rule.
func BuildR4(queries *db.Queries) Rule {
	return Rule{
		ID:       "R4",
		Priority: PriorityR4,
		Match: func(tctx *TriggerContext) *MatchResult {
			issue := tctx.Issue
			if !issue.ParentIssueID.Valid {
				return nil
			}
			payload, ok := tctx.Event.Payload.(map[string]any)
			if !ok {
				return nil
			}
			prevStatus, _ := payload["prev_status"].(string)
			if isTerminalChildStatus(prevStatus) || !isTerminalChildStatus(issue.Status) {
				return nil
			}

			parentID := util.UUIDToString(issue.ParentIssueID)
			return &MatchResult{
				TargetType: "issue",
				TargetID:   parentID,
				IssueID:    util.UUIDToString(issue.ID),
				NotifType:  "stage_closed",
				Title:      "Stage completed: " + issue.Title,
				Body:       "All sub-issues in the current stage are now complete.",
				Details: map[string]any{
					"parent_id": parentID,
					"child_id":  util.UUIDToString(issue.ID),
					"status":    issue.Status,
				},
			}
		},
		Action: func(ctx context.Context, deps RuleDeps, result *MatchResult) error {
			parentID, _ := result.Details["parent_id"].(string)
			childID, _ := result.Details["child_id"].(string)
			if parentID == "" || childID == "" {
				return nil
			}
			deps.Bus.Publish(events.Event{
				Type:        "stage:closed",
				WorkspaceID: "", // set by caller
				ActorType:   "system_rule",
				Payload: map[string]any{
					"parent_id": parentID,
					"child_id":  childID,
				},
			})
			return nil
		},
	}
}

// ============================================================================
// R5: Dispatch contract callback
// ============================================================================
// When a dispatch contract's trigger events occur on the source issue,
// mark the contract as triggered and notify the contract's target_issue.
// Priority: 10 (highest).
// ============================================================================

// BuildR5 returns the R5 rule.
func BuildR5(queries *db.Queries) Rule {
	return Rule{
		ID:       "R5",
		Priority: PriorityR5,
		Match: func(tctx *TriggerContext) *MatchResult {
			eventType := tctx.Event.Type
			issueID := util.UUIDToString(tctx.Issue.ID)

			contracts, err := queries.ListPendingDispatchContractsByIssue(
				context.Background(),
				db.ListPendingDispatchContractsByIssueParams{
					IssueID:    util.MustParseUUID(issueID),
					EventTypes: []string{eventType},
				},
			)
			if err != nil || len(contracts) == 0 {
				return nil
			}

			contract := contracts[0]
			var targetID string
			if contract.TargetIssue.Valid {
				targetID = util.UUIDToString(contract.TargetIssue)
			} else {
				targetID = util.UUIDToString(contract.ToIssueID)
			}

			return &MatchResult{
				TargetType: "issue",
				TargetID:   targetID,
				IssueID:    util.UUIDToString(tctx.Issue.ID),
				NotifType:  "contract_triggered",
				Title:      "Dispatch contract triggered: " + tctx.Issue.Title,
				Body:       "A contract-linked event has occurred on the source issue.",
				Details: map[string]any{
					"contract_id":   util.UUIDToString(contract.ID),
					"from_issue_id": util.UUIDToString(contract.FromIssueID),
					"to_issue_id":   util.UUIDToString(contract.ToIssueID),
				},
			}
		},
		Action: func(ctx context.Context, deps RuleDeps, result *MatchResult) error {
			contractID, _ := result.Details["contract_id"].(string)
			if contractID != "" {
				if err := queries.FulfillDispatchContract(ctx,
					util.MustParseUUID(contractID)); err != nil {
					slog.Error("R5: failed to fulfill contract",
						"contract_id", contractID, "error", err)
				}
			}
			deps.Bus.Publish(events.Event{
				Type:      "contract:fulfilled",
				ActorType: "system_rule",
				Payload:   result.Details,
			})
			if deps.NotifyFn != nil {
				return deps.NotifyFn(ctx, result)
			}
			return nil
		},
	}
}

// ============================================================================
// Helpers
// ============================================================================

func isTerminalChildStatus(status string) bool {
	return status == "done" || status == "cancelled"
}

// buildReviewBody constructs a human-readable review completion message.
func buildReviewBody(score, findings string) string {
	var parts []string
	if score != "" {
		parts = append(parts, "Score: "+score)
	}
	if findings != "" {
		if len(findings) > 200 {
			findings = findings[:200] + "..."
		}
		parts = append(parts, "Findings: "+findings)
	}
	if len(parts) == 0 {
		return "Review has been completed."
	}
	return strings.Join(parts, "\n")
}
