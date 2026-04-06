-- Auth user queries.
-- These are the persistence operations required by the auth module.
-- Parsed at init() time by schema.go via -- name: annotations.

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
