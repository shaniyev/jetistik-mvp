-- name: CreateAuditLog :one
INSERT INTO audit_logs (actor_id, action, object_type, object_id, meta)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAuditLogs :many
SELECT al.*, u.username as actor_username
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
ORDER BY al.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT count(*) FROM audit_logs;

-- name: ListAuditLogsByActor :many
SELECT al.*, u.username as actor_username
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
WHERE al.actor_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByAction :many
SELECT al.*, u.username as actor_username
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
WHERE al.action = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByObject :many
SELECT al.*, u.username as actor_username
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
WHERE al.object_type = $1 AND al.object_id = $2
ORDER BY al.created_at DESC
LIMIT $3 OFFSET $4;
