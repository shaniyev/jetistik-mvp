-- name: CreateOrganization :one
INSERT INTO organizations (name, domain, logo_path, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: CountOrganizations :one
SELECT count(*) FROM organizations;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = COALESCE(sqlc.narg('name'), name),
    domain = COALESCE(sqlc.narg('domain'), domain),
    logo_path = COALESCE(sqlc.narg('logo_path'), logo_path),
    status = COALESCE(sqlc.narg('status'), status),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;

-- name: ListOrganizationMembers :many
SELECT om.id, om.organization_id, om.user_id, om.role, om.created_at,
       u.username, u.email
FROM organization_members om
JOIN users u ON u.id = om.user_id
WHERE om.organization_id = $1
ORDER BY om.created_at DESC;

-- name: AddOrganizationMember :one
INSERT INTO organization_members (organization_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: GetOrganizationMember :one
SELECT * FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: GetUserOrganization :one
SELECT o.* FROM organizations o
JOIN organization_members om ON om.organization_id = o.id
WHERE om.user_id = $1
LIMIT 1;
