#!/usr/bin/env bash
# deploy.sh — Application deployment for forge
#
# Clones the repository, builds the production Docker image, applies schema
# migrations, and starts (or restarts) the app. Infrastructure (db, kv) must
# already be running — use setup.sh for initial provisioning.
#
# Usage:
#   ./scripts/deploy.sh              # deploy from main branch
#   ./scripts/deploy.sh staging      # deploy from a specific branch
#
# Environment variables:
#   DEPLOY_DIR             Persistent directory for compose state and .env
#                          (default: /opt/forge)
#   DEPLOY_REPO            Git clone URL (required)
#   GITHUB_ACCESS_TOKEN    GitHub PAT for private Go modules (required)
#   DEPLOY_FORCE           Set to "true" to skip the commit-SHA check
#   DEPLOY_SCHEMA_TIMEOUT  Schema migration timeout in seconds (default: 120)
#
# Prerequisites:
#   - Infrastructure provisioned via setup.sh (db, kv, tools image)
#   - $DEPLOY_DIR/.env configured with production secrets

set -euo pipefail

BRANCH="${1:-main}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/forge}"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"
PROJECT_NAME="forge-prod"
APP_SERVICE="app"
TOOLS_IMAGE="${PROJECT_NAME}-tools"

COMPOSE="docker compose --project-directory ${DEPLOY_DIR} -p ${PROJECT_NAME} --env-file ${DEPLOY_DIR}/${ENV_FILE}"

# ── Validation ──────────────────────────────────────────────────────────────

command -v docker >/dev/null 2>&1 || {
    echo "ERROR: docker is not installed" >&2
    exit 1
}

docker compose version >/dev/null 2>&1 || {
    echo "ERROR: docker compose plugin is not available" >&2
    exit 1
}

docker info >/dev/null 2>&1 || {
    echo "ERROR: current user cannot access the Docker daemon." >&2
    echo "       Add this user to the docker group, use sudo, or configure rootless Docker." >&2
    exit 1
}

if [[ ! -f "${DEPLOY_DIR}/${ENV_FILE}" ]]; then
    echo "ERROR: ${DEPLOY_DIR}/${ENV_FILE} not found." >&2
    echo "       Copy .env.example and configure it:" >&2
    echo "       mkdir -p ${DEPLOY_DIR}" >&2
    echo "       cp .env.example ${DEPLOY_DIR}/${ENV_FILE}" >&2
    exit 1
fi

if [[ -z "${DEPLOY_REPO:-}" ]]; then
    echo "ERROR: DEPLOY_REPO is required" >&2
    exit 1
fi

GITHUB_ACCESS_TOKEN="${GITHUB_ACCESS_TOKEN:-$(grep '^GITHUB_ACCESS_TOKEN=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2- || echo "")}"
if [[ -z "${GITHUB_ACCESS_TOKEN:-}" ]]; then
    echo "ERROR: GITHUB_ACCESS_TOKEN is required for building." >&2
    echo "       Set it as an environment variable or add it to ${DEPLOY_DIR}/${ENV_FILE}" >&2
    exit 1
fi
export GITHUB_ACCESS_TOKEN

# ── Validate infrastructure is running ──────────────────────────────────────

if ! ${COMPOSE} ps --status running -q db 2>/dev/null | grep -q .; then
    echo "ERROR: Database is not running." >&2
    echo "       Run setup.sh first to provision infrastructure." >&2
    exit 1
fi

if ! ${COMPOSE} ps --status running -q kv 2>/dev/null | grep -q .; then
    echo "ERROR: KV store is not running." >&2
    echo "       Run setup.sh first to provision infrastructure." >&2
    exit 1
fi

# ── Clone source ────────────────────────────────────────────────────────────

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TEMP_DIR}"' EXIT

echo "==> Cloning ${DEPLOY_REPO} (branch: ${BRANCH})"
git clone --depth 1 --branch "${BRANCH}" "${DEPLOY_REPO}" "${TEMP_DIR}/src"

COMMIT_SHA="$(git -C "${TEMP_DIR}/src" rev-parse HEAD)"
echo "    Commit: ${COMMIT_SHA:0:12}"

# ── Check if already deployed ──────────────────────────────────────────────

