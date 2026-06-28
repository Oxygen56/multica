package notifications

import (
	"fmt"
	"strings"
)

// Dual-layer status semantics (Feature #4596 Module C).
//
// The existing `status` field remains unchanged. A second dimension is added
// via metadata keys that extend status semantics:
//
//   interaction_posture — "submitted_for_review" or "review_completed"
//   review_target      — UUID of the issue being reviewed
//   review_score       — optional score (string)
//   review_findings    — optional findings summary (string)
//   waiting_on_issue   — UUID of the blocking issue
//   dispatch_contract_id — UUID of the related dispatch contract
//   callback_target     — UUID of the issue to notify on completion
//   children_progress   — summary of child issue progress
//
// These keys combine with `status` to produce richer semantics:
//   - in_review + submitted_for_review = waiting for reviewer
//   - in_review + review_completed   = review done, waiting for author
//   - blocked + waiting_on_issue=X   = blocked on X specifically

// Known metadata keys for the notification system.
var KnownNotificationKeys = map[string]KeyConstraint{
	"interaction_posture": {
		Description: "Current interaction posture for the issue",
		AllowedValues: []string{
			"submitted_for_review",
			"review_completed",
		},
	},
	"review_target": {
		Description:   "UUID of the issue being reviewed (set by reviewer)",
		AllowedValues: nil, // any non-empty string
	},
	"review_score": {
		Description:   "Review score (e.g. pass, changes_requested)",
		AllowedValues: nil,
	},
	"review_findings": {
		Description:   "Summary of review findings",
		AllowedValues: nil,
	},
	"waiting_on_issue": {
		Description:   "UUID of the issue this issue is blocked on",
		AllowedValues: nil,
	},
	"dispatch_contract_id": {
		Description:   "UUID of the related dispatch contract",
		AllowedValues: nil,
	},
	"callback_target": {
		Description:   "UUID of the issue to notify when this issue completes",
		AllowedValues: nil,
	},
	"children_progress": {
		Description:   "Summary of child issues progress (e.g. '2/5 done')",
		AllowedValues: nil,
	},
}

// KeyConstraint defines validation rules for a metadata key.
type KeyConstraint struct {
	Description   string
	AllowedValues []string // nil means any non-empty string
}

// ValidateNotificationKey checks whether a key and value are valid for the
// notification system's dual-layer semantics. Returns nil if valid, or an
// error describing the constraint violation.
func ValidateNotificationKey(key, value string) error {
	c, ok := KnownNotificationKeys[key]
	if !ok {
		return nil // unknown keys are allowed (extensibility)
	}
	if value == "" {
		return fmt.Errorf("notification key %q requires a non-empty value", key)
	}
	if c.AllowedValues != nil {
		for _, allowed := range c.AllowedValues {
			if value == allowed {
				return nil
			}
		}
		return fmt.Errorf("notification key %q: value %q is not allowed (must be one of: %s)",
			key, value, strings.Join(c.AllowedValues, ", "))
	}
	return nil
}

// IsNotificationKey reports whether a key is one of the known notification
// system metadata keys.
func IsNotificationKey(key string) bool {
	_, ok := KnownNotificationKeys[key]
	return ok
}
