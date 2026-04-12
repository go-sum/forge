# MONOREPO_PACKAGES.md — Package Development & Monorepo Rules

> **Additive overlay for agents working in this monorepo.**
>
> This file supplements the starter-zone rules (DESIGN_GUIDE, PATTERNS_PRINCIPLES,
> r-plan.md, r-code.md, r-test.md). Load it when the repo contains a `pkg/`
> directory. It does not apply to app-only clones.

---

## Package Ownership Model

This repo is a monorepo: the nine `pkg/*` modules are separate Go modules
developed here and published as independent `github.com/go-sum/*` packages.
The root `internal/` application consumes them as first-class external
dependencies (wired via `go.work` during development; `go.prod.mod` for
production releases).

### Two Design Zones

| Zone | Location | What lives here |
|---|---|---|
| App-owned | `internal/` | Product logic, page composition, app-specific orchestration |
| Package-owned | `pkg/<name>/` | Reusable capabilities deployable independently |

---

## Layer Assignment: Package-owned code (`pkg/`)

| Concern | Location |
|---------|----------|
| Package domain types and error sentinels | `pkg/<name>/model/` |
| Package service logic | `pkg/<name>/service/` |
| Package repository interfaces | `pkg/<name>/repository/interfaces.go` |
| Package pgstore implementation | `pkg/<name>/pgstore/pgstore.go` |
| Package SQL schema and queries | `pkg/<name>/pgstore/sql/schema.sql`, `queries.sql` |
| Package sqlc-generated code (never edit) | `pkg/<name>/pgstore/db/` |
| Package HTTP handlers | `pkg/<name>/` (when the pkg intentionally exposes HTTP) |
| Shared UI components | `pkg/componentry/` |

**Strict rule**: `pkg/` packages are separate Go modules — each has its own
`go.mod`. They MUST NOT import `internal/` or other `pkg/` packages.
Cross-boundary imports fail at compile time.

---

## Feature Development: Step 0 (Ownership)

Before step 1 of the canonical sequence, answer:

- Could this deploy cleanly in a different app with no `internal/` dependency? → **pkg-owned**
- Does it define a reusable schema contract? → **pkg-owned**
- Is it specific to this app's product logic or page composition? → **app-owned**
- Default to **app-owned** if still unclear; extract later if real reuse demand appears.

The ownership answer determines which location every subsequent step uses:
- Model: `pkg/<name>/model/` vs `internal/model/`
- SQL: `pkg/<name>/pgstore/sql/` vs `db/sql/`
- Repository: `pkg/<name>/pgstore/pgstore.go` vs `internal/repository/`
- Service: `pkg/<name>/service/` vs `internal/features/<name>/service.go`
- Handler: `pkg/<name>/` vs `internal/features/<name>/handler.go`
- Tests: co-located in `pkg/<name>/...` vs `internal/features/<name>/`

Views always live in `internal/view/` regardless of ownership.

---

## Package Extraction Test

Before adding logic to `internal/`, ask: *"Could this live in `pkg/` and
deploy cleanly in a new project?"*

**It belongs in `pkg/` if:**
- It converts between two types that both already live in `pkg/`
- It duplicates behavior that exists (or should exist) in a `pkg/` package
- It has no dependency on `internal/model`, Echo contexts, or the DB pool

**`pkg/` leaf-node rule** — each `pkg/` package MUST NOT:
- Import from `internal/`
- Import from other `pkg/` packages
- Reference application-specific types

This is enforced architecturally: each `pkg/*` has its own `go.mod`
(separate Go module). Cross-boundary imports are caught at compile time.

---

## sqlc Ownership Table (extended)

| Domain ownership | Query source | Generated code |
|---|---|---|
| App-owned tables | `db/sql/queries/*.sql` | root sqlc output |
| `pkg/auth` tables | `pkg/auth/pgstore/sql/queries.sql` | `pkg/auth/pgstore/db/` |
| `pkg/queue` tables | `pkg/queue/pgstore/sql/queries.sql` | `pkg/queue/pgstore/db/` |

If a query is missing in a `pkg/` module, add it to the owning `.sql` file
and run `task ws:db:gen`. Never hand-edit files in `*/pgstore/db/`.

---

## Component DAG (`pkg/componentry/`) — In-repo Paths

When working directly on `pkg/componentry` source files, use these in-repo
import paths (the workspace `replace` directives map them automatically):

```
Tier 0  pkg/componentry/ui/core          — external modules only
Tier 1  pkg/componentry/ui/{data,feedback,layout}
        pkg/componentry/form/*
        pkg/componentry/interactive/*    — external + Tier 0
Tier 2  pkg/componentry/patterns/*       — external + Tier 0 + Tier 1
Tier 3  pkg/componentry/examples         — any tier within pkg/componentry/
```

In `internal/` and in app clones, use the published import path
`github.com/go-sum/componentry/...` — the workspace handles the redirect.

---

## Canonical Locations: Package-owned rows

| Concern | Location |
|---------|----------|
| Auth domain (schema, queries, generated, service, handler) | `pkg/auth/pgstore/sql/`, `pkg/auth/pgstore/db/`, `pkg/auth/service/` |
| Queue domain (schema, queries, generated) | `pkg/queue/pgstore/sql/`, `pkg/queue/pgstore/db/` |
| Shared UI components | `pkg/componentry/` |

---

## Monorepo CLI: `tools/cli/`

Tooling that operates on `pkg/*`, `go.work`, or mirror repos lives in
`tools/cli/` (excluded from app clones):

| Binary | Purpose |
|---|---|
| `tools/cli/package` | Subtree-split, push, release, status, sync |
| `tools/cli/workspace` | Fan-out a command across all `go.work` modules |
| `tools/cli/starter` | Clone forge into a new app (exclude + rename) |

Run with `go run ./tools/cli/<name> <subcommand>` or via `task ws:*`
targets defined in `tools/Taskfile.yml`.

---

## Planning Rules: Package-aware Checklist

Pre-planning checklist additions when `pkg/` is present:

1. Which layer? Transport / Service / Repository / Database / **pkg/**
2. Apply the Package Extraction Test before assigning to `internal/`.
3. If adding to `pkg/`, confirm the leaf-node rule holds (no `internal/`
   or cross-`pkg/` imports).
4. `db:gen` requires `task ws:db:gen` — plan which `.sqlc.yaml` configs are
   affected (root and/or `pkg/*/pgstore/.sqlc.yaml`).

---

## Echo v5 API Rules

See [`ECHOV5_API_REFACTOR.md`](./ECHOV5_API_REFACTOR.md) for the full
transport-layer reference: handler signatures, breaking changes from v4,
type-safe parameter extraction, binding, response methods, and removed APIs.
