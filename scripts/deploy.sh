#!/usr/bin/env bash
# deploy.sh — Self-hosted pull-and-run deployment
#
# Clones the repository into a temporary directory, builds the production
# Docker image, applies schema migrations, and starts (or restarts) the
# production stack via docker-compose.yml.
#
# Usage:
#   ./scripts/deploy.sh              # deploy from main branch
#   ./scripts/deploy.sh staging      # deploy from a specific branch
#
# Environment variables:
#   DEPLOY_DIR    Persistent directory for compose state and .env
#                 (default: /opt/forge)
#   DEPLOY_REPO   Git clone URL (default: auto-detected from current repo)
#
# Prerequisites:
#   - Docker and Docker Compose installed on the server
#   - $DEPLOY_DIR/.env configured with production secrets
#     (see .env.example)

set -euo pipefail

BRANCH="${1:-main}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/forge}"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"
PROJECT_NAME="forge-prod"
APP_SERVICE="app"

# ── Validation ──────────────────────────────────────────────────────────────

command -v docker >/dev/null 2>&1 || {
    echo "ERROR: docker is not installed" >&2
    exit 1
}

docker compose version >/dev/null 2>&1 || {
    echo "ERROR: docker compose plugin is not available" >&2
    exit 1
}

if [[ ! -f "${DEPLOY_DIR}/${ENV_FILE}" ]]; then
    echo "ERROR: ${DEPLOY_DIR}/${ENV_FILE} not found." >&2
    echo "       Copy .env.example and configure it:" >&2
    echo "       mkdir -p ${DEPLOY_DIR}" >&2
    echo "       cp .env.example ${DEPLOY_DIR}/${ENV_FILE}" >&2
    exit 1
fi

# ── Resolve repository URL ─────────────────────────────────────────────────

if [[ -z "${DEPLOY_REPO:-}" ]]; then
    if git rev-parse --git-dir >/dev/null 2>&1; then
        DEPLOY_REPO="$(git remote get-url origin 2>/dev/null)" || true
    fi
fi

if [[ -z "${DEPLOY_REPO:-}" ]]; then
    echo "ERROR: DEPLOY_REPO is required (no git origin found)" >&2
    exit 1
fi

# ── Step 1: Clone into temp directory ───────────────────────────────────────

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TEMP_DIR}"' EXIT

echo "==> Cloning ${DEPLOY_REPO} (branch: ${BRANCH})"
git clone --depth 1 --branch "${BRANCH}" "${DEPLOY_REPO}" "${TEMP_DIR}/src"

COMMIT_SHA="$(git -C "${TEMP_DIR}/src" rev-parse --short HEAD)"
echo "    Commit: ${COMMIT_SHA}"

# ── Step 2: Prepare deploy directory ────────────────────────────────────────

mkdir -p "${DEPLOY_DIR}"

# Copy compose file (base only — no override), db init scripts, and this script
cp "${TEMP_DIR}/src/${COMPOSE_FILE}" "${DEPLOY_DIR}/${COMPOSE_FILE}"
cp -r "${TEMP_DIR}/src/db" "${DEPLOY_DIR}/db"
cp "${TEMP_DIR}/src/scripts/deploy.sh" "${DEPLOY_DIR}/deploy.sh"
chmod +x "${DEPLOY_DIR}/deploy.sh"

# ── Step 3: Build the production image ──────────────────────────────────────
# Build runs from the cloned source (needs Dockerfile + full context).

echo "==> Building production image (commit: ${COMMIT_SHA})"
docker build \
    --target production_target \
    --tag forge:latest \
    "${TEMP_DIR}/src"

# ── Step 4: Detect initial vs update deployment ────────────────────────────

COMPOSE="docker compose --project-directory ${DEPLOY_DIR} -p ${PROJECT_NAME} --env-file ${DEPLOY_DIR}/${ENV_FILE}"
IS_INITIAL=false

if ! ${COMPOSE} ps --status running -q db 2>/dev/null | grep -q .; then
    IS_INITIAL=true
    echo "==> Initial deployment detected"
else
    echo "==> Update deployment detected"
fi

# ── Step 5: Ensure database and KV are running ─────────────────────────────

if [[ "${IS_INITIAL}" == "true" ]]; then
    echo "==> Starting database and KV services"
    ${COMPOSE} up -d --wait db kv
else
    echo "==> Database and KV already running"
fi

# ── Step 6: Apply schema ───────────────────────────────────────────────────
# pgschema lives in the dev image (production image is minimal).
# Build the dev image on demand, connect it to the prod network.

DEV_IMAGE="${PROJECT_NAME}-dev"
docker image inspect "${DEV_IMAGE}" >/dev/null 2>&1 || {
    echo "==> Building dev image for schema migration"
    docker build \
        --target dev_target \
        --tag "${DEV_IMAGE}" \
        "${TEMP_DIR}/src"
}

PROD_DB_URL="$(grep '^DATABASE_URL=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2-)"
NETWORK="${PROJECT_NAME}_app_network"

echo "==> Applying schema"
docker run --rm \
    -v "${DEPLOY_DIR}:/app" \
    -w /app \
    --network "${NETWORK}" \
    -e DATABASE_URL="${PROD_DB_URL}" \
    "${DEV_IMAGE}" \
    pgschema apply --file db/sql/schema.sql --auto-approve

# ── Step 7: Start or restart app ───────────────────────────────────────────

if [[ "${IS_INITIAL}" == "true" ]]; then
    echo "==> Starting app service"
    ${COMPOSE} up -d "${APP_SERVICE}"
else
    echo "==> Restarting app service"
    ${COMPOSE} up -d --force-recreate --no-deps "${APP_SERVICE}"
fi

# ── Step 8: Health check ───────────────────────────────────────────────────

echo "==> Waiting for health check..."
APP_PORT="$(grep '^APP_PORT=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2- || echo "8080")"
APP_PORT="${APP_PORT:-8080}"
MAX_ATTEMPTS=30
ATTEMPT=0

while [[ ${ATTEMPT} -lt ${MAX_ATTEMPTS} ]]; do
    ATTEMPT=$((ATTEMPT + 1))
    if curl -sf "http://localhost:${APP_PORT}" >/dev/null 2>&1; then
        echo "    Health check passed (attempt ${ATTEMPT}/${MAX_ATTEMPTS})"
        break
    fi
    if [[ ${ATTEMPT} -eq ${MAX_ATTEMPTS} ]]; then
        echo "ERROR: Health check failed after ${MAX_ATTEMPTS} attempts" >&2
        echo "==> Recent logs:"
        ${COMPOSE} logs --tail=50 "${APP_SERVICE}"
        exit 1
    fi
    sleep 2
done

# ── Done ───────────────────────────────────────────────────────────────────

echo ""
echo "==> Deployment complete"
echo "    Branch:  ${BRANCH}"
echo "    Commit:  ${COMMIT_SHA}"
echo "    URL:     http://localhost:${APP_PORT}"
