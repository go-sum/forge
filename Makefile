PROJECT_NAME    ?= $(notdir $(CURDIR))
APP_NAME        := $(PROJECT_NAME)-dev
DATABASE_URL    ?= postgres://postgres:postgres@db:5432/starter?sslmode=disable
COVERAGE_FILE   ?= coverage.out
PACKAGE         ?=
VERSION         ?=

APP_NETWORK  := $(PROJECT_NAME)_app_network
TEST_NETWORK := $(PROJECT_NAME)_test_network

D_RUN     := docker run --rm -v $(PWD):/app -w /app --env-file .env
D_COMPOSE := docker compose -f docker-compose.yml -f docker-compose.dev.yml --project-name $(PROJECT_NAME) --profile
RUN_APP   := $(D_RUN) --network $(APP_NETWORK) $(APP_NAME)
RUN_TEST  := $(D_RUN) --network $(TEST_NETWORK) $(APP_NAME)

# Start $(2) via $(1) if not running, run $(3), stop any services we started.
define with-svc
@_b=$$($(1) ps --status running --services 2>/dev/null | tr '\n' ' '); $(1) ps --status running -q $(2) 2>/dev/null | grep -q . || { $(1) up -d --wait $(2) && _new=1; }; $(3); _e=$$?; [ -n "$$_new" ] && { _stop=""; for _s in $$($(1) ps --status running --services 2>/dev/null); do case " $$_b " in *" $$_s "*) ;; *) _stop="$$_stop $$_s";; esac; done; [ -n "$$_stop" ] && $(1) rm -fs $$_stop; }; exit $$_e
endef

.PHONY: help \
        build clean lint vet hash-air-csp test test-cover test-watch test-up \
        db-apply db-gen db-plan db-dump \
        assets health \
        package-sync package-release prod-sync \
        dev ddev docker-build docker-dev docker-down docker-logs docker-up \
        deploy \
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
	$(call with-svc,$(D_COMPOSE) dev,db,$(RUN_APP) go run ./cli health $(ARGS))

lint: _ensure-available ## Run golangci-lint
	$(D_RUN) $(APP_NAME) ./scripts/workspace.sh exec golangci-lint run ./...
	$(D_RUN) $(APP_NAME) ./scripts/workspace.sh exec go vet ./...

test: _ensure-available ## Run tests (auto-starts/stops test_db + test_kv)
	$(call with-svc,$(D_COMPOSE) test,test_db test_kv,$(RUN_TEST) ./scripts/workspace.sh exec go test -v -race -count=1 ./...)

test-up: ## Start the ephemeral test database and KV store manually
	$(D_COMPOSE) test up -d --wait test_db test_kv

# ── Database ──────────────────────────────────────────────────────────────────

db-apply: _ensure-available ## Apply schema.sql to the database (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema apply --file db/sql/schema.sql --auto-approve)

db-gen: _ensure-available ## Regenerate sqlc Go code from SQL queries
	$(D_RUN) $(APP_NAME) sqlc generate -f .sqlc.yaml

db-plan: _ensure-available ## Preview schema changes only (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema plan --file db/sql/schema.sql --output-human stdout)

db-dump: _ensure-available ## Dump current live database schema to stdout for preview
	$(call with-svc,$(D_COMPOSE) db,db,$(RUN_APP) pgschema dump)

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

prod-sync: ## Regenerate go.prod.mod + go.prod.sum from go.mod (requires local GitHub credentials)
	cp go.mod go.prod.mod
	GOWORK=off go mod edit \
	    -dropreplace=github.com/go-sum/auth \
	    -dropreplace=github.com/go-sum/componentry \
	    -dropreplace=github.com/go-sum/kv \
	    -dropreplace=github.com/go-sum/security \
	    -dropreplace=github.com/go-sum/send \
	    -dropreplace=github.com/go-sum/server \
	    -dropreplace=github.com/go-sum/session \
	    -dropreplace=github.com/go-sum/site \
	    -modfile=go.prod.mod
	GOWORK=off GOPRIVATE=github.com/go-sum/* go mod tidy -modfile=go.prod.mod

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: ## Start all services with hot-reload
	$(D_COMPOSE) dev up --build

deploy: ## Deploy production stack (run on server: make deploy BRANCH=main)
	./scripts/deploy.sh $(BRANCH)

docker-build: ## Build production Docker image
	docker build --target production_target -t starter:latest .

docker-dev: ## Build dev Docker image
	docker build --target dev_target -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) dev down $(ARGS)

docker-logs: ## Follow container logs
	$(D_COMPOSE) dev logs -f

docker-up: ## Apply schema, then start containers in background
	$(D_COMPOSE) dev up -d $(ARGS)

_ensure-available:
	@docker image inspect $(APP_NAME) > /dev/null 2>&1 || \
	  docker build --target dev_target -t $(APP_NAME) .
