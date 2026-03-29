-- name: CreateImportBatch :one
INSERT INTO import_batches (event_id, file_path, status, rows_total, tokens)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetImportBatchByID :one
SELECT * FROM import_batches WHERE id = $1;

-- name: ListImportBatchesByEvent :many
SELECT * FROM import_batches
WHERE event_id = $1
ORDER BY created_at DESC;

-- name: UpdateImportBatchMapping :one
UPDATE import_batches
SET mapping = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateImportBatchStatus :one
UPDATE import_batches
SET status = $2, rows_ok = $3, rows_failed = $4, report = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteImportBatch :exec
DELETE FROM import_batches WHERE id = $1;

-- name: CreateParticipantRow :one
INSERT INTO participant_rows (batch_id, iin, name, payload, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListParticipantRowsByBatch :many
SELECT * FROM participant_rows
WHERE batch_id = $1
ORDER BY id;

-- name: UpdateParticipantRowStatus :exec
UPDATE participant_rows
SET status = $2, error = $3
WHERE id = $1;

-- name: CountParticipantRowsByBatch :one
SELECT count(*) FROM participant_rows WHERE batch_id = $1;
