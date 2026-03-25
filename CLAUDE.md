# CLAUDE.md — Architectural Guidelines & Conventions

> This is the architectural constitution for the **starter** project.
> It defines how the application works: technology choices, API rules, conventions,
> and dependency boundaries. Review before writing any code.
> The application is a high-performance modern web application starter built around
> server-rendered HTML, progressive enhancement, and reusable supporting
> packages under `pkg/`.
>
> For the application design guide, see [`DESIGN_GUIDE.md`](./.decisions/DESIGN_GUIDE.md).
> For the UI design guide, see [`UI_GUIDE.md`](./.decisions/UI_GUIDE.md).
> For the implementation roadmap and task progress, see [`PROJECT_PLAN.md`](./PROJECT_PLAN.md).

---

## Behavioral Rules (Always Enforced)
- ONLY do what's been asked 
- ALWAYS make recommendations and obtain approval for any additions or deviations first 
- NEVER proactively create documentation files (*.md) or README files unless explicitly requested
- NEVER hardcode API keys, secrets, or credentials in source files
- NEVER commit secrets, credentials, or .env files
- ALWAYS validate user input at system boundaries
- ALWAYS sanitize file paths to prevent directory traversal
- ALWAYS ensure any implementation leverages and comply with pkg/security packages
- ALWAYS run tests after making code changes
- ALWAYS trace ALL callers and update references when refactoring Go config structs or YAML mappings
- ALWAYS account for HTML-encoded entities when writing test assertions for HTML
- ALWAYS enforce exact-match test assertions over substring matching to avoid false positives.

## Project Purpose

**starter** is optimized for building modern Go web applications that keep most
of the work on the server:

- HTML is rendered on the server with Gomponents.
- HTMX handles incremental page updates without a SPA runtime.
- PostgreSQL, sqlc, and pgx keep data access explicit and type-safe.
- Reusable packages in `pkg/` are designed to support this app and to be
  extractable into other projects when their dependency rules allow it.

This repository is not a generic UI playground or a client-heavy frontend
starter. It is a performance-oriented web application foundation with reusable
supporting packages and strict application boundaries.

### `internal/` is a thin wiring layer

`internal/` is intentionally **minimal**. It exists to connect, configure, and
wire up the focused, reusable packages in `pkg/`. It is not the home for
general-purpose logic.

> **The package extraction test:** before adding any non-trivial logic to
> `internal/`, ask: *"Could this live in a `pkg/` package and be deployed
> cleanly in a new project?"* If yes, build it there first. `internal/` should
> contain only the application-specific glue that cannot be generalized —
> handler routing, request/response plumbing, DI wiring, and config loading.

Symptoms that logic belongs in `pkg/` instead of `internal/`:
- It converts between two types that both already live in `pkg/`
- It duplicates behavior that exists (or should exist) in a `pkg/` package
- It has no dependency on `internal/model`, Echo contexts, or the DB pool

### `pkg/` strict leaf-node rule

Each `pkg/` **infrastructure package** is a **leaf node** in the dependency graph:
- **MUST NOT** import from `internal/`
- **MUST NOT** import from other `pkg/` packages
- **MAY** import only from the standard library and external modules
- **COULD** be extracted into a standalone Go module without changes

---

## Technology Stack

| Layer | Technology | Version | Why |
|-------|-----------|---------|-----|
| Language | Go | 1.26.0 | Strong typing, fast compilation, single binary deploys |
| HTTP Framework | Echo | v5 | Lightweight, middleware-friendly, type-safe parameter extraction |
|||||
| Database | PostgreSQL | 16 | ACID, UUID support, `gen_random_uuid()`, triggers |
| DB Driver | pgx | v5 | Pure Go, connection pooling via `pgxpool`, context-aware |
| Query Codegen | sqlc | latest | Type-safe SQL → Go, no ORM overhead |
| Migrations | pgschema | 1.7.4 | Declarative schema-as-code, automatic diff computation |
|||||
| HTML Rendering | Gomponents | v1.2.0 | Type-safe HTML in Go, no template parsing, `io.Writer` streaming |
| HTMX | HTMX | 2.0.4 | Progressive enhancement, HTML-over-the-wire |
| CSS | Tailwind CSS | v4.1+ | Utility-first, standalone CLI (no Node.js), `@source` scanning |
| Component Library | `pkg/components` | repo-local | Reusable Go UI packages layered under a strict dependency DAG |
| Browser JS | delegated vanilla JS + esbuild | esbuild v0.25.12 | Small progressive-enhancement runtime bundled to `public/js/app.js` |
|||||
| Containerization | Docker + Compose | latest | Dev/prod parity, single-command startup, CI/CD-ready |
| Hot Reload (dev) | air | latest | Watches Go files, rebuilds and restarts inside the dev container |

