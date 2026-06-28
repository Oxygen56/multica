-- name: CreateNotificationRecord :exec
INSERT INTO notification_records (workspace_id, notification_type, source_issue_id, target_issue_id, rule_id, status, details)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListNotificationRecordsByIssue :many
SELECT * FROM notification_records
WHERE target_issue_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: AcknowledgeNotificationRecord :exec
UPDATE notification_records
SET status = 'acknowledged', acknowledged_at = now()
WHERE id = $1;
