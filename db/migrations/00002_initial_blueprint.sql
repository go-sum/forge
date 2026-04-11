-- +goose Up
CREATE TABLE IF NOT EXISTS queue_jobs (
    id uuid DEFAULT gen_random_uuid(),
    queue varchar(128) NOT NULL,
    priority integer DEFAULT 20 NOT NULL,
    payload jsonb DEFAULT '{}' NOT NULL,
    status varchar(20) DEFAULT 'pending' NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    max_attempts integer DEFAULT 3 NOT NULL,
    last_error text DEFAULT '' NOT NULL,
    run_at timestamptz DEFAULT now() NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT queue_jobs_pkey PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS idx_queue_jobs_dequeue ON queue_jobs (queue, priority, run_at) WHERE ((status)::text = 'pending'::text);

CREATE INDEX IF NOT EXISTS idx_queue_jobs_reap ON queue_jobs (status, updated_at) WHERE ((status)::text = 'running'::text);

CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT uuidv7(),
    email citext NOT NULL,
    display_name varchar(255) NOT NULL,
    role varchar(50) DEFAULT 'user' NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    webauthn_id bytea,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_webauthn_id_key UNIQUE (webauthn_id)
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);

CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id uuid DEFAULT uuidv7(),
    user_id uuid NOT NULL,
    credential_id bytea NOT NULL,
    name varchar(255) DEFAULT '' NOT NULL,
    public_key bytea NOT NULL,
    public_key_alg bigint NOT NULL,
    attestation_type varchar(255) DEFAULT '' NOT NULL,
    aaguid bytea NOT NULL,
    sign_count bigint DEFAULT 0 NOT NULL,
    clone_warning boolean DEFAULT false NOT NULL,
    backup_eligible boolean DEFAULT false NOT NULL,
    backup_state boolean DEFAULT false NOT NULL,
    transports text[] DEFAULT '{}' NOT NULL,
    attachment varchar(64) DEFAULT '' NOT NULL,
    last_used_at timestamptz,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT webauthn_credentials_pkey PRIMARY KEY (id),
    CONSTRAINT webauthn_credentials_credential_id_key UNIQUE (credential_id),
    CONSTRAINT webauthn_credentials_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_user_created ON webauthn_credentials (user_id, created_at DESC);

CREATE OR REPLACE TRIGGER queue_jobs_updated_at
    BEFORE UPDATE ON queue_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE OR REPLACE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE OR REPLACE TRIGGER webauthn_credentials_updated_at
    BEFORE UPDATE ON webauthn_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS webauthn_credentials_updated_at ON webauthn_credentials;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TRIGGER IF EXISTS queue_jobs_updated_at ON queue_jobs;
DROP INDEX IF EXISTS idx_webauthn_credentials_user_created;
DROP TABLE IF EXISTS webauthn_credentials CASCADE;
DROP INDEX IF EXISTS idx_users_role;
DROP TABLE IF EXISTS users CASCADE;
DROP INDEX IF EXISTS idx_queue_jobs_reap;
DROP INDEX IF EXISTS idx_queue_jobs_dequeue;
DROP TABLE IF EXISTS queue_jobs CASCADE;