PREV_SHA="$(cat "${DEPLOY_DIR}/.deployed_sha" 2>/dev/null || echo "")"

if [[ "${DEPLOY_FORCE:-}" != "true" && "${COMMIT_SHA}" == "${PREV_SHA}" ]]; then
    echo "==> Already deployed at ${COMMIT_SHA:0:12}, skipping"
    echo "    Use DEPLOY_FORCE=true to override"
    exit 0
fi

# ── Load version pins ──────────────────────────────────────────────────────
# shellcheck source=.versions
. "${TEMP_DIR}/src/.versions"

# ── Save previous state for rollback ────────────────────────────────────────

PREV_IMAGE_ID="$(docker inspect --format '{{.Id}}' forge:latest 2>/dev/null || echo "")"

if [[ -f "${DEPLOY_DIR}/${COMPOSE_FILE}" ]]; then
    cp "${DEPLOY_DIR}/${COMPOSE_FILE}" "${DEPLOY_DIR}/${COMPOSE_FILE}.prev"
fi

# ── Build production image ──────────────────────────────────────────────────

echo "==> Building production image (commit: ${COMMIT_SHA:0:12})"
docker build \
    --target production_target \
    --file "${TEMP_DIR}/src/docker/app/Dockerfile" \
    --build-arg GO_VERSION="${GO_VERSION}" \
    --build-arg PGSCHEMA_VERSION="${PGSCHEMA_VERSION}" \
    --build-arg TAILWIND_VERSION="${TAILWIND_VERSION}" \
    --build-arg HTMX_VERSION="${HTMX_VERSION}" \
    --build-arg HUGO_VERSION="${HUGO_VERSION}" \
    --build-arg AIR_VERSION="${AIR_VERSION}" \
    --build-arg SQLC_VERSION="${SQLC_VERSION}" \
    --build-arg GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION}" \
    --secret id=github_token,env=GITHUB_ACCESS_TOKEN \
    --tag "forge:${COMMIT_SHA}" \
    "${TEMP_DIR}/src"

docker tag "forge:${COMMIT_SHA}" forge:latest

# ── Verify tools image ─────────────────────────────────────────────────────

TOOLS_FINGERPRINT="go${GO_VERSION}-pgschema${PGSCHEMA_VERSION}-sqlc${SQLC_VERSION}"
EXISTING_FINGERPRINT="$(docker inspect --format '{{index .Config.Labels "tools.versions.fingerprint"}}' "${TOOLS_IMAGE}" 2>/dev/null || echo "")"

if [[ -z "${EXISTING_FINGERPRINT}" ]]; then
    echo "WARNING: Tools image not found. Run setup.sh to build it." >&2
    echo "         Attempting to build inline as fallback..." >&2
    docker build \
        --target cli_toolchain \
        --file "${TEMP_DIR}/src/docker/app/Dockerfile" \
        --build-arg GO_VERSION="${GO_VERSION}" \
        --build-arg PGSCHEMA_VERSION="${PGSCHEMA_VERSION}" \
        --build-arg TAILWIND_VERSION="${TAILWIND_VERSION}" \
        --build-arg HUGO_VERSION="${HUGO_VERSION}" \
        --build-arg AIR_VERSION="${AIR_VERSION}" \
        --build-arg SQLC_VERSION="${SQLC_VERSION}" \
        --build-arg GOLANGCI_LINT_VERSION="${GOLANGCI_LINT_VERSION}" \
        --label "tools.versions.fingerprint=${TOOLS_FINGERPRINT}" \
        --tag "${TOOLS_IMAGE}" \
        "${TEMP_DIR}/src"
elif [[ "${EXISTING_FINGERPRINT}" != "${TOOLS_FINGERPRINT}" ]]; then
    echo "WARNING: Tools image is outdated (have: ${EXISTING_FINGERPRINT}, need: ${TOOLS_FINGERPRINT})" >&2
    echo "         Run setup.sh to rebuild. Proceeding with existing image." >&2
fi

# ── Copy artifacts (only after build succeeds) ──────────────────────────────

mkdir -p "${DEPLOY_DIR}/scripts"

