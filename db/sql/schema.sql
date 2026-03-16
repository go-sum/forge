-- Schema: starter application
-- This file is the single source of truth for the database schema.
-- pgschema uses this file to compute schema diffs (make db-plan / make db-apply).
-- Note: extensions (pgcrypto, citext) are managed separately via db/init/01-extensions.sql

-- ─── Trigger function ───────────────────────────────────────────────────────
-- Automatically sets updated_at = NOW() on any UPDATE row.
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ─── Users ──────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    -- CITEXT enforces case-insensitive uniqueness: user@example.com == User@Example.com
    email         CITEXT       NOT NULL UNIQUE,
    display_name  VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_role ON users (role);

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ─── Passwords ───────────────────────────────────────────────────────────────
-- Append-only credential history. Current password = ORDER BY created_at DESC LIMIT 1.
CREATE TABLE passwords (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hash       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Simple index retained for FK integrity checks and joins
CREATE INDEX idx_passwords_user_id ON passwords (user_id);

-- Composite index for "current password" queries: ORDER BY created_at DESC LIMIT 1
-- Eliminates the sort pass on the login hot-path and GetCurrentPasswordByUserID.
CREATE INDEX idx_passwords_user_id_created_at ON passwords (user_id, created_at DESC);
