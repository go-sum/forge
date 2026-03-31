-- Schema: starter application
-- This file is the single source of truth for the database schema.
-- pgschema uses this file to compute schema diffs (make db-plan / make db-apply).
-- Note: extensions (citext) are managed separately via db/init/01-extensions.sql

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
