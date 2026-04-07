include .gitenv
include .versions
export

PROJECT_NAME    ?= $(notdir $(CURDIR))
APP_NAME        := $(PROJECT_NAME)-dev
PACKAGE         ?=
VERSION         ?=

TOOLS_IMAGE  := $(PROJECT_NAME)-tools
TOOLS_DEV    := $(TOOLS_IMAGE):dev
TOOLS_PROD   := $(TOOLS_IMAGE):prod
TOOLS_DIR    := docker/tools
LINT_PATHS   := $(shell awk '/use \(/{f=1;next} /\)/{f=0} f{printf "%s/... ", $$1}' go.work)

# ── Compose helpers ──────────────────────────────────────────────────────────
# Tools and app commands run via compose, which auto-builds images on first use.
D_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.dev.yml --project-name $(PROJECT_NAME)
RUN_TOOLS := $(D_COMPOSE) --profile tools run --rm
RUN_APP   := $(D_COMPOSE) --profile dev run --rm
WORKSPACE := go run ./cli/workspace

.PHONY: help \
        build clean lint vet hash-air-csp \
        db-compose db-gen db-migrate db-rollback db-status \
        assets \
        deploy \
        package-list package-push package-release package-status package-sync \
        dev prod test test-race \
        docker-build docker-dev docker-down docker-logs docker-prune docker-up \
        dev-tools prod-tools

# ── Build & Quality ───────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build production binary into bin/
	$(RUN_TOOLS) tools sh -c 'CGO_ENABLED=0 go build -o ./bin/server ./cmd/server'

clean: ## Remove build artifacts
	rm -rf ./bin ./tmp ./public/css/app.css

hash-air-csp: ## Recompute CSP hash for air's proxy script and update config/app.development.yaml
	$(RUN_TOOLS) tools go run ./cli/util hash-air-csp

lint: ## Run golangci-lint
	$(RUN_TOOLS) tools golangci-lint run internal/... $(LINT_PATHS)

test: ## Run tests (fast, no race detector — skips cli/package which needs CGO)
	$(RUN_APP) app sh -c 'PGDATABASE=$${PGDATABASE}_test go run ./cli/db migrate'
	$(RUN_APP) app sh -c 'TEST_DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable TEST_KV_URL=redis://$$KV_HOST:$$KV_PORT/1 $(WORKSPACE) exec -j 4 -x cli/package -- go test -count=1 ./...'

test-race: ## Run tests with race detector (uses tools container with CGO)
	$(D_COMPOSE) --profile dev up -d db kv
	$(RUN_TOOLS) tools sh -c 'PGDATABASE=$${PGDATABASE}_test go run ./cli/db migrate'
	$(RUN_TOOLS) tools sh -c 'CGO_ENABLED=1 TEST_DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable TEST_KV_URL=redis://$$KV_HOST:$$KV_PORT/1 $(WORKSPACE) exec -j 1 -- go test -race -count=1 ./...'

# ── Database ──────────────────────────────────────────────────────────────────

db-compose: ## Compose schemas and generate migration with diff (NAME=add_queue_jobs)
	@test -n "$(NAME)" || { echo "error: NAME is required  e.g. make db-compose NAME=add_queue_jobs" >&2; exit 1; }
	$(RUN_APP) \
	  -e PGSCHEMA_PLAN_HOST=$$PGHOST \
	  -e PGSCHEMA_PLAN_DB=$${PGDATABASE}_plan \
	  -e PGSCHEMA_PLAN_USER=$$PGUSER \
	  -e PGSCHEMA_PLAN_PASSWORD=$$PGPASSWORD \
	  app go run ./cli/db compose "$(NAME)"

db-gen: ## Regenerate sqlc Go code from SQL queries
	$(RUN_TOOLS) tools sqlc generate -f pkg/auth/pgstore/.sqlc.yaml
	$(RUN_TOOLS) tools sqlc generate -f pkg/queue/pgstore/.sqlc.yaml

db-migrate: ## Apply pending migrations
	$(RUN_APP) app go run ./cli/db migrate

