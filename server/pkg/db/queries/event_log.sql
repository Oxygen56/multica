-- name: InsertEventLog :exec
INSERT INTO event_log (workspace_id, event_type, event_data, sequence_num)
VALUES ($1, $2, $3, $4);

-- name: ListEventLogByWorkspace :many
SELECT * FROM event_log
WHERE workspace_id = $1
ORDER BY sequence_num DESC
LIMIT $2;

-- name: GetLastEventSequence :one
SELECT COALESCE(MAX(sequence_num), 0)::bigint AS last_seq
FROM event_log
WHERE workspace_id = $1;

-- name: ListEventLogByType :many
SELECT * FROM event_log
WHERE workspace_id = $1 AND event_type = $2
ORDER BY sequence_num DESC
LIMIT $3;
