-- Install extensions required by schema.sql in the test database.
-- citext: case-insensitive text type for email columns
CREATE EXTENSION IF NOT EXISTS citext;
