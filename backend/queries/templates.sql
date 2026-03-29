-- name: CreateTemplate :one
INSERT INTO templates (event_id, file_path, tokens)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTemplateByEventID :one
SELECT * FROM templates
WHERE event_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteTemplatesByEventID :exec
DELETE FROM templates WHERE event_id = $1;

-- name: GetTemplateByID :one
SELECT * FROM templates WHERE id = $1;
