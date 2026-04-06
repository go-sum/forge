-- Schema: auth users table
-- This file is the canonical source of truth for the users table.
-- It is embedded by pkg/auth/pgstore and used for idempotent schema installation.
-- The application's db/sql/schema.sql mirrors this definition for sqlc code generation.

-- ─── Extensions ─────────────────────────────────────────────────────────────
-- citext extension must be installed by the host application before Install() is called.
-- See db/init/01-extensions.sql.

-- ─── Trigger function ───────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ─── Users ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuidv7(),
    -- CITEXT enforces case-insensitive uniqueness: user@example.com == User@Example.com
    email         CITEXT       NOT NULL UNIQUE,
    display_name  VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'user',
    verified      BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE OR REPLACE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