---

## Echo v5 Critical API Rules

- Handlers use Echo v5's concrete pointer signature: `func(c *echo.Context) error`.
- Import paths are `github.com/labstack/echo/v5` and `github.com/labstack/echo/v5/middleware`.
- Prefer explicit parsing at the transport boundary. Use Echo's typed helpers when
  they improve clarity, or parse `c.Param(...)` / `c.QueryParam(...)` values
  directly when domain-specific conversion is clearer.
- Response helpers (`c.JSON`, `c.String`, `c.NoContent`, `c.Redirect`) and
  gomponents render helpers (`render.Component`, `render.Fragment`) are the
  normal response paths.

For the detailed v5 reference, see [`API_RULES.md`](./.decisions/API_RULES.md).

## Go Toolchain Location

If `go` is not present in `PATH`, use `/usr/local/go/bin/go`.

---

## Layered Architecture

Data flows down through four layers. Each layer only talks to the one directly below it.

1. **Transport** — HTTP concerns only. Parses requests, validates input, calls service, renders response. Knows about Echo and HTTP.
2. **Service** — Business logic. Orchestrates operations, applies business rules. Returns domain types and errors. Knows nothing about HTTP.
3. **Repository** — Data access. Wraps sqlc-generated code, maps `db` structs to domain models. Isolates schema changes from the service layer.
4. **Database** - Generated Go code and raw SQL, all under one roof. `db/schema/*.go`
   is machine-generated (do not edit). `db/sql/` is human-written (source of
   truth for schema and queries).

Handlers never import repository. Services never import handlers. Each layer communicates through domain models defined in `internal/model/`.

---

#### Component library — tiered DAG model (`pkg/components/` subtree)

`pkg/components/` packages may import each other, but only **downward** through the tier hierarchy. No lateral imports within a tier. No upward imports. No `internal/` imports.

```
Tier 0 — Primitives:   pkg/components/ui/core
                        imports: external modules only

Tier 1 — Composed:     pkg/components/ui/{data,feedback,layout}
                        pkg/components/interactive/*
                        pkg/components/form/*
                        imports: external modules + Tier 0 only

Tier 2 — Patterns:     pkg/components/patterns/*
                        imports: external modules + Tier 0 + Tier 1

Tier 3 — Examples:     pkg/components/examples
                        imports: anything within pkg/components/
```

`pkg/components/patterns/form/submission.go` accepts `*validator.Validate` directly (not `*validate.Validator`) to avoid importing `pkg/validate` (an infrastructure package, outside this tree).

### `db/` — Database Package (Root Level)

Contains both sqlc-generated Go code and human-written SQL:

- `db/schema/*.go` — **Generated by sqlc.** Do not edit by hand. The `internal/repository/` layer wraps this code and translates between `db.User` (schema-coupled) and `model.User` (domain).
- `db/sql/schema.sql` — Single source of truth for database schema
- `db/sql/queries/*.sql` — sqlc query definitions

---

## Database Workflow

### Schema Changes

1. Edit `db/sql/schema.sql` (the single source of truth)
2. Preview changes: `make db-plan` (shows human-readable diff)
3. Apply changes: `make db-apply` (computes diff and applies directly)
4. If queries changed, regenerate: `make db-gen`

**No migration files.** pgschema is purely declarative — it computes the diff between `schema.sql` and the live database on every `apply`, then executes only what changed. There are no numbered migration files to manage.

**Extensions are not managed by pgschema.** `CREATE EXTENSION` statements are outside pgschema's scope. Extensions (`pgcrypto`, `citext`) are installed via `db/init/01-extensions.sql` on first container startup (in all three database services).

**Plan Database:** pgschema uses the `schema_data` service — an ephemeral PostgreSQL container on tmpfs — for safe diff computation. Extensions are installed via `db/init-plan/01-extensions.sql` on first startup.

### sqlc Query Annotations

```sql
-- name: GetUserByID :one       → returns single struct
-- name: ListUsers :many        → returns []struct
-- name: CreateUser :exec       → returns error only
-- name: UpdateUser :one        → returns single struct (via RETURNING)
-- name: CountUsers :one        → returns single value
-- name: DeleteUser :exec       → returns error only
```

