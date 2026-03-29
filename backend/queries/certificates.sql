-- name: CreateCertificate :one
INSERT INTO certificates (event_id, organization_id, iin, name, code, pdf_path, status, payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCertificateByID :one
SELECT * FROM certificates WHERE id = $1;

-- name: GetCertificateByCode :one
SELECT * FROM certificates WHERE code = $1;

-- name: GetCertificateByCodeWithDetails :one
SELECT c.*, e.title as event_title, o.name as org_name
FROM certificates c
JOIN events e ON e.id = c.event_id
LEFT JOIN organizations o ON o.id = c.organization_id
WHERE c.code = $1;

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

-- name: ListAllCertificates :many
SELECT * FROM certificates
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllCertificates :one
SELECT count(*) FROM certificates;

-- name: ListCertificatesByUserID :many
SELECT c.id, c.event_id, c.organization_id, c.iin, c.name, c.code, c.pdf_path, c.status, c.revoked_reason, c.payload, c.created_at, c.updated_at, e.title as event_title, o.name as org_name
FROM certificates c
JOIN events e ON e.id = c.event_id
LEFT JOIN organizations o ON o.id = c.organization_id
WHERE c.iin = (SELECT users.iin FROM users WHERE users.id = $1 AND users.iin IS NOT NULL AND users.iin != '')
ORDER BY c.created_at DESC;

-- name: ListCertificatesForTeacher :many
SELECT c.id, c.event_id, c.organization_id, c.iin, c.name, c.code, c.pdf_path, c.status, c.revoked_reason, c.payload, c.created_at, c.updated_at, e.title as event_title, o.name as org_name
FROM certificates c
JOIN events e ON e.id = c.event_id
LEFT JOIN organizations o ON o.id = c.organization_id
WHERE c.iin IN (
  SELECT ts.student_iin FROM teacher_students ts WHERE ts.teacher_id = $1
  UNION
  SELECT u.iin FROM users u WHERE u.id = $1 AND u.iin IS NOT NULL AND u.iin != ''
)
ORDER BY c.created_at DESC;
