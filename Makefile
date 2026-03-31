include .versions
export

PROJECT_NAME    ?= $(notdir $(CURDIR))
APP_NAME        := $(PROJECT_NAME)-dev
TOOLS_NAME      := $(PROJECT_NAME)-tools
DATABASE_URL    ?= postgres://postgres:postgres@db:5432/starter?sslmode=disable
COVERAGE_FILE   ?= coverage.out
PACKAGE         ?=
VERSION         ?=

APP_NETWORK  := $(PROJECT_NAME)_app_network
TEST_NETWORK := $(PROJECT_NAME)_test_network

D_RUN     := docker run --rm -v $(PWD):/app -w /app --env-file .env
D_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.dev.yml --project-name $(PROJECT_NAME) --profile
RUN_APP   := $(D_RUN) --network $(APP_NETWORK) $(TOOLS_NAME)
RUN_TEST  := $(D_RUN) --network $(TEST_NETWORK) $(TOOLS_NAME)

# Build ARGs — Dockerfile stages only use what it declares.
BUILD_FLAGS := \
    --file docker/app/Dockerfile \
    --build-arg GO_VERSION=$(GO_VERSION) \
    --build-arg AIR_VERSION=$(AIR_VERSION) \
    --build-arg TAILWIND_VERSION=$(TAILWIND_VERSION) \
    --build-arg SQLC_VERSION=$(SQLC_VERSION) \
    --build-arg PGSCHEMA_VERSION=$(PGSCHEMA_VERSION) \
    --build-arg HUGO_VERSION=$(HUGO_VERSION) \
    --build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION)

# Start $(2) via $(1) if not running, run $(3), stop any services we started.
define with-svc
@_b=$$($(1) ps --status running --services 2>/dev/null | tr '\n' ' '); $(1) ps --status running -q $(2) 2>/dev/null | grep -q . || { $(1) up -d --wait $(2) && _new=1; }; $(3); _e=$$?; [ -n "$$_new" ] && { _stop=""; for _s in $$($(1) ps --status running --services 2>/dev/null); do case " $$_b " in *" $$_s "*) ;; *) _stop="$$_stop $$_s";; esac; done; [ -n "$$_stop" ] && $(1) rm -fs $$_stop; }; exit $$_e
endef

.PHONY: help \
        build clean lint vet hash-air-csp health\
        db-apply db-gen db-plan db-dump \
        assets \
        package-list package-push package-release package-status package-sync \
        dev prod test \
        docker-build docker-dev docker-down docker-logs docker-up \
        _ensure-available _ensure-tools

# ── Build & Quality ───────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: _ensure-tools ## Build production binary into bin/
	$(D_RUN) $(TOOLS_NAME) sh -c 'CGO_ENABLED=0 go build -o ./bin/server ./cmd/server'

clean: ## Remove build artifacts
	rm -rf ./bin ./tmp ./public/css/app.css

hash-air-csp: _ensure-tools ## Recompute CSP hash for air's proxy script and update config/app.development.yaml
	$(D_RUN) $(TOOLS_NAME) go run ./cli/util hash-air-csp

health: _ensure-tools ## Run health checks (use ARGS='--verbose' or '--json')
	$(call with-svc,$(D_COMPOSE) dev,db,$(RUN_APP) go run ./cli/util health $(ARGS))

lint: _ensure-tools ## Run golangci-lint
	$(D_RUN) $(TOOLS_NAME) ./scripts/workspace.sh exec golangci-lint run ./...
	$(D_RUN) $(TOOLS_NAME) ./scripts/workspace.sh exec go vet ./...

test: _ensure-tools ## Run tests (auto-starts/stops test_db + test_kv)
	$(call with-svc,$(D_COMPOSE) test,test_db test_kv,$(RUN_TEST) ./scripts/workspace.sh exec go test -v -race -count=1 ./...)

# ── Database ──────────────────────────────────────────────────────────────────

db-apply: _ensure-tools ## Apply schema.sql to the database (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema apply --file db/sql/schema.sql --auto-approve)

db-gen: _ensure-tools ## Regenerate sqlc Go code from SQL queries
	$(D_RUN) $(TOOLS_NAME) sqlc generate -f .sqlc.yaml

db-plan: _ensure-tools ## Preview schema changes only (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema plan --file db/sql/schema.sql --output-human stdout)

db-dump: _ensure-tools ## Dump current live database schema to stdout for preview
	$(call with-svc,$(D_COMPOSE) db,db,$(RUN_APP) pgschema dump)

# ── Assets ────────────────────────────────────────────────────────────────────

assets: _ensure-tools ## Build all generated frontend assets
	$(D_RUN) -e HTMX_VERSION=$(HTMX_VERSION) $(TOOLS_NAME) go run ./cli/build assets --minify

# ── Package Sync & Release ────────────────────────────────────────────────────

package-list: _ensure-tools ## List all discovered packages
	$(D_RUN) $(TOOLS_NAME) go run ./cli/package list

package-push: _ensure-tools ## Push a package subtree to its mirror repo (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-push PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_NAME) go run ./cli/package push "$(PACKAGE)"

package-release: _ensure-tools ## Release a package (PACKAGE=auth [VERSION=v0.1.0])
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-release PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_NAME) go run ./cli/package release "$(PACKAGE)" $(if $(VERSION),"$(VERSION)")

package-status: _ensure-tools ## Show sync status for a package (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-status PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(TOOLS_NAME) go run ./cli/package status "$(PACKAGE)"

package-sync: _ensure-tools ## Regenerate go.prod.mod + go.prod.sum from go.mod
	$(D_RUN) $(TOOLS_NAME) go run ./cli/package sync
	$(D_RUN) $(TOOLS_NAME) sh -c '\
	  git config --global url."https://x-access-token:$${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
	  GOWORK=off GONOSUMDB=github.com/go-sum/* GOPRIVATE=github.com/go-sum/* go mod tidy -modfile=go.prod.mod'

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: ## Start all services with hot-reload
	$(D_COMPOSE) dev up --build

prod: docker-build ## Build and start the production stack locally
	docker compose up -d

docker-build: ## Build production Docker image
	@GITHUB_ACCESS_TOKEN="$${GITHUB_ACCESS_TOKEN:-$$(grep '^GITHUB_ACCESS_TOKEN=' .env 2>/dev/null | cut -d= -f2-)}" && \
	  export GITHUB_ACCESS_TOKEN && \
	  docker build --target production_target $(BUILD_FLAGS) \
	    --build-arg HTMX_VERSION=$(HTMX_VERSION) \
	    --secret id=github_token,env=GITHUB_ACCESS_TOKEN \
	    -t forge:latest .

docker-dev: ## Build dev Docker images (wolfi app + bookworm toolchain)
	docker build --target dev_target $(BUILD_FLAGS) -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) dev down $(ARGS)

docker-logs: ## Follow container logs
	$(D_COMPOSE) dev logs -f

docker-up: ## Apply schema, then start containers in background
	$(D_COMPOSE) dev up -d $(ARGS)

_ensure-available:
	@docker image inspect $(APP_NAME) > /dev/null 2>&1 || \
	  docker build --target dev_target $(BUILD_FLAGS) -t $(APP_NAME) .

_ensure-tools:
	@docker image inspect $(TOOLS_NAME) > /dev/null 2>&1 || \
	  docker build --target cli_toolchain $(BUILD_FLAGS) -t $(TOOLS_NAME) .
