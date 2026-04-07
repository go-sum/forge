-- Schema: root sqlc input
-- This file is used only by root sqlc generation for internal/repository.
-- It is not part of db/sql/schemas.yaml and does not participate in migration
-- composition.
--
-- Application-specific tables that need sqlc code generation belong here.
-- User tables are owned by pkg/auth/pgstore/sql/schema.sql.

-- ─── Extensions ─────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS citext;

-- ─── Trigger function ───────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
