-- +goose Up
-- Baseline migration: matches db/sql/schema.sql at initial commit.
-- Idempotent — safe to apply to databases provisioned before goose was adopted.

CREATE EXTENSION IF NOT EXISTS citext;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS update_updated_at();
