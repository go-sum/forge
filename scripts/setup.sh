#!/usr/bin/env bash
# setup.sh — Infrastructure provisioning for forge
#
# Sets up the database, KV store, and tools image on a remote server.
# Run once on initial server setup, or again when infrastructure changes
# (PG_VERSION bump, new service, tool version changes in .versions).
#
# Usage:
#   ./scripts/setup.sh              # provision from main branch
#   ./scripts/setup.sh staging      # provision from a specific branch
#
# Environment variables:
#   DEPLOY_DIR    Persistent directory for compose state and .env
#                 (default: /opt/forge)
#   DEPLOY_REPO   Git clone URL (required)
#   GITHUB_ACCESS_TOKEN  GitHub PAT for private Go modules (required)
#
# Prerequisites:
#   - Docker and Docker Compose installed on the server
#   - $DEPLOY_DIR/.env configured with production secrets

set -euo pipefail

BRANCH="${1:-main}"
DEPLOY_DIR="${DEPLOY_DIR:-/opt/forge}"
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"
PROJECT_NAME="forge-prod"
TOOLS_IMAGE="${PROJECT_NAME}-tools"

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
    echo "ERROR: GITHUB_ACCESS_TOKEN is required." >&2
    echo "       Set it as an environment variable or add it to ${DEPLOY_DIR}/${ENV_FILE}" >&2
    exit 1
fi
export GITHUB_ACCESS_TOKEN

# ── Clone source ────────────────────────────────────────────────────────────

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TEMP_DIR}"' EXIT

echo "==> Cloning ${DEPLOY_REPO} (branch: ${BRANCH})"
git clone --depth 1 --branch "${BRANCH}" "${DEPLOY_REPO}" "${TEMP_DIR}/src"

COMMIT_SHA="$(git -C "${TEMP_DIR}/src" rev-parse HEAD)"
echo "    Commit: ${COMMIT_SHA:0:12}"

# shellcheck source=.versions
. "${TEMP_DIR}/src/.versions"

# ── Copy infrastructure files to DEPLOY_DIR ─────────────────────────────────

mkdir -p "${DEPLOY_DIR}/scripts"

cp "${TEMP_DIR}/src/${COMPOSE_FILE}" "${DEPLOY_DIR}/${COMPOSE_FILE}"
cp -r "${TEMP_DIR}/src/docker" "${DEPLOY_DIR}/docker"
cp -r "${TEMP_DIR}/src/db" "${DEPLOY_DIR}/db"
# Copy scripts from cloned source when available, otherwise from current install.
for script in setup.sh deploy.sh; do
    if [[ -f "${TEMP_DIR}/src/scripts/${script}" ]]; then
        cp "${TEMP_DIR}/src/scripts/${script}" "${DEPLOY_DIR}/scripts/${script}"
    elif [[ -f "${DEPLOY_DIR}/scripts/${script}" ]]; then
        : # already in place
    fi
done
chmod +x "${DEPLOY_DIR}"/scripts/*.sh

# ── Build infrastructure images ─────────────────────────────────────────────

COMPOSE="docker compose --project-directory ${DEPLOY_DIR} -p ${PROJECT_NAME} --env-file ${DEPLOY_DIR}/${ENV_FILE}"

echo "==> Building database image (PG ${PG_VERSION})"
${COMPOSE} build db

echo "==> Building KV image (Dragonfly ${DRAGONFLY_VERSION})"
${COMPOSE} build kv

# ── Build tools image (version-fingerprinted) ───────────────────────────────

TOOLS_FINGERPRINT="go${GO_VERSION}-pgschema${PGSCHEMA_VERSION}-sqlc${SQLC_VERSION}"

EXISTING_FINGERPRINT="$(docker inspect --format '{{index .Config.Labels "tools.versions.fingerprint"}}' "${TOOLS_IMAGE}" 2>/dev/null || echo "")"

if [[ "${EXISTING_FINGERPRINT}" == "${TOOLS_FINGERPRINT}" ]]; then
    echo "==> Tools image already up to date (${TOOLS_FINGERPRINT})"
else
    echo "==> Building tools image (${TOOLS_FINGERPRINT})"
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
fi

# ── Start infrastructure services ───────────────────────────────────────────

echo "==> Starting database and KV services"
${COMPOSE} up -d --wait db kv
echo "    Database and KV are healthy"

# ── Record state ────────────────────────────────────────────────────────────

sha256sum "${TEMP_DIR}/src/.versions" | cut -d' ' -f1 > "${DEPLOY_DIR}/.versions_hash"

echo ""
echo "==> Infrastructure provisioning complete"
echo "    Deploy dir:  ${DEPLOY_DIR}"
echo "    Tools:       ${TOOLS_FINGERPRINT}"
echo ""
echo "    Next: run ./scripts/deploy.sh to deploy the application"
