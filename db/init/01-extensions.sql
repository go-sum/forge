-- Install extensions required by schema.sql in the starter database.
-- pgcrypto: gen_random_uuid() for UUID primary keys
-- citext: case-insensitive text type for email columns
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
