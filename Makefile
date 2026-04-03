include .versions
export

PROJECT_NAME    ?= $(notdir $(CURDIR))
APP_NAME        := $(PROJECT_NAME)-dev
PACKAGE         ?=
VERSION         ?=

APP_NETWORK  := $(PROJECT_NAME)_app_network

# ── Tool images ──────────────────────────────────────────────────────────────
TOOLS_IMAGE  := $(PROJECT_NAME)-tools
TOOLS_DEV    := $(TOOLS_IMAGE):dev
TOOLS_PROD   := $(TOOLS_IMAGE):prod

TOOLS_DIR    := docker/tools

D_RUN     := docker run --rm -v $(PWD):/app -w /app --env-file .env
D_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.dev.yml --project-name $(PROJECT_NAME) --profile
RUN_APP   := $(D_RUN) --network $(APP_NETWORK) $(TOOLS_DEV)

# Build ARGs — each target passes only what its stages consume.
TOOLS_DEV_BUILD_FLAGS := \
    --file $(TOOLS_DIR)/Dockerfile \
    --build-arg GO_VERSION=$(GO_VERSION) \
    --build-arg AIR_VERSION=$(AIR_VERSION) \
    --build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
    --build-arg HUGO_VERSION=$(HUGO_VERSION) \
    --build-arg PGSCHEMA_VERSION=$(PGSCHEMA_VERSION) \
    --build-arg SQLC_VERSION=$(SQLC_VERSION) \
    --build-arg TAILWIND_VERSION=$(TAILWIND_VERSION)

TOOLS_PROD_BUILD_FLAGS := \
    --file $(TOOLS_DIR)/Dockerfile \
    --build-arg GO_VERSION=$(GO_VERSION) \
    --build-arg HUGO_VERSION=$(HUGO_VERSION) \
    --build-arg TAILWIND_VERSION=$(TAILWIND_VERSION)

DEV_BUILD_FLAGS := \
    --file docker/app/Dockerfile \
    --build-arg TOOLS_DEV_IMAGE=$(TOOLS_DEV)

PROD_BUILD_FLAGS := \
    --file docker/app/Dockerfile \
    --build-arg GO_VERSION=$(GO_VERSION) \
    --build-arg HTMX_VERSION=$(HTMX_VERSION) \
    --build-arg TOOLS_PROD_IMAGE=$(TOOLS_PROD)

# Host OS/arch for cross-compiling CLI tools inside containers.
HOST_GOOS   := $(shell uname -s | tr A-Z a-z)
HOST_GOARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

.PHONY: help \
        build clean lint vet hash-air-csp \
        db-create db-diff db-gen db-migrate db-rollback db-status \
        assets \
        package-list package-push package-release package-status package-sync \
        dev prod test \
        docker-build docker-dev docker-down docker-logs docker-up \
        dev-tools prod-tools \
        init-admin \
        _ensure-available _ensure-dev-tools _ensure-prod-tools

# ── Build & Quality ───────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: _ensure-dev-tools ## Build production binary into bin/
	$(D_RUN) $(TOOLS_DEV) sh -c 'CGO_ENABLED=0 go build -o ./bin/server ./cmd/server'

clean: ## Remove build artifacts
	rm -rf ./bin ./tmp ./public/css/app.css

hash-air-csp: _ensure-dev-tools ## Recompute CSP hash for air's proxy script and update config/app.development.yaml
	$(D_RUN) $(TOOLS_DEV) go run ./cli/util hash-air-csp

lint: _ensure-dev-tools ## Run golangci-lint
	$(D_RUN) $(TOOLS_DEV) ./scripts/workspace.sh exec golangci-lint run ./...
	$(D_RUN) $(TOOLS_DEV) ./scripts/workspace.sh exec go vet ./...

test: _ensure-dev-tools ## Run tests
	$(D_RUN) --network $(APP_NETWORK) \
	  -e TEST_DATABASE_URL=postgres://$$PGUSER:$$PGPASSWORD@$$PGHOST:$$PGPORT/$${PGDATABASE}_test?sslmode=disable \
	  -e TEST_KV_URL=redis://$$KV_HOST:$$KV_PORT/1 \
	  $(TOOLS_DEV) ./scripts/workspace.sh exec go test -v -race -count=1 ./...

# ── Database ──────────────────────────────────────────────────────────────────

db-create: _ensure-dev-tools ## Create a new empty migration file (NAME=add_posts_table)
	@test -n "$(NAME)" || { echo "error: NAME is required  e.g. make db-create NAME=add_posts_table" >&2; exit 1; }
	$(D_RUN) $(TOOLS_DEV) go run ./cli/db create "$(NAME)"

db-diff: _ensure-dev-tools ## Generate a migration file and show schema diff (NAME=add_posts_table)
	@test -n "$(NAME)" || { echo "error: NAME is required  e.g. make db-diff NAME=add_posts_table" >&2; exit 1; }
	$(D_RUN) $(TOOLS_DEV) go run ./cli/db create "$(NAME)"
	$(D_RUN) --network $(APP_NETWORK) \
	  -e PGSCHEMA_PLAN_HOST=$$PGHOST \
	  -e PGSCHEMA_PLAN_DB=$${PGDATABASE}_plan \
	  -e PGSCHEMA_PLAN_USER=$$PGUSER \
	  -e PGSCHEMA_PLAN_PASSWORD=$$PGPASSWORD \
	  $(TOOLS_DEV) pgschema plan --file db/sql/schema.sql --output-human stdout
	@echo "Review the diff above, then edit the migration file in db/migrations/"

