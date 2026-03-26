PROJECT_NAME    ?= $(notdir $(CURDIR))
APP_NAME        := $(PROJECT_NAME)-dev
DATABASE_URL    ?= postgres://postgres:postgres@app_data:5432/starter?sslmode=disable
COVERAGE_FILE   ?= coverage.out
PACKAGE         ?=
VERSION         ?=

APP_NETWORK  := $(PROJECT_NAME)_app_network
TEST_NETWORK := $(PROJECT_NAME)_test_network

D_RUN     := docker run --rm -v $(PWD):/app -w /app --env-file .env
D_COMPOSE := docker compose --project-name $(PROJECT_NAME) --profile
RUN_APP   := $(D_RUN) --network $(APP_NETWORK) $(APP_NAME)
RUN_TEST  := $(D_RUN) --network $(TEST_NETWORK) $(APP_NAME)

# Start $(2) via $(1) if not running, run $(3), stop any services we started.
define with-svc
@_b=$$($(1) ps --status running --services 2>/dev/null | tr '\n' ' '); $(1) ps --status running -q $(2) 2>/dev/null | grep -q . || { $(1) up -d --wait $(2) && _new=1; }; $(3); _e=$$?; [ -n "$$_new" ] && { _stop=""; for _s in $$($(1) ps --status running --services 2>/dev/null); do case " $$_b " in *" $$_s "*) ;; *) _stop="$$_stop $$_s";; esac; done; [ -n "$$_stop" ] && $(1) rm -fs $$_stop; }; exit $$_e
endef

.PHONY: help \
        build clean lint hash-air-csp test test-cover test-watch test-up \
        db-apply db-gen db-plan db-dump \
        assets health \
        package-sync package-release \
        dev ddev docker-build docker-dev docker-down docker-logs docker-up \
        _ensure-available

# ── Build & Quality ───────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: _ensure-available ## Build production binary
	$(D_RUN) $(APP_NAME) sh -c 'CGO_ENABLED=0 go build -o ./bin/server ./cmd/server'

clean: ## Remove build artifacts
	rm -rf ./bin/server ./tmp ./public/css/app.css

hash-air-csp: _ensure-available ## Recompute CSP hash for air's proxy script and update config/app.development.yaml
	$(D_RUN) $(APP_NAME) go run ./cli hash-air-csp

health: _ensure-available ## Run health checks (use ARGS='--verbose' or '--json')
	$(call with-svc,$(D_COMPOSE) dev,app_data,$(RUN_APP) go run ./cli health $(ARGS))

lint: _ensure-available ## Run golangci-lint
	$(D_RUN) $(APP_NAME) golangci-lint run ./...

test: _ensure-available ## Run tests (auto-starts/stops test_data)
	$(call with-svc,$(D_COMPOSE) test,test_data,$(RUN_TEST) go test -v -race -count=1 ./...)

test-cover: _ensure-available ## Run tests with coverage and print the summary
	$(call with-svc,$(D_COMPOSE) test,test_data,$(RUN_TEST) sh -c 'go test -coverpkg=./... -coverprofile=$(COVERAGE_FILE) ./... && go tool cover -func=$(COVERAGE_FILE)')

test-watch: _ensure-available ## Run tests with hot-reload (auto-starts test_data)
	$(call with-svc,$(D_COMPOSE) test,test_data,$(RUN_TEST) air -c .air.test.toml)

test-up: ## Start the ephemeral test database manually
	$(D_COMPOSE) test up -d --wait test_data

# ── Database ──────────────────────────────────────────────────────────────────

db-apply: _ensure-available ## Apply schema.sql to the database (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema apply --file db/sql/schema.sql --auto-approve)

db-gen: _ensure-available ## Regenerate sqlc Go code from SQL queries
	$(D_RUN) $(APP_NAME) sqlc generate -f .sqlc.yaml

db-plan: _ensure-available ## Preview schema changes only (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema plan --file db/sql/schema.sql --output-human stdout)

db-dump: _ensure-available ## Dump current live database schema to stdout for preview
	$(call with-svc,$(D_COMPOSE) db,app_data,$(RUN_APP) pgschema dump)

# ── Assets ────────────────────────────────────────────────────────────────────

assets: _ensure-available ## Build all generated frontend assets
	$(D_RUN) -e HTMX_VERSION=$(HTMX_VERSION) $(APP_NAME) go run ./cli build-assets --minify

# ── Package Sync & Release ────────────────────────────────────────────────────

package-sync: _ensure-available ## Sync a package subtree to its mirror repo (PACKAGE=auth)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-sync PACKAGE=auth" >&2; exit 1; }
	$(D_RUN) $(APP_NAME) ./scripts/package-sync.sh "$(PACKAGE)"

package-release: _ensure-available ## Release a versioned package to its mirror repo (PACKAGE=auth VERSION=v0.1.0)
	@test -n "$(PACKAGE)" || { echo "error: PACKAGE is required  e.g. make package-release PACKAGE=auth VERSION=v0.1.0" >&2; exit 1; }
	@test -n "$(VERSION)" || { echo "error: VERSION is required  e.g. make package-release PACKAGE=auth VERSION=v0.1.0" >&2; exit 1; }
	$(D_RUN) $(APP_NAME) ./scripts/package-release.sh "$(PACKAGE)" "$(VERSION)"

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: ## Start all services with hot-reload
	$(D_COMPOSE) dev up --build

docker-build: ## Build production Docker image
	docker build --target production -t starter:latest .

docker-dev: ## Build dev Docker image
	docker build --target dev -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) dev down $(ARGS)

docker-logs: ## Follow container logs
	$(D_COMPOSE) dev logs -f

docker-up: ## Apply schema, then start containers in background
	$(D_COMPOSE) dev up -d $(ARGS)

_ensure-available:
	@docker image inspect $(APP_NAME) > /dev/null 2>&1 || \
	  docker build --target dev -t $(APP_NAME) .
