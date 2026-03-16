ALPINE_VERSION := 3.14.8
HTMX_VERSION   := 2.0.4
APP_NAME       := starter-dev
DATABASE_URL   ?= postgres://postgres:postgres@app_data:5432/starter?sslmode=disable

APP_NETWORK  := app_network
TEST_NETWORK := test_network

D_RUN  := docker run --rm -v $(PWD):/app -w /app --env-file .env
D_COMPOSE := docker compose --profile

RUN_APP  := $(D_RUN) --network $(APP_NETWORK) $(APP_NAME)
RUN_TEST := $(D_RUN) --network $(TEST_NETWORK) $(APP_NAME)

# Start $(2) via $(1) if not running, run $(3), stop any services we started.
define with-svc
@_b=$$($(1) ps --status running --services 2>/dev/null | tr '\n' ' '); $(1) ps --status running -q $(2) 2>/dev/null | grep -q . || { $(1) up -d --wait $(2) && _new=1; }; $(3); _e=$$?; [ -n "$$_new" ] && { _stop=""; for _s in $$($(1) ps --status running --services 2>/dev/null); do case " $$_b " in *" $$_s "*) ;; *) _stop="$$_stop $$_s";; esac; done; [ -n "$$_stop" ] && $(1) rm -fs $$_stop; }; exit $$_e
endef

.PHONY: help \
        build clean lint test test-watch test-up \
        db-apply db-gen db-plan db-dump \
        css js \
        dev docker-build docker-dev docker-down docker-logs docker-up \
        _ensure-dev-image

# ── Build & Quality ───────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: _ensure-dev-image ## Build production binary
	$(D_RUN) $(APP_NAME) sh -c 'CGO_ENABLED=0 go build -o ./bin/server ./cmd/server'

clean: ## Remove build artifacts
	rm -rf ./bin/server ./tmp ./public/css/app.css

lint: _ensure-dev-image ## Run golangci-lint
	$(D_RUN) $(APP_NAME) golangci-lint run ./...

test: _ensure-dev-image ## Run tests (auto-starts/stops test_data)
	$(call with-svc,$(D_COMPOSE) test,test_data,$(RUN_TEST) go test -v -race -count=1 ./...)

test-watch: _ensure-dev-image ## Run tests with hot-reload (auto-starts test_data)
	$(call with-svc,$(D_COMPOSE) test,test_data,$(RUN_TEST) air -c .air.test.toml)

test-up: ## Start the ephemeral test database manually
	$(D_COMPOSE) test up -d --wait test_data

# ── Database ──────────────────────────────────────────────────────────────────

db-apply: _ensure-dev-image ## Apply schema.sql to the database (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema apply --file db/sql/schema.sql --auto-approve)

db-gen: _ensure-dev-image ## Regenerate sqlc Go code from SQL queries
	$(D_RUN) $(APP_NAME) sqlc generate -f .sqlc.yaml

db-plan: _ensure-dev-image ## Preview schema changes (auto-starts/stops schema_data)
	$(call with-svc,$(D_COMPOSE) db,schema_data,$(RUN_APP) pgschema plan --file db/sql/schema.sql --output-human stdout)

db-dump: _ensure-dev-image ## Dump current live database schema to stdout for preview
	$(call with-svc,$(D_COMPOSE) db,app_data,$(RUN_APP) pgschema dump)

# ── Assets ────────────────────────────────────────────────────────────────────

css: _ensure-dev-image ## Compile Tailwind CSS
	$(D_RUN) $(APP_NAME) tailwindcss -i ./static/css/tailwind.css -o ./public/css/app.css --minify

js: ## Download HTMX and Alpine.js
	mkdir -p public/js
	curl -fsSL "https://unpkg.com/htmx.org@$(HTMX_VERSION)/dist/htmx.min.js" -o public/js/htmx.min.js
	curl -fsSL "https://unpkg.com/@alpinejs/csp@$(ALPINE_VERSION)/dist/cdn.min.js" -o public/js/alpine.min.js

# ── Docker & Dev ──────────────────────────────────────────────────────────────

dev: ## Start all services with hot-reload
	$(D_COMPOSE) dev up

docker-build: ## Build production Docker image
	docker build --target production -t starter:latest .

docker-dev: ## Build dev Docker image
	docker build --target dev -t $(APP_NAME) .

docker-down: ## Stop and remove containers
	$(D_COMPOSE) dev down

docker-logs: ## Follow container logs
	$(D_COMPOSE) dev logs -f

docker-up: ## Start containers in background
	$(D_COMPOSE) dev up -d

_ensure-dev-image:
	@docker image inspect $(APP_NAME) > /dev/null 2>&1 || \
	  docker build --target dev -t $(APP_NAME) .
