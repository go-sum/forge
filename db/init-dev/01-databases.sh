#!/bin/bash
set -e

# Create additional databases for pgschema plan and tests.
# The main database ($POSTGRES_DB) is created by the entrypoint.
# Same credentials — simplifies development.
for db in "${POSTGRES_DB}_plan" "${POSTGRES_DB}_test"; do
    echo "Creating database: $db"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        SELECT 'CREATE DATABASE "$db"'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')
        \gexec
EOSQL
done

# pgschema needs citext + pgcrypto pre-installed in the plan database
# (it doesn't run goose migrations, it applies schema.sql directly).
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "${POSTGRES_DB}_plan" <<-EOSQL
    CREATE EXTENSION IF NOT EXISTS citext;
    CREATE EXTENSION IF NOT EXISTS pgcrypto;
EOSQL
