-- name: CreateEvent :one
INSERT INTO events (organization_id, created_by, title, date, city, description, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetEventByID :one
SELECT * FROM events WHERE id = $1;

-- name: ListEventsByOrganization :many
SELECT * FROM events
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEventsByOrganization :one
SELECT count(*) FROM events WHERE organization_id = $1;

-- name: UpdateEvent :one
UPDATE events
SET title = COALESCE(sqlc.narg('title'), title),
    date = COALESCE(sqlc.narg('date'), date),
    city = COALESCE(sqlc.narg('city'), city),
    description = COALESCE(sqlc.narg('description'), description),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM events WHERE id = $1;

-- name: ListAllEvents :many
SELECT * FROM events
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllEvents :one
SELECT count(*) FROM events;
