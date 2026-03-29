-- name: CreateCertificate :one
INSERT INTO certificates (event_id, organization_id, iin, name, code, pdf_path, status, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCertificateByID :one
SELECT * FROM certificates WHERE id = $1;

-- name: GetCertificateByCode :one
SELECT * FROM certificates WHERE code = $1;

-- name: ListCertificatesByEvent :many
SELECT * FROM certificates
WHERE event_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCertificatesByEvent :one
SELECT count(*) FROM certificates WHERE event_id = $1;

-- name: ListCertificatesByIIN :many
SELECT c.*, e.title as event_title, o.name as org_name
FROM certificates c
JOIN events e ON e.id = c.event_id
LEFT JOIN organizations o ON o.id = c.organization_id
WHERE c.iin = $1 AND c.status = 'valid'
ORDER BY c.created_at DESC;

-- name: UpdateCertificateStatus :one
UPDATE certificates
SET status = $2, revoked_reason = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCertificate :exec
DELETE FROM certificates WHERE id = $1;

-- name: ListCertificatesByOrganization :many
SELECT * FROM certificates
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCertificatesByOrganization :one
SELECT count(*) FROM certificates WHERE organization_id = $1;

-- name: SearchCertificatesByIIN :many
SELECT c.id, c.event_id, c.iin, c.name, c.code, c.status, c.created_at,
       e.title as event_title, o.name as org_name
FROM certificates c
JOIN events e ON e.id = c.event_id
LEFT JOIN organizations o ON o.id = c.organization_id
WHERE c.iin = $1
ORDER BY c.created_at DESC;