db-gen: _ensure-dev-tools ## Regenerate sqlc Go code from SQL queries
	$(D_RUN) $(TOOLS_DEV) sqlc generate -f .sqlc.yaml

db-migrate: _ensure-dev-tools ## Apply pending migrations
	$(RUN_APP) go run ./cli/db migrate

db-rollback: _ensure-dev-tools ## Rollback the last migration
	$(RUN_APP) go run ./cli/db rollback

db-status: _ensure-dev-tools ## Show migration status
	$(RUN_APP) go run ./cli/db status


# ── Assets ────────────────────────────────────────────────────────────────────

assets: _ensure-dev-tools ## Build all generated frontend assets
	$(D_RUN) -e HTMX_VERSION=$(HTMX_VERSION) $(TOOLS_DEV) go run ./cli/build assets --minify
	$(D_RUN) $(TOOLS_DEV) go run ./cli/build docs
	$(D_RUN) $(TOOLS_DEV) go run ./cli/build sprites

# ── Package Sync & Release ────────────────────────────────────────────────────

package-list: _ensure-dev-tools ## List all discovered packages
	$(D_RUN) $(TOOLS_DEV) go run ./cli/package list

package-push: _ensure-dev-tools ## Push a package subtree to its mirror repo (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-push PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_DEV) go run ./cli/package push "$(PACKAGE)"

package-release: _ensure-dev-tools ## Release a package (PACKAGE=auth [VERSION=v0.1.0])
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-release PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_DEV) go run ./cli/package release "$(PACKAGE)" $(if $(VERSION),"$(VERSION)")

package-status: _ensure-dev-tools ## Show sync status for a package (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-status PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_DEV) go run ./cli/package status "$(PACKAGE)"

package-sync: _ensure-dev-tools ## Regenerate go.prod.mod + go.prod.sum from go.mod
	$(D_RUN) $(TOOLS_DEV) go run ./cli/package sync
	$(D_RUN) $(TOOLS_DEV) sh -c '\
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  GOWORK=off GONOSUMDB=github.com/go-sum/* GOPRIVATE=github.com/go-sum/* go mod tidy -modfile=go.prod.mod'

# ── Toolchain ────────────────────────────────────────────────────────────────

dev-tools: ## Build dev toolchain image (lint, test, db, assets)
	docker build --target dev $(TOOLS_DEV_BUILD_FLAGS) -t $(TOOLS_DEV) $(TOOLS_DIR)

prod-tools: ## Build production toolchain image (assets only)
	docker build --target prod $(TOOLS_PROD_BUILD_FLAGS) -t $(TOOLS_PROD) $(TOOLS_DIR)

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: _ensure-dev-tools ## Start all services with hot-reload
	$(D_COMPOSE) dev up --build

prod: docker-build ## Build and start the production stack locally
	docker compose up -d

docker-build: _ensure-prod-tools ## Build production Docker image
	@GITHUB_ACCESS_TOKEN="$${GITHUB_ACCESS_TOKEN:-$$(grep '^GITHUB_ACCESS_TOKEN=' .env 2>/dev/null | cut -d= -f2-)}" && \
	  export GITHUB_ACCESS_TOKEN && \
	  docker build --target production_target $(PROD_BUILD_FLAGS) \
	    --secret id=github_token,env=GITHUB_ACCESS_TOKEN \
	    -t forge:latest .

docker-dev: _ensure-dev-tools ## Build dev Docker image
	docker build --target dev_target $(DEV_BUILD_FLAGS) -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) dev down $(ARGS)

docker-logs: ## Follow container logs
	$(D_COMPOSE) dev logs -f

docker-up: ## Apply schema, then start containers in background
	$(D_COMPOSE) dev up -d $(ARGS)

_ensure-available: _ensure-dev-tools
	@docker image inspect $(APP_NAME) > /dev/null 2>&1 || \
	  docker build --target dev_target $(DEV_BUILD_FLAGS) -t $(APP_NAME) .

_ensure-dev-tools:
	@docker image inspect $(TOOLS_DEV) > /dev/null 2>&1 || \
	  docker build --target dev $(TOOLS_DEV_BUILD_FLAGS) -t $(TOOLS_DEV) $(TOOLS_DIR)
	@test -x bin/compose || { $(D_RUN) -e CGO_ENABLED=0 -e GOOS=$(HOST_GOOS) -e GOARCH=$(HOST_GOARCH) $(TOOLS_DEV) go build -o ./bin/compose ./cli/compose && chmod +x bin/compose; }

_ensure-prod-tools:
	@docker image inspect $(TOOLS_PROD) > /dev/null 2>&1 || \
	  docker build --target prod $(TOOLS_PROD_BUILD_FLAGS) -t $(TOOLS_PROD) $(TOOLS_DIR)
