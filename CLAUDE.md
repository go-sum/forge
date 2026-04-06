# CLAUDE.md — Architectural Constitution

> **starter** is a high-performance Go web application built on server-rendered HTML,
> progressive enhancement, and reusable packages under `pkg/`.
> Before writing code, consult: [`DESIGN_GUIDE.md`](.decisions/DESIGN_GUIDE.md) · [`UI_GUIDE.md`](.decisions/UI_GUIDE.md) · [`API_RULES.md`](.decisions/API_RULES.md)

---

## Behavioral Rules (always enforced)

- ONLY do what has been asked — recommend and get approval before any additions
- NEVER create documentation files (`*.md`) unless explicitly requested
- NEVER hardcode API keys, secrets, or credentials in source files
- NEVER commit secrets, credentials, or `.env` files
- ALWAYS validate user input at system boundaries; sanitize file paths (prevent `../` traversal)
- ALWAYS ensure implementations leverage `pkg/security` packages
- ALWAYS run tests after making code changes
- ALWAYS trace ALL callers when refactoring Go config structs or YAML mappings
- ALWAYS account for HTML-encoded entities in test assertions for HTML output
- ALWAYS enforce exact-match test assertions — never substring matching
- ALWAYS use LSP (`mcp__gomcp__lsp_*`) ahead of Grep/Glob for Go code navigation
- FALLBACK to Grep only for non-code text or when `gomcp` MCP server is unavailable

---

## MCP Server — gomcp (LSP)

Registered in `.mcp.json`. Available in all agents. **Prefer over Grep/Glob for Go.**

| Tool | Use |
|------|-----|
| `mcp__gomcp__lsp_workspace_symbols` | Find types, functions, interfaces by name |
| `mcp__gomcp__lsp_find_references` | All callers / all implementors |
| `mcp__gomcp__lsp_definition` | Jump to any symbol definition |
| `mcp__gomcp__lsp_document_symbols` | Inventory all symbols in a file |
| `mcp__gomcp__ping` | Verify server availability |

---

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language | Go | 1.26.0 |
| HTTP Framework | Echo | v5.0.4 |
| Database | PostgreSQL | 16 |
| DB Driver | pgx | v5.9.1 |
| Query Codegen | sqlc | latest |
| Migrations | goose (via `cli/db`) | latest |
| HTML Rendering | Gomponents | v1.2.0 |
| Frontend | HTMX | 2.0.4 |
| CSS | Tailwind | v4.1+ (standalone CLI, no Node.js) |
| Components | `pkg/componentry` | repo-local |

Go toolchain: `/usr/local/go/bin/go` if `go` is not in `PATH`.

---

## Layered Architecture

Data flows **down only** — each layer talks only to the one directly below it.

| Layer | Location | Responsibility |
|-------|----------|---------------|
| Transport | `internal/handler/` | Parse request · validate · call service · render |
| Service | `internal/service/` | Business rules · orchestration · domain types |
| Repository | `internal/repository/` | SQL via sqlc · map `db.*` → `model.*` |
| Database | `db/sql/` + `db/schema/` | Human SQL + generated Go (do not edit schema/) |

Domain types and error sentinels live in `internal/model/`. Handlers never import repository. Services never import handlers. `internal/` is thin wiring — general logic belongs in `pkg/`.

---

## Package Rules

**`pkg/` module-boundary leaf rule** — top-level reusable modules under `pkg/` (`auth`, `componentry`, `security`, `send`, `server`, `site`) MUST NOT import each other or `internal/`. Inside a given module family, package-local layering may be used when documented by that module.

**Component DAG** (`pkg/componentry/`) — imports flow downward only:
```
Tier 0  ui/core           (external only)
Tier 1  ui/{data,feedback,layout}, form/*, interactive/*   (+ Tier 0)
Tier 2  patterns/*        (+ Tier 0 + Tier 1)
Tier 3  examples          (any tier)
```

---

## Database Workflow

1. Edit `db/sql/schema.sql` (single source of truth for desired state)
2. Create migration: `make db-compose NAME=description` (auto-generates diff)
3. Apply migrations: `make db-migrate`
4. Regenerate Go: `make db-gen` (only if queries changed)
5. Check status: `make db-status` · Rollback: `make db-rollback`

Migrations live in `db/migrations/` and are applied via goose. Extensions (`pgcrypto`, `citext`) are managed via `db/init/01-extensions.sql`.

---

## Config Files

| File | Purpose | Required |
|------|---------|----------|
| `config/app.yaml` | Base config — no secrets | **Yes** |
| `config/app.development.yaml` | Dev overrides (Docker hostnames, debug log) | No |
| `config/site.yaml` | Site metadata, fonts, robots, sitemap | **Yes** |
| `config/nav.yaml` | Navigation structure | No |

Secrets via env vars: `encrypt_key: ${AUTH_SESSION_ENCRYPT_KEY}`

All duration-valued config fields (`expires_in`, `ttl`, `max_age`, `timeout`, etc.)
ALWAYS use **integer seconds**. NEVER use `time.Duration` string values (`"3m"`, `"1h"`) in YAML.
If required, convert to `time.Duration` at the adapter boundary: `time.Duration(n) * time.Second`.

---

## Development Phase Guide

Invoke the right agent for each phase. Each agent reads its paired rules file first.

| Phase | Agent | Rules | When |
|-------|-------|-------|------|
| Analysis & Design | `cc-plan` | `.claude/rules/r-plan.md` | Before any code — layer assignment, architecture |
| Implementation | `cc-dev` | `.claude/rules/r-code.md` | After plan approved — write code in correct layers |
| Testing | `cc-test` | `.claude/rules/r-test.md` | After implementation — happy-path + failure tests |
| Architecture Review | `cc-plan` | `.claude/rules/r-plan.md` | After tests pass — refactor planning |

**Agent flow:** `cc-plan` → `cc-dev` → `cc-test` → (if issues) back to `cc-plan`

Agents and rules live in `.claude/agents/` and `.claude/rules/`.
