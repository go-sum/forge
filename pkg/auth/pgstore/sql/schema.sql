-- Schema: auth users table
-- This file is the canonical source of truth for the users table.
-- It is composed into application migrations via db/sql/schemas.yaml.
-- The application's db/sql/sqlc_schema.sql mirrors this definition for root sqlc code generation.

-- ─── Extensions ─────────────────────────────────────────────────────────────
-- citext extension is installed by the host application migration workflow.
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
    webauthn_id   BYTEA        UNIQUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

-- ─── WebAuthn credentials ────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id               UUID         PRIMARY KEY DEFAULT uuidv7(),
    user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id    BYTEA        NOT NULL UNIQUE,
    name             VARCHAR(255) NOT NULL DEFAULT '',
    public_key       BYTEA        NOT NULL,
    public_key_alg   BIGINT       NOT NULL,
    attestation_type VARCHAR(255) NOT NULL DEFAULT '',
    aaguid           BYTEA        NOT NULL,
    sign_count       BIGINT       NOT NULL DEFAULT 0,
    clone_warning    BOOLEAN      NOT NULL DEFAULT false,
    backup_eligible  BOOLEAN      NOT NULL DEFAULT false,
    backup_state     BOOLEAN      NOT NULL DEFAULT false,
    transports       TEXT[]       NOT NULL DEFAULT '{}',
    attachment       VARCHAR(64)  NOT NULL DEFAULT '',
    last_used_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_user_created ON webauthn_credentials (user_id, created_at DESC);

CREATE OR REPLACE TRIGGER webauthn_credentials_updated_at
    BEFORE UPDATE ON webauthn_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE OR REPLACE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
