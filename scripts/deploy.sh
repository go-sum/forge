#!/usr/bin/env bash
# deploy.sh — platform-agnostic production deploy
#
# Performs three steps in order:
#   1. Build and push a versioned production Docker image.
#   2. Apply the schema to the production database via pgschema (runs in the dev image).
#   3. Print a reminder to restart the service (add your platform step here).
#
# Only Docker is required on the host — all tools run inside containers.
#
# Required environment variables:
#   DATABASE_URL   Production database DSN, e.g. postgres://user:pass@host/db
#   REGISTRY       Image registry prefix,  e.g. ghcr.io/my-org
#
# Optional environment variables:
#   IMAGE_NAME   Image base name (default: starter)
#   IMAGE_TAG    Image tag       (default: short git SHA)
#   APP_NAME     Dev image name  (default: <directory>-dev, matching Makefile)

set -euo pipefail

# ── Validation ────────────────────────────────────────────────────────────────

: "${DATABASE_URL:?DATABASE_URL is required}"
: "${REGISTRY:?REGISTRY is required}"

IMAGE_NAME="${IMAGE_NAME:-starter}"
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD)}"
APP_NAME="${APP_NAME:-$(basename "$(pwd)")-dev}"

FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

# ── Dev image availability ─────────────────────────────────────────────────────
# pgschema lives inside the dev image. Mirrors the _ensure-available Makefile target.

docker image inspect "${APP_NAME}" > /dev/null 2>&1 || \
    docker build --target dev -t "${APP_NAME}" .

# ── Step 1: Build and push ────────────────────────────────────────────────────

echo "==> Building production image: ${FULL_IMAGE}"
docker build --target production -t "${FULL_IMAGE}" .

echo "==> Pushing image: ${FULL_IMAGE}"
docker push "${FULL_IMAGE}"

# ── Step 2: Apply schema ──────────────────────────────────────────────────────
# Runs pgschema inside the dev container — no host installation needed.
# pgschema reads DATABASE_URL from the environment; no --dsn flag required.
# No --network flag: the external production DB is reachable from the default bridge.

echo "==> Applying schema to production database"
docker run --rm \
    -v "$(pwd):/app" \
    -w /app \
    -e DATABASE_URL="${DATABASE_URL}" \
    "${APP_NAME}" \
    pgschema apply --file db/sql/schema.sql --auto-approve

# ── Step 3: Restart ───────────────────────────────────────────────────────────
# This step is intentionally left open — it depends on your hosting platform.
# Examples:
#   Self-hosted SSH:   ssh user@host "cd /app && docker compose pull && docker compose up -d"
#   Docker Swarm:      docker service update --image "${FULL_IMAGE}" my_service
#   Fly.io:            fly deploy --image "${FULL_IMAGE}"
#
# Add your restart command here or call it from the CI workflow after this script.

echo ""
echo "==> Deploy complete."
echo "    Image:    ${FULL_IMAGE}"
echo "    Next:     restart your service to pull the new image."