### pgschema Plan Database

pgschema uses the default `postgres` database on the `schema_data` service as a scratch database for safe diff computation (the `PGSCHEMA_PLAN_HOST` / `PGSCHEMA_PLAN_DB` env vars). The `schema_data` container runs on tmpfs — it's ephemeral, fast, and never persists data. This allows pgschema to apply the desired schema state and compute the delta without touching the main `starter` database.

The `schema_data` service is managed automatically by `make db-plan` and `make db-apply` via the `with-svc` macro, and can be safely removed at any time.

**Network Setup:** Compose creates project-scoped bridge networks for `app_network` and `test_network` (for example, `forge_app_network`). The Makefile's `D_RUN` joins the current project's app network so pgschema reaches the matching `app_data` / `schema_data` services via Docker DNS without cross-project collisions.

### Tailwind CSS Standalone CLI

No Node.js is required. Assets are built through `go run ./cli build-assets` or
`make assets`, which compile CSS into `public/css/app.css`.

The Tailwind entrypoint is `static/css/tailwind.css`. It scans:

- `internal/view/**/*.go`
- `pkg/components/**/*.go`
- `static/**/*.js`

Theme tokens and custom utility behavior live in `static/css/theme-*.css` and
`static/css/tailwind.css`. This includes dark-mode activation, HTMX indicator
styles, dialog backdrops, and the extra CSS needed by native form controls.

---

## Docker Development Environment

Both the Go server and PostgreSQL run inside Docker during development. This mirrors the production deployment (also Docker-based), ensuring dev/prod parity.

### Profile-Based Architecture

`docker-compose.yml` defines four services across three profiles:

- **`dev` profile**: `app_server` (Go app, `air` hot-reload, `:8080`) + `app_data` (PostgreSQL, persistent volume)
- **`db` profile**: `app_data` + `schema_data` (ephemeral PostgreSQL on tmpfs, for pgschema diff computation)
- **`test` profile**: `test_data` (ephemeral PostgreSQL on tmpfs, isolated on `test_network`)

The `app_server` service depends on `app_data` via health check (`service_healthy`). `DATABASE_URL` uses `app_data` as the hostname — Docker's internal DNS resolves this to the database container.

### Dockerfile — Multi-Stage Build

The project uses a single `Dockerfile` with multiple targets:

- **`dev` target**: Installs `air` for hot-reload. Source code is mounted as a Docker volume (not copied), so changes on the host are immediately visible. `air` watches for `.go` file changes, rebuilds, and restarts the server automatically.
- **`production` target**: Two-stage build — compiles a static binary with `CGO_ENABLED=0`, then copies it into a minimal Alpine image. No Go toolchain in the final image.

### Hot Reload with `air`

The `.air.toml` config file at the project root controls how `air` watches and rebuilds:

When any `.go` file changes, `air` rebuilds and restarts the server inside the container. The developer edits files on the host — the volume mount makes changes instantly visible inside the container.

### Production Build Path

The current production image is built from the same repository and asset
pipeline used in development:

```bash
# Build production image
docker build --target production -t starter:latest .
```

Release automation for image build, schema apply, and service restart is tracked
in [`PROJECT_PLAN.md`](./PROJECT_PLAN.md). The codebase already supports the
production image build itself.

---

### Config files

| File | Purpose | Required? |
|------|---------|-----------|
| `config/config.yaml` | Base config (committed, no secrets) | **Required** |
| `config/config.development.yaml` | Dev overrides (Docker hostnames, debug logging) | Optional |
| `config/site.yaml` | Site metadata: title, description, keywords, OG image, fonts, robots, sitemap | **Required** |
| `config/nav.yaml` | Navigation structure | Optional |

Only `config/config.yaml` and `config/site.yaml` is required. All other files are silently skipped when absent.

Env vars can be used to inject into the existing schema:
  encrypt_key: ${AUTH_SESSION_ENCRYPT_KEY}

### Secrets

Secrets belong in environment variables, never committed YAML files.

---

## Request Flow

```
Browser → GET /users
  - Echo Router
  - Middleware: Recover → RequestID → Secure → RequestLogger ...
  - Handler
  - Service
  - Repository
     - Wraps sqlc queries, maps db → model
     - Executes SQL against pgxpool
      - PostgreSQL returns rows
  - Handler decides rendering path:
     HTMX-capable page → view.Render(full, partial)
     HTMX-only endpoint → render.Fragment
  - Browser receives HTML
```
