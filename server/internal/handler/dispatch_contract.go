package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/logger"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
	"github.com/multica-ai/multica/server/pkg/protocol"
)

// ── Dispatch Contract HTTP Handlers ──────────────────────────────────────
//
// A dispatch contract formalizes a cross-squad task delegation. When Squad A
// needs work from Squad B, it creates a dispatch contract that binds the
// two issues together and prescribes callback behavior when Squad B completes.
//
// These handlers use raw SQL via the DB executor because the sqlc-generated
// types for dispatch_contracts and notification_records are not yet available
// until `make sqlc` is run against the 128_notification_bus migration.

// CreateDispatchContractRequest is the JSON body for creating a dispatch.
type CreateDispatchContractRequest struct {
	Title          string   `json:"title"`
	Description    *string  `json:"description"`
	Priority       *string  `json:"priority"`
	ToSquadID      string   `json:"to_squad_id"`
	FromIssueID    string   `json:"from_issue_id"`
	TriggerEvents  []string `json:"trigger_events"`
	TargetIssueID  *string  `json:"target_issue_id"`
	NotifyAssignee *bool    `json:"notify_assignee"`
	NotifyCreator  *bool    `json:"notify_creator"`
	IncludeSummary *bool    `json:"include_summary"`
	HandoffNote    *string  `json:"handoff_note"`
}

// CreateDispatchContract creates a child issue AND a dispatch contract
// linking the parent and child, with prescribed callback behavior.
//
// POST /api/dispatch
func (h *Handler) CreateDispatchContract(w http.ResponseWriter, r *http.Request) {
	var req CreateDispatchContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.FromIssueID == "" {
		writeError(w, http.StatusBadRequest, "from_issue_id is required")
		return
	}
	if req.ToSquadID == "" {
		writeError(w, http.StatusBadRequest, "to_squad_id is required")
		return
	}

	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	// Load the parent issue.
	parentIssue, ok := h.loadIssueForUser(w, r, req.FromIssueID)
	if !ok {
		return
	}

	// Set defaults.
	triggerEvents := req.TriggerEvents
	if len(triggerEvents) == 0 {
		triggerEvents = []string{"child_terminal"}
	}

	targetIssueID := req.FromIssueID
	if req.TargetIssueID != nil && *req.TargetIssueID != "" {
		targetIssueID = *req.TargetIssueID
	}

	notifyAssignee := true
	if req.NotifyAssignee != nil {
		notifyAssignee = *req.NotifyAssignee
	}
	notifyCreator := false
	if req.NotifyCreator != nil {
		notifyCreator = *req.NotifyCreator
	}
	includeSummary := true
	if req.IncludeSummary != nil {
		includeSummary = *req.IncludeSummary
	}

	// Build child issue description.
	childDescription := ""
	if req.HandoffNote != nil {
		childDescription = *req.HandoffNote
	}

	// Set priority default.
	priority := "none"
	if req.Priority != nil {
		priority = *req.Priority
	}

	// Assignee info: the child issue is assigned to the target squad.
	assigneeType := "squad"
	assigneeID := req.ToSquadID

	// Determine creator type: agent if X-Agent-ID header present, else member.
	creatorType := "member"
	creatorID := userID
	agentID := r.Header.Get("X-Agent-ID")
	if agentID != "" {
		creatorType = "agent"
		creatorID = agentID
	}

	descText := util.StrToText(childDescription)
	assigneeTypeText := util.StrToText(assigneeType)
	assigneeIDUUID := parseUUID(assigneeID)
	parentID := parseUUID(req.FromIssueID)

	// Create child issue via Queries.
	childIssue, err := h.Queries.CreateIssue(r.Context(), db.CreateIssueParams{
		WorkspaceID:   parentIssue.WorkspaceID,
		Title:         req.Title,
		Description:   descText,
		Status:        "todo",
		Priority:      priority,
		CreatorType:   creatorType,
		CreatorID:     parseUUID(creatorID),
		AssigneeType:  assigneeTypeText,
		AssigneeID:    assigneeIDUUID,
		ParentIssueID: parentID,
	})
	if err != nil {
		slog.Warn("dispatch: failed to create child issue", append(logger.RequestAttrs(r), "error", err)...)
		writeError(w, http.StatusInternalServerError, "failed to create child issue")
		return
	}

	childIDStr := uuidToString(childIssue.ID)

	// Marshal trigger events to JSON.
	triggerJSON, _ := json.Marshal(triggerEvents)

	// Insert dispatch contract using raw SQL (sqlc types not yet generated for
	// the dispatch_contracts table — see 128_notification_bus migration).
	var fromSquadID, toSquadID, targetIssue pgtype.UUID
	if parentIssue.AssigneeType.Valid && parentIssue.AssigneeType.String == "squad" {
		fromSquadID = parentIssue.AssigneeID
	}
	toSquadID = parseUUID(req.ToSquadID)
	targetIssue = parseUUID(targetIssueID)

	var contractID pgtype.UUID
	err = h.DB.QueryRow(r.Context(), `
		INSERT INTO dispatch_contracts (
			creator_agent_id, from_issue_id, to_issue_id,
			from_squad_id, to_squad_id, trigger_events,
			target_issue, notify_assignee, notify_creator, include_summary
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING contract_id
	`,
		parseUUID(creatorID),
		parentIssue.ID,
		childIssue.ID,
		fromSquadID,
		toSquadID,
		triggerJSON,
		targetIssue,
		notifyAssignee,
		notifyCreator,
		includeSummary,
	).Scan(&contractID)
	if err != nil {
		slog.Warn("dispatch: failed to create contract", append(logger.RequestAttrs(r), "error", err)...)
		writeError(w, http.StatusInternalServerError, "failed to create dispatch contract")
		return
	}

	workspaceID := uuidToString(parentIssue.WorkspaceID)
	actorType, actorID := h.resolveActor(r, userID, workspaceID)

	// Publish dispatch contract created event.
	h.publish("dispatch_contract:created", workspaceID, actorType, actorID, map[string]any{
		"contract_id":   uuidToString(contractID),
		"from_issue_id": req.FromIssueID,
		"to_issue_id":   childIDStr,
		"from_squad_id": uuidToString(parentIssue.AssigneeID),
		"to_squad_id":   req.ToSquadID,
		"target_issue":  targetIssueID,
	})

	// Publish issue:created for the child.
	prefix := h.getIssuePrefix(r.Context(), parentIssue.WorkspaceID)
	childResp := issueToResponse(childIssue, prefix)
	h.publish(protocol.EventIssueCreated, workspaceID, actorType, actorID, map[string]any{
		"issue": childResp,
	})

	writeJSON(w, http.StatusCreated, map[string]any{
		"contract_id": uuidToString(contractID),
		"issue":       childResp,
		"status":      "pending",
	})
}

