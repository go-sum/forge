-- Auth user queries.
-- These are the persistence operations required by the auth module.
-- Query names and return modes follow sqlc conventions (-- name: X :one/:exec).

-- name: CreateUser :one
INSERT INTO users (email, display_name, role, verified)
VALUES ($1, $2, $3, $4)
RETURNING id, email, display_name, role, verified, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, display_name, role, verified, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, display_name, role, verified, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateUserEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING id, email, display_name, role, verified, created_at, updated_at;

-- ─── Admin operations ───────────────────────────────────────────────────────

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
-- COALESCE(NULLIF(value, ''), column) means: empty string = keep existing value.
-- sqlc.arg() assigns named parameters independent of positional $N numbering.
UPDATE users
SET
    email        = COALESCE(NULLIF(sqlc.arg(email)::text, ''), email),
    display_name = COALESCE(NULLIF(sqlc.arg(display_name)::text, ''), display_name),
    role         = COALESCE(NULLIF(sqlc.arg(role)::text, ''), role)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: HasAdminUser :one
SELECT EXISTS(SELECT 1 FROM users WHERE role = 'admin');
