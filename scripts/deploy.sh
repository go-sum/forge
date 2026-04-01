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

docker info >/dev/null 2>&1 || {
    echo "ERROR: current user cannot access the Docker daemon." >&2
    echo "       docker commands require access to /var/run/docker.sock or a rootless Docker socket." >&2
    echo "       On Linux, fix one of these before deploying:" >&2
    echo "       1. Add this user to the docker group and start a new login session" >&2
    echo "       2. Run the deploy via sudo/root" >&2
    echo "       3. Use a rootless Docker daemon for this user" >&2
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

# ── Load version pins from cloned source ────────────────────────────────────
# shellcheck source=.versions
. "${TEMP_DIR}/src/.versions"

# ── Step 2: Prepare deploy directory ────────────────────────────────────────

mkdir -p "${DEPLOY_DIR}"

# Copy compose file (base only — no override), db init scripts, and this script
cp "${TEMP_DIR}/src/${COMPOSE_FILE}" "${DEPLOY_DIR}/${COMPOSE_FILE}"
cp -r "${TEMP_DIR}/src/db" "${DEPLOY_DIR}/db"
cp "${TEMP_DIR}/src/scripts/deploy.sh" "${DEPLOY_DIR}/deploy.sh"
chmod +x "${DEPLOY_DIR}/deploy.sh"

# ── Step 3: Build the production image ──────────────────────────────────────
# Build runs from the cloned source (needs Dockerfile + full context).

GITHUB_ACCESS_TOKEN="${GITHUB_ACCESS_TOKEN:-$(grep '^GITHUB_ACCESS_TOKEN=' "${DEPLOY_DIR}/${ENV_FILE}" | cut -d= -f2-)}"

echo "==> Building production image (commit: ${COMMIT_SHA})"
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

TOOLS_IMAGE="${PROJECT_NAME}-tools"
docker image inspect "${TOOLS_IMAGE}" >/dev/null 2>&1 || {
    echo "==> Building tools image for schema migration"
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
        --tag "${TOOLS_IMAGE}" \
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
    "${TOOLS_IMAGE}" \
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