cp "${TEMP_DIR}/src/${COMPOSE_FILE}" "${DEPLOY_DIR}/${COMPOSE_FILE}"
cp -r "${TEMP_DIR}/src/docker" "${DEPLOY_DIR}/docker"
cp -r "${TEMP_DIR}/src/db" "${DEPLOY_DIR}/db"
# Copy scripts from cloned source when available, otherwise keep current install.
for script in setup.sh deploy.sh; do
    if [[ -f "${TEMP_DIR}/src/scripts/${script}" ]]; then
        cp "${TEMP_DIR}/src/scripts/${script}" "${DEPLOY_DIR}/scripts/${script}"
    fi
done
chmod +x "${DEPLOY_DIR}"/scripts/*.sh

# ── Apply schema ────────────────────────────────────────────────────────────

PROD_DB_URL="$(grep '^DATABASE_URL=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2-)"
NETWORK="${PROJECT_NAME}_app_network"
SCHEMA_TIMEOUT="${DEPLOY_SCHEMA_TIMEOUT:-120}"

echo "==> Applying schema (timeout: ${SCHEMA_TIMEOUT}s)"
if ! timeout "${SCHEMA_TIMEOUT}" docker run --rm \
    -v "${DEPLOY_DIR}:/app" \
    -w /app \
    --network "${NETWORK}" \
    -e DATABASE_URL="${PROD_DB_URL}" \
    "${TOOLS_IMAGE}" \
    pgschema apply --file db/sql/schema.sql --auto-approve; then
    echo "ERROR: Schema migration failed or timed out after ${SCHEMA_TIMEOUT}s" >&2
    exit 1
fi

# ── Start or restart app ───────────────────────────────────────────────────

if [[ -z "${PREV_IMAGE_ID}" ]]; then
    echo "==> Starting app service"
    ${COMPOSE} up -d "${APP_SERVICE}"
else
    echo "==> Restarting app service"
    ${COMPOSE} up -d --force-recreate --no-deps "${APP_SERVICE}"
fi

# ── Health check ────────────────────────────────────────────────────────────

echo "==> Waiting for health check..."
APP_PORT="$(grep '^APP_PORT=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2- || echo "8080")"
APP_PORT="${APP_PORT:-8080}"
MAX_ATTEMPTS=30
ATTEMPT=0
HEALTHY=false

while [[ ${ATTEMPT} -lt ${MAX_ATTEMPTS} ]]; do
    ATTEMPT=$((ATTEMPT + 1))
    if curl -sf "http://localhost:${APP_PORT}/health" >/dev/null 2>&1; then
        echo "    Health check passed (attempt ${ATTEMPT}/${MAX_ATTEMPTS})"
        HEALTHY=true
        break
    fi
    sleep 2
done

if [[ "${HEALTHY}" == "true" ]]; then
    # Record successful deployment
    echo "${COMMIT_SHA}" > "${DEPLOY_DIR}/.deployed_sha"

    echo ""
    echo "==> Deployment complete"
    echo "    Branch:  ${BRANCH}"
    echo "    Commit:  ${COMMIT_SHA:0:12}"
    echo "    URL:     http://localhost:${APP_PORT}"
    exit 0
fi

# ── Rollback on health check failure ────────────────────────────────────────

echo "ERROR: Health check failed after ${MAX_ATTEMPTS} attempts" >&2

echo "==> Recent logs:"
${COMPOSE} logs --tail=50 "${APP_SERVICE}" || true

if [[ -n "${PREV_IMAGE_ID}" ]]; then
    echo "==> Rolling back to previous version..."

    ${COMPOSE} stop "${APP_SERVICE}" || true

    docker tag "${PREV_IMAGE_ID}" forge:latest

    if [[ -f "${DEPLOY_DIR}/${COMPOSE_FILE}.prev" ]]; then
        cp "${DEPLOY_DIR}/${COMPOSE_FILE}.prev" "${DEPLOY_DIR}/${COMPOSE_FILE}"
    fi

    ${COMPOSE} up -d --force-recreate --no-deps "${APP_SERVICE}"

    echo "    Rolled back to previous version"
    echo "    Check the logs above to diagnose the failure"
else
    echo "    No previous version to roll back to (initial deploy)" >&2
fi

exit 1
