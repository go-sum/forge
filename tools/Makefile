# tools/Makefile — Package management and workspace fan-out targets.
#
# Loaded automatically by the root Makefile via `-include tools/Makefile`.
# This file is excluded when forge is cloned into a new app — only the root
# Makefile and the starter-zone files are copied.
#
# Targets defined here override their root-Makefile counterparts (lint, test,
# test-race, db-gen) to fan out across all workspace modules.

LINT_PATHS := $(shell awk '/use \(/{f=1;next} /\)/{f=0} f{printf "%s/... ", $$1}' go.work)
WORKSPACE  := go run ./tools/cli/workspace

# ── Workspace-aware overrides ─────────────────────────────────────────────────

lint: ## Run golangci-lint across internal/ and all pkg/* modules
	$(RUN_TOOLS) tools golangci-lint run internal/... $(LINT_PATHS)

test: ## Run tests across all workspace modules (fast, no race detector)
	$(RUN_APP) app sh -c 'DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable go run ./cli/db migrate'
	$(RUN_APP) app sh -c 'TEST_DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable TEST_KV_URL=redis://$$KV_HOST:$$KV_PORT/1 $(WORKSPACE) exec -j 4 -x tools/cli/package -- go test -count=1 ./...'

test-race: ## Run tests with race detector across all workspace modules
	$(D_COMPOSE) --profile dev up -d db kv
	$(RUN_TOOLS) tools sh -c 'DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable go run ./cli/db migrate'
	$(RUN_TOOLS) tools sh -c 'CGO_ENABLED=1 TEST_DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable TEST_KV_URL=redis://$$KV_HOST:$$KV_PORT/1 $(WORKSPACE) exec -j 1 -- go test -race -count=1 ./...'

db-gen: ## Regenerate sqlc Go code for app and all pkg/* modules
	$(RUN_TOOLS) tools sqlc generate -f .sqlc.yaml
	$(RUN_TOOLS) tools sqlc generate -f pkg/auth/pgstore/.sqlc.yaml
	$(RUN_TOOLS) tools sqlc generate -f pkg/queue/pgstore/.sqlc.yaml

# ── Deploy ────────────────────────────────────────────────────────────────────

deploy: ## Validate and deploy (AUTO=1 to auto-release and push)
	$(RUN_TOOLS) tools sh -c '\
	  git config --global user.email "$${GIT_USER_EMAIL}" && \
	  git config --global user.name "$${GIT_USER_NAME}" && \
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  go run ./tools/cli/package deploy $(if $(AUTO),--auto) $(if $(VERSION),"$(VERSION)")'

# ── Package Sync & Release ────────────────────────────────────────────────────

package-list: ## List all discovered packages
	$(RUN_TOOLS) tools go run ./tools/cli/package list

package-push: ## Push a package subtree to its mirror repo (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-push PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./tools/cli/package push "$(PACKAGE)"

package-release: ## Release a package (PACKAGE=auth|all [VERSION=v0.1.0, single name only])
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-release PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./tools/cli/package release "$(PACKAGE)" $(if $(VERSION),"$(VERSION)")

package-status: ## Show sync status for a package (PACKAGE=auth|all)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-status PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./tools/cli/package status "$(PACKAGE)"

package-sync: ## Regenerate go.prod.mod + go.prod.sum from go.mod
	$(RUN_TOOLS) tools sh -c '\
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  go run ./tools/cli/package sync'