// ListDispatchContracts lists dispatch contracts, optionally filtered by status.
//
// GET /api/dispatch?status=pending
func (h *Handler) ListDispatchContracts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	_ = userID

	// Query contracts using raw SQL.
	rows, err := h.DB.Query(r.Context(), `
		SELECT contract_id, created_at, creator_agent_id, from_issue_id,
		       to_issue_id, from_squad_id, to_squad_id, trigger_events,
		       target_issue, notify_assignee, notify_creator, include_summary,
		       status, fulfilled_at
		FROM dispatch_contracts
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, status)
	if err != nil {
		slog.Warn("dispatch: list contracts failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list dispatch contracts")
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var (
			contractID, creatorAgentID, fromIssueID           pgtype.UUID
			toIssueID, fromSquadID, toSquadID, targetIssueStr pgtype.UUID
			createdAt                                         pgtype.Timestamptz
			fulfilledAt                                       pgtype.Timestamptz
			contractStatus                                    string
			notifyAssignee, notifyCreator, includeSummary     bool
		)
		// The trigger_events column is JSONB; we scan it as []byte.
		var triggerJSON []byte
		if err := rows.Scan(
			&contractID, &createdAt, &creatorAgentID, &fromIssueID,
			&toIssueID, &fromSquadID, &toSquadID, &triggerJSON,
			&targetIssueStr, &notifyAssignee, &notifyCreator, &includeSummary,
			&contractStatus, &fulfilledAt,
		); err != nil {
			continue
		}

		results = append(results, map[string]any{
			"contract_id":      uuidToString(contractID),
			"created_at":       createdAt.Time.Format("2006-01-02T15:04:05Z"),
			"creator_agent_id": uuidToString(creatorAgentID),
			"from_issue_id":    uuidToString(fromIssueID),
			"to_issue_id":      uuidToPtr(toIssueID),
			"from_squad_id":    uuidToPtr(fromSquadID),
			"to_squad_id":      uuidToPtr(toSquadID),
			"target_issue":     uuidToPtr(targetIssueStr),
			"trigger_events":   string(triggerJSON),
			"notify_assignee":  notifyAssignee,
			"notify_creator":   notifyCreator,
			"include_summary":  includeSummary,
			"status":           contractStatus,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"contracts": results})
}

// GetDispatchContract returns a single dispatch contract by ID.
//
// GET /api/dispatch/{contractId}
func (h *Handler) GetDispatchContract(w http.ResponseWriter, r *http.Request) {
	contractUUID := chi.URLParam(r, "contractId")

	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	_ = userID

	var (
		contractID, creatorAgentID, fromIssueID           pgtype.UUID
		toIssueID, fromSquadID, toSquadID, targetIssueStr pgtype.UUID
		createdAt                                         pgtype.Timestamptz
		fulfilledAt                                       pgtype.Timestamptz
		contractStatus                                    string
		notifyAssignee, notifyCreator, includeSummary     bool
		triggerJSON                                       []byte
	)

	err := h.DB.QueryRow(r.Context(), `
		SELECT contract_id, created_at, creator_agent_id, from_issue_id,
		       to_issue_id, from_squad_id, to_squad_id, trigger_events,
		       target_issue, notify_assignee, notify_creator, include_summary,
		       status, fulfilled_at
		FROM dispatch_contracts
		WHERE contract_id = $1
	`, parseUUID(contractUUID)).Scan(
		&contractID, &createdAt, &creatorAgentID, &fromIssueID,
		&toIssueID, &fromSquadID, &toSquadID, &triggerJSON,
		&targetIssueStr, &notifyAssignee, &notifyCreator, &includeSummary,
		&contractStatus, &fulfilledAt,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "dispatch contract not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"contract_id":      uuidToString(contractID),
		"created_at":       createdAt.Time.Format("2006-01-02T15:04:05Z"),
		"creator_agent_id": uuidToString(creatorAgentID),
		"from_issue_id":    uuidToString(fromIssueID),
		"to_issue_id":      uuidToPtr(toIssueID),
		"from_squad_id":    uuidToPtr(fromSquadID),
		"to_squad_id":      uuidToPtr(toSquadID),
		"target_issue":     uuidToPtr(targetIssueStr),
		"trigger_events":   string(triggerJSON),
		"notify_assignee":  notifyAssignee,
		"notify_creator":   notifyCreator,
		"include_summary":  includeSummary,
		"status":           contractStatus,
	})
}

// CancelDispatchContract cancels a pending dispatch contract.
//
// POST /api/dispatch/{contractId}/cancel
func (h *Handler) CancelDispatchContract(w http.ResponseWriter, r *http.Request) {
	contractUUID := chi.URLParam(r, "contractId")

	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	_ = userID

	_, err := h.DB.Exec(r.Context(), `
		UPDATE dispatch_contracts SET status = 'cancelled'
		WHERE contract_id = $1 AND status = 'pending'
	`, parseUUID(contractUUID))
	if err != nil {
		writeError(w, http.StatusNotFound, "dispatch contract not found or already fulfilled")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"contract_id": contractUUID,
		"status":      "cancelled",
	})
}

// ── Notification Query Handlers ──────────────────────────────────────────

// ListPendingNotifications lists pending notifications.
//
// GET /api/notifications/pending?target_type=agent&target_id=<id>
func (h *Handler) ListPendingNotifications(w http.ResponseWriter, r *http.Request) {
	targetType := r.URL.Query().Get("target_type")
	targetID := r.URL.Query().Get("target_id")
	if targetType == "" || targetID == "" {
		writeError(w, http.StatusBadRequest, "target_type and target_id are required")
		return
	}

	rows, err := h.DB.Query(r.Context(), `
		SELECT id, created_at, target_type, target_id, source_issue_id,
		       type, message, status, acknowledged_at
		FROM notification_records
		WHERE target_type = $1 AND target_id = $2 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 50
	`, targetType, parseUUID(targetID))
	if err != nil {
		slog.Warn("notifications: list pending failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list notifications")
		return
	}
	defer rows.Close()

	results := make([]map[string]any, 0)
	for rows.Next() {
		var (
			id, targetID, sourceIssueID                      pgtype.UUID
			createdAt, acknowledgedAt                        pgtype.Timestamptz
			notifType, notifTargetType, message, notifStatus string
		)
		if err := rows.Scan(&id, &createdAt, &notifTargetType, &targetID, &sourceIssueID,
			&notifType, &message, &notifStatus, &acknowledgedAt); err != nil {
			continue
		}
		results = append(results, map[string]any{
			"id":              uuidToString(id),
			"created_at":      createdAt.Time.Format("2006-01-02T15:04:05Z"),
			"target_type":     notifTargetType,
			"target_id":       uuidToString(targetID),
			"source_issue_id": uuidToString(sourceIssueID),
			"type":            notifType,
			"message":         message,
			"status":          notifStatus,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"notifications": results})
}

// AcknowledgeNotification marks a notification as acknowledged.
//
// POST /api/notifications/{id}/acknowledge
func (h *Handler) AcknowledgeNotification(w http.ResponseWriter, r *http.Request) {
	notifUUID := chi.URLParam(r, "id")

	_, err := h.DB.Exec(r.Context(), `
		UPDATE notification_records
		SET status = 'acknowledged', acknowledged_at = now()
		WHERE id = $1
	`, parseUUID(notifUUID))
	if err != nil {
		writeError(w, http.StatusNotFound, "notification not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":     notifUUID,
		"status": "acknowledged",
	})
}
