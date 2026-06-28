-- name: CreateNotificationRecord :one
INSERT INTO notification_records (
    target_type, target_id, source_issue_id, type, message
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetNotificationRecord :one
SELECT * FROM notification_records WHERE id = $1;

-- name: ListPendingNotificationsByTarget :many
SELECT * FROM notification_records
WHERE target_type = $1 AND target_id = $2 AND status = 'pending'
ORDER BY created_at DESC;

-- name: ListPendingNotificationsBySource :many
SELECT * FROM notification_records
WHERE source_issue_id = $1 AND status = 'pending'
ORDER BY created_at DESC;

-- name: ListRecentNotificationsByTarget :many
SELECT * FROM notification_records
WHERE target_type = $1 AND target_id = $2
ORDER BY created_at DESC LIMIT $3;

-- name: AcknowledgeNotification :one
UPDATE notification_records SET
    status = 'acknowledged',
    acknowledged_at = now()
WHERE id = $1 RETURNING *;

-- name: CountPendingNotificationsByTarget :one
SELECT COUNT(*) FROM notification_records
WHERE target_type = $1 AND target_id = $2 AND status = 'pending';

-- name: FindDuplicateNotification :one
SELECT * FROM notification_records
WHERE target_type = $1 AND target_id = $2
  AND source_issue_id = $3 AND type = $4
  AND status = 'pending'
  AND created_at > now() - INTERVAL '5 minutes'
ORDER BY created_at DESC LIMIT 1;
