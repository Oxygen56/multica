-- name: CreateDispatchContract :one
INSERT INTO dispatch_contracts (
    creator_agent_id, from_issue_id, to_issue_id,
    from_squad_id, to_squad_id, trigger_events,
    target_issue, notify_assignee, notify_creator, include_summary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetDispatchContract :one
SELECT * FROM dispatch_contracts WHERE contract_id = $1;

-- name: GetDispatchContractByToIssue :one
SELECT * FROM dispatch_contracts WHERE to_issue_id = $1;

-- name: GetDispatchContractByFromIssue :one
SELECT * FROM dispatch_contracts WHERE from_issue_id = $1;

-- name: ListDispatchContractsByStatus :many
SELECT * FROM dispatch_contracts WHERE status = $1 ORDER BY created_at DESC;

-- name: ListPendingDispatchContracts :many
SELECT * FROM dispatch_contracts WHERE status = 'pending' ORDER BY created_at DESC;

-- name: UpdateDispatchContractStatus :one
UPDATE dispatch_contracts SET
    status = $2,
    fulfilled_at = CASE WHEN $2 = 'fulfilled' THEN now() ELSE fulfilled_at END
WHERE contract_id = $1 RETURNING *;

-- name: FulfillDispatchContract :one
UPDATE dispatch_contracts SET
    status = 'fulfilled',
    fulfilled_at = now()
WHERE contract_id = $1 AND status = 'pending' RETURNING *;

-- name: CancelDispatchContract :one
UPDATE dispatch_contracts SET
    status = 'cancelled'
WHERE contract_id = $1 AND status = 'pending' RETURNING *;

-- name: ListDispatchContractsByFromSquad :many
SELECT * FROM dispatch_contracts WHERE from_squad_id = $1 ORDER BY created_at DESC;
