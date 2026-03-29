-- name: CreateUser :one
INSERT INTO users (username, email, password, iin, role, language)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, password, iin, role, is_active, language, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByIIN :one
SELECT id, username, email, password, iin, role, is_active, language, created_at, updated_at
FROM users
WHERE iin = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET email = COALESCE(sqlc.narg('email'), email),
    iin = COALESCE(sqlc.narg('iin'), iin),
    language = COALESCE(sqlc.narg('language'), language),
    updated_at = now()
WHERE id = sqlc.arg('id')
RETURNING id, username, email, iin, role, is_active, language, created_at, updated_at;

-- name: ListUsers :many
SELECT id, username, email, iin, role, is_active, language, created_at, updated_at
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: ListTeacherStudents :many
SELECT id, teacher_id, student_iin, created_at
FROM teacher_students
WHERE teacher_id = $1
ORDER BY created_at DESC;

-- name: AddTeacherStudent :one
INSERT INTO teacher_students (teacher_id, student_iin)
VALUES ($1, $2)
RETURNING id, teacher_id, student_iin, created_at;

-- name: RemoveTeacherStudent :exec
DELETE FROM teacher_students
WHERE teacher_id = $1 AND student_iin = $2;
