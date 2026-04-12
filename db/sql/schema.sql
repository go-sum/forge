-- Schema: starter application
--   1. task db:compose — compose migration files in db/migrations/
--   2. task db:gen — generate go queries in db/sql/queries/
--   3. task db:migrate — apply migrations from db/migrations/

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
