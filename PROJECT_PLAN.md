# PROJECT_PLAN.md — Implementation Roadmap & Progress Tracker

> This is the living roadmap for the **starter** project.
> It tracks what to build, what's done, and what's next.
>
> For architectural guidelines and conventions, see [`CLAUDE.md`](./CLAUDE.md).

---

## Context

**starter** is a high-performance modern Go web application starter. It is
optimized for server-rendered HTML, progressive enhancement with HTMX, explicit
data access, and reusable supporting packages that can outlive this specific
application.

The architecture follows a strict layered pattern:
**Transport → Service → Repository → Database**.

- `internal/` contains application-specific code.
- `pkg/` contains reusable supporting packages with strict dependency rules.
- `db/` contains schema SQL, query SQL, and sqlc-generated database bindings.

---

## Key Design Decisions

| Decision | Source |
|----------|--------|
| Echo v5 handler signatures (`*echo.Context`) | `.decisions/API_RULES.md` |
| Layered architecture: Transport → Service → Repository → Database | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| sqlc + pgx + PostgreSQL for explicit, type-safe data access | `CLAUDE.md` |
| pgschema for declarative schema apply (no numbered migration files) | `CLAUDE.md` |
| Koanf-based layered config with `CTX_` env overrides | `config/config.go`, `pkg/config` |
| Gomponents for type-safe HTML rendering | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| HTMX for partial-page interactions | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| Reusable supporting packages under `pkg/` | `CLAUDE.md`, `pkg/components/INSTALL.md` |

---

## TASKS

Status markers: `[ ]` outstanding · `[x]` completed · `[~]` in progress

Task ID format: `TXXYY` where `XX` = phase, `YY` = task number within that phase.
This allows inserting new tasks within any phase without renumbering downstream task numbers.

---

### Phase 01:

- [ ] **T0101** — Wire an explicit admin route group into the application router
  - Cover: add an `admin` group in `internal/server/routes.go`, apply `RequireAdmin`, move admin-only surfaces under that group, and add route/middleware coverage proving public vs protected vs admin behavior.

- [ ] **T0102** — Automate the production deployment path described by the docs
  - Cover: script or pipeline the production image build, `pgschema apply`, and service restart flow so the documented production path is executable rather than manual.
