-- Password queries
-- Passwords are append-only: INSERT for new password, old rows retained as history.
-- Current password = ORDER BY created_at DESC LIMIT 1.

-- name: CreatePassword :one
INSERT INTO passwords (user_id, hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetCurrentPasswordByUserID :one
SELECT * FROM passwords
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetCurrentPasswordByEmail :one
-- Used for signin: fetches the current hash for a given email in one query.
SELECT p.* FROM passwords p
JOIN users u ON u.id = p.user_id
WHERE u.email = $1
ORDER BY p.created_at DESC
LIMIT 1;

-- name: ListPasswordsByUserID :many
-- Returns full password history for a user, newest first.
SELECT * FROM passwords
WHERE user_id = $1
ORDER BY created_at DESC;
