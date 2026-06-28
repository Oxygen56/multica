// Package dispatch implements the dispatch contract system (Feature #4596 Module B).
// A dispatch contract is a formal link between two issues: the source
// (from_issue) agrees to notify the target (to_issue) when certain trigger
// events occur. This is distinct from plain parent-child relationships:
// it's a temporary, event-driven coordination agreement.
//
// CLI: `multica dispatch create|list|show|cancel`
package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/multica-ai/multica/server/internal/events"
	"github.com/multica-ai/multica/server/internal/util"
	db "github.com/multica-ai/multica/server/pkg/db/generated"
)

// ContractStatus represents the lifecycle state of a dispatch contract.
type ContractStatus string

const (
	StatusPending   ContractStatus = "pending"
	StatusTriggered ContractStatus = "triggered"
	StatusFulfilled ContractStatus = "fulfilled"
	StatusCancelled ContractStatus = "cancelled"
)

// Contract is the application-level representation of a dispatch contract.
type Contract struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id"`
	FromIssueID   string         `json:"from_issue_id"`
	ToIssueID     string         `json:"to_issue_id"`
	TriggerEvents []string       `json:"trigger_events"`
	TargetIssue   *string        `json:"target_issue,omitempty"`
	Status        ContractStatus `json:"status"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	FulfilledAt   *time.Time     `json:"fulfilled_at,omitempty"`
	CancelledAt   *time.Time     `json:"cancelled_at,omitempty"`
}

// Service manages dispatch contract lifecycle operations.
type Service struct {
	Queries *db.Queries
	Bus     *events.Bus
}

// NewService creates a new dispatch contract service.
func NewService(queries *db.Queries, bus *events.Bus) *Service {
	return &Service{Queries: queries, Bus: bus}
}

// CreateParams holds the input for creating a new dispatch contract.
type CreateParams struct {
	WorkspaceID   string
	FromIssueID   string
	ToIssueID     string
	TriggerEvents []string // e.g. ["status_changed", "comment:created"]
	TargetIssue   string   // optional
	Metadata      map[string]any
}

// Create creates a new dispatch contract between two issues.
func (s *Service) Create(ctx context.Context, params CreateParams) (*Contract, error) {
	if params.FromIssueID == params.ToIssueID {
		return nil, fmt.Errorf("dispatch: cannot create contract from issue to itself")
	}
	if len(params.TriggerEvents) == 0 {
		return nil, fmt.Errorf("dispatch: at least one trigger event is required")
	}

	var targetIssue pgtype.UUID
	if params.TargetIssue != "" {
		targetIssue = util.MustParseUUID(params.TargetIssue)
	}

	metaJSON := []byte("{}")
	if params.Metadata != nil {
		var err error
		metaJSON, err = json.Marshal(params.Metadata)
		if err != nil {
			return nil, fmt.Errorf("dispatch: invalid metadata: %w", err)
		}
	}

	row, err := s.Queries.CreateDispatchContract(ctx, db.CreateDispatchContractParams{
		WorkspaceID:   util.MustParseUUID(params.WorkspaceID),
		FromIssueID:   util.MustParseUUID(params.FromIssueID),
		ToIssueID:     util.MustParseUUID(params.ToIssueID),
		TriggerEvents: params.TriggerEvents,
		TargetIssue:   targetIssue,
		Metadata:      metaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("dispatch: create contract: %w", err)
	}

	contract := toContract(row)

	s.Bus.Publish(events.Event{
		Type:        "contract:created",
		WorkspaceID: params.WorkspaceID,
		ActorType:   "system",
		Payload: map[string]any{
			"contract": contract,
		},
	})

	return contract, nil
}

// List returns all contracts for a workspace, filtered by status.
func (s *Service) List(ctx context.Context, workspaceID string, status ContractStatus) ([]Contract, error) {
	var rows []db.DispatchContract
	var err error

	if status == "" {
		rows, err = s.Queries.ListDispatchContractsByWorkspace(ctx,
			util.MustParseUUID(workspaceID))
	} else {
		rows, err = s.Queries.ListDispatchContractsByStatus(ctx,
			db.ListDispatchContractsByStatusParams{
				WorkspaceID: util.MustParseUUID(workspaceID),
				Status:      string(status),
			})
	}
	if err != nil {
		return nil, fmt.Errorf("dispatch: list contracts: %w", err)
	}

	contracts := make([]Contract, len(rows))
	for i, row := range rows {
		contracts[i] = *toContract(row)
	}
	return contracts, nil
}

// Get returns a single contract by ID.
func (s *Service) Get(ctx context.Context, contractID string) (*Contract, error) {
	row, err := s.Queries.GetDispatchContract(ctx, util.MustParseUUID(contractID))
	if err != nil {
		return nil, fmt.Errorf("dispatch: get contract: %w", err)
	}
	return toContract(row), nil
}

// Cancel cancels a pending dispatch contract.
func (s *Service) Cancel(ctx context.Context, contractID string) (*Contract, error) {
	row, err := s.Queries.CancelDispatchContract(ctx, util.MustParseUUID(contractID))
	if err != nil {
		return nil, fmt.Errorf("dispatch: cancel contract: %w", err)
	}
	contract := toContract(row)

	s.Bus.Publish(events.Event{
		Type:        "contract:cancelled",
		WorkspaceID: util.UUIDToString(row.WorkspaceID),
		ActorType:   "system",
		Payload: map[string]any{
			"contract": contract,
		},
	})

	return contract, nil
}

// ListByIssue returns all contracts associated with an issue.
func (s *Service) ListByIssue(ctx context.Context, issueID string) ([]Contract, error) {
	rows, err := s.Queries.ListDispatchContractsByIssue(ctx,
		util.MustParseUUID(issueID))
	if err != nil {
		return nil, fmt.Errorf("dispatch: list contracts by issue: %w", err)
	}
	contracts := make([]Contract, len(rows))
	for i, row := range rows {
		contracts[i] = *toContract(row)
	}
	return contracts, nil
}

// toContract converts a DB row to an application-level Contract.
func toContract(row db.DispatchContract) *Contract {
	c := &Contract{
		ID:            util.UUIDToString(row.ID),
		WorkspaceID:   util.UUIDToString(row.WorkspaceID),
		FromIssueID:   util.UUIDToString(row.FromIssueID),
		ToIssueID:     util.UUIDToString(row.ToIssueID),
		TriggerEvents: row.TriggerEvents,
		Status:        ContractStatus(row.Status),
		CreatedAt:     row.CreatedAt.Time,
	}
	if row.TargetIssue.Valid {
		s := util.UUIDToString(row.TargetIssue)
		c.TargetIssue = &s
	}
	if row.FulfilledAt.Valid {
		t := row.FulfilledAt.Time
		c.FulfilledAt = &t
	}
	if row.CancelledAt.Valid {
		t := row.CancelledAt.Time
		c.CancelledAt = &t
	}
	if len(row.Metadata) > 0 {
		json.Unmarshal(row.Metadata, &c.Metadata)
	}
	return c
}
