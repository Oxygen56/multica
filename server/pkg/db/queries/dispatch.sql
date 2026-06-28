-- name: CreateDispatchContract :one
INSERT INTO dispatch_contracts (workspace_id, from_issue_id, to_issue_id, trigger_events, target_issue, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDispatchContract :one
SELECT * FROM dispatch_contracts WHERE id = $1;

-- name: ListDispatchContractsByWorkspace :many
SELECT * FROM dispatch_contracts
WHERE workspace_id = $1
ORDER BY created_at DESC;

-- name: ListDispatchContractsByStatus :many
SELECT * FROM dispatch_contracts
WHERE workspace_id = $1 AND status = $2
ORDER BY created_at DESC;

-- name: ListDispatchContractsByIssue :many
SELECT * FROM dispatch_contracts
WHERE from_issue_id = $1 OR to_issue_id = $1
ORDER BY created_at DESC;

-- name: ListPendingDispatchContractsByIssue :many
SELECT * FROM dispatch_contracts
WHERE (from_issue_id = $1 OR to_issue_id = $1)
  AND status = 'pending'
  AND trigger_events && $2::text[]
ORDER BY created_at;

-- name: FulfillDispatchContract :exec
UPDATE dispatch_contracts
SET status = 'fulfilled', fulfilled_at = now(), updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: CancelDispatchContract :one
UPDATE dispatch_contracts
SET status = 'cancelled', cancelled_at = now(), updated_at = now()
WHERE id = $1 AND status IN ('pending', 'triggered')
RETURNING *;

-- name: TriggerDispatchContract :exec
UPDATE dispatch_contracts
SET status = 'triggered', updated_at = now()
WHERE id = $1 AND status = 'pending';

-- name: ListStaleDispatchContracts :many
SELECT * FROM dispatch_contracts
WHERE workspace_id = $1
  AND status = 'pending'
  AND created_at < now() - INTERVAL '24 hours';

-- name: ListCrossSquadSilentContracts :many
SELECT dc.* FROM dispatch_contracts dc
WHERE dc.workspace_id = $1
  AND dc.status = 'pending'
  AND NOT EXISTS (
    SELECT 1 FROM comments c
    WHERE c.issue_id = dc.to_issue_id
      AND c.created_at > now() - INTERVAL '48 hours'
  );
