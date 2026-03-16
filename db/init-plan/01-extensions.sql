-- Install extensions required for pgschema plan diff computation.
-- pgschema applies schema.sql here to compute diffs without touching starter.
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
