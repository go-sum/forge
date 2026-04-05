-- Schema: starter application
-- This file is the single source of truth for the DESIRED database schema state.
-- It is used by:
--   1. sqlc — reads this to generate type-safe Go code (.sqlc.yaml)
--   2. make db-diff — diffs this against the live DB to generate migration files in db/migrations/
-- Migrations are applied via goose: make db-migrate

-- ─── Extensions ─────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS citext;

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
    id            UUID         PRIMARY KEY DEFAULT uuidv7(),
    -- CITEXT enforces case-insensitive uniqueness: user@example.com == User@Example.com
    email         CITEXT       NOT NULL UNIQUE,
    display_name  VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'user',
    verified      BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_role ON users (role);

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