db-rollback: ## Rollback the last migration
	$(RUN_APP) app go run ./cli/db rollback

db-status: ## Show migration status
	$(RUN_APP) app go run ./cli/db status

# ── Assets ────────────────────────────────────────────────────────────────────

assets: ## Build all generated frontend assets
	$(RUN_TOOLS) -e HTMX_VERSION=$(HTMX_VERSION) tools go run ./cli/build assets --minify
	$(RUN_TOOLS) tools go run ./cli/build docs
	$(RUN_TOOLS) tools go run ./cli/build sprites

# ── Deploy ────────────────────────────────────────────────────────────────────

deploy: ## Validate and deploy (AUTO=1 to auto-release and push)
	$(RUN_TOOLS) tools sh -c '\
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  go run ./cli/package deploy $(if $(AUTO),--auto) $(if $(VERSION),"$(VERSION)")'

# ── Package Sync & Release ────────────────────────────────────────────────────

package-list: ## List all discovered packages
	$(RUN_TOOLS) tools go run ./cli/package list

package-push: ## Push a package subtree to its mirror repo (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-push PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./cli/package push "$(PACKAGE)"

package-release: ## Release a package (PACKAGE=auth|all [VERSION=v0.1.0, single name only])
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-release PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./cli/package release "$(PACKAGE)" $(if $(VERSION),"$(VERSION)")

package-status: ## Show sync status for a package (PACKAGE=auth|all)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-status PACKAGE=auth" >&2; exit 1; }
	$(RUN_TOOLS) tools go run ./cli/package status "$(PACKAGE)"

package-sync: ## Regenerate go.prod.mod + go.prod.sum from go.mod
	$(RUN_TOOLS) tools sh -c '\
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  go run ./cli/package sync'

# ── Toolchain ────────────────────────────────────────────────────────────────

dev-tools: ## Rebuild dev toolchain image
	$(D_COMPOSE) --profile tools build tools

prod-tools: ## Rebuild production toolchain image
	$(D_COMPOSE) --profile tools build prod-tools

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: ## Start all services with hot-reload
	$(D_COMPOSE) --profile dev up --build

prod: docker-build ## Build and start the production stack locally
	docker compose up -d

docker-build: ## Build production Docker image
	@GITHUB_ACCESS_TOKEN="$${GITHUB_ACCESS_TOKEN:-$$(grep '^GITHUB_ACCESS_TOKEN=' .env 2>/dev/null | cut -d= -f2-)}" && \
	  export GITHUB_ACCESS_TOKEN && \
	  docker build --target production_target \
	    --file docker/app/Dockerfile \
	    --build-arg GO_VERSION=$(GO_VERSION) \
	    --build-arg HTMX_VERSION=$(HTMX_VERSION) \
	    --build-arg APP_VERSION=$(APP_VERSION) \
	    --build-arg TOOLS_PROD_IMAGE=$(TOOLS_PROD) \
	    --secret id=github_token,env=GITHUB_ACCESS_TOKEN \
	    -t forge:latest .

docker-dev: ## Build dev Docker image
	docker build --target dev_target \
	  --file docker/app/Dockerfile \
	  --build-arg GO_VERSION=$(GO_VERSION) \
	  --build-arg TOOLS_DEV_IMAGE=$(TOOLS_DEV) \
	  -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) --profile dev down $(ARGS)

caddy-up: ## Start Caddy reverse proxy for local production testing
	docker compose -f docker/caddy/docker-compose.yml up -d

caddy-down: ## Stop Caddy reverse proxy
	docker compose -f docker/caddy/docker-compose.yml down

docker-logs: ## Follow container logs
	$(D_COMPOSE) --profile dev logs -f

docker-prune: ## Remove all project containers, images, networks, and volumes
	$(D_COMPOSE) --profile dev --profile tools down -v --rmi all --remove-orphans
	@docker rmi forge:latest 2>/dev/null || true

docker-up: ## Apply schema, then start containers in background
	$(D_COMPOSE) --profile dev up -d $(ARGS)
