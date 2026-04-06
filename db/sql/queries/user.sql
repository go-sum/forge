-- Admin user queries.
-- Auth-related queries (CreateUser, GetUserByEmail, UpdateUserEmail) are owned
-- by pkg/auth/pgstore/sql/queries.sql and managed by that package.
-- These queries support admin user management only.
-- sqlc annotations control the return type:
--   :one  → returns a single struct
--   :many → returns []struct
--   :exec → returns error only

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

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
