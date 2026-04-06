# PROJECT_PLAN.md — Implementation Roadmap & Progress Tracker

> This is the living roadmap for the **starter** project.
> It tracks what to build, what is already aligned, and what should be adopted next.
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

This roadmap now incorporates a review of [`../examples/r3-cf/`](../examples/r3-cf/),
but only for capabilities that transfer cleanly into this Go/Echo/HTMX monorepo.

---

## Repository Topology

The local monorepo is the canonical source of truth for the application and the
extractable packages:

- Monorepo: `https://github.com/go-sum/forge`
- Auth package: `https://github.com/go-sum/auth`
- Component library: `https://github.com/go-sum/componentry`
- Server package: `https://github.com/go-sum/server`
- Security package: `https://github.com/go-sum/security`
- Site package: `https://github.com/go-sum/side`

Development happens in the monorepo with `go.work`. Package repositories are
synced from `pkg/auth`, `pkg/componentry`, `pkg/server`, `pkg/security`, and `pkg/site`
through the GitHub Actions workflows under `.github/workflows/`.

---

## Key Design Decisions

| Decision | Source |
|----------|--------|
| Echo v5 handler signatures (`*echo.Context`) | `.decisions/API_RULES.md` |
| Layered architecture: Transport → Service → Repository → Database | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| sqlc + pgx + PostgreSQL for explicit, type-safe data access | `CLAUDE.md` |
| goose-based migrations in `db/migrations/` (pgschema used for diff generation) | `CLAUDE.md` |
| Koanf-based layered config with `CTX_` env overrides | `config/config.go` |
| Gomponents for type-safe HTML rendering | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| HTMX for partial-page interactions | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| Reusable supporting packages under `pkg/` | `CLAUDE.md`, `go.work` |
| Tailwind standalone CLI + Go-driven asset pipeline | `CLAUDE.md`, `cli/build-assets.go` |

---

## Verification Expectations

Every phase should include the appropriate level of verification:

- Unit tests for all new package APIs.
- Route and middleware tests for application wiring.
- CLI tests for build-time tooling such as docs generation.
- Config parsing tests for new YAML/env-driven capabilities.
- At least one end-to-end integration path for each new subsystem once it is
  wired into the starter.

---

## Current Defaults and Assumptions

- Reusable package design takes priority over app-only implementation shortcuts.
- The starter remains server-rendered and HTMX-first; no SPA runtime is planned.
- The existing asset pipeline remains Go-driven and Node-free.

---

## TASKS

Status markers: `[ ]` outstanding · `[x]` completed · `[~]` in progress

Task ID format: `TXXYY` where `XX` = phase, `YY` = task number within that
phase. This allows inserting new tasks within any phase without renumbering
downstream task numbers.

---

### Phase 00: Bugfixes

- [x] **T0001** — Fix session fixation: rotate session ID after authentication
  Severity: **High**
  Cover: successful authentication does not rotate the session ID, leaving the
  app vulnerable to session fixation. In `internal/adapters/authui/handler.go:253`
  the Verify flow loads the existing session, writes auth state, and commits it
  unchanged; `pkg/session/manager.go:123` already exposes `RotateID` but it is
  never called anywhere in the codebase. Replace `h.sessions.Commit(...)` with
  `h.sessions.RotateID(...)` at the Verify handler's authentication commit point
  (line 260). Audit the other session commit sites in the same file — Signin,
  Signup, ResendVerify, and BeginEmailChange write pending flow state (not auth),
  so rotation is only mandatory at the Verify step where the user transitions
  from unauthenticated to authenticated. Add a test verifying the session ID
  changes after successful verification.
  Files: `internal/adapters/authui/handler.go`, auth handler test file

- [x] **T0002** — Harden security headers: add missing CSP directives, Referrer-Policy, Permissions-Policy
  Severity: **High**
  Cover: the security-header baseline has gaps. `config/app.yaml:89` CSP is
  missing `object-src 'none'` and `base-uri 'self'`. There is no
  `Referrer-Policy` or `Permissions-Policy` header anywhere in the codebase, and
  `config/app.go:96` `HeadersConfig` has no fields for them. Echo's
  `SecureWithConfig` (used in `internal/server/security.go:85`) does not support
  these two headers natively. Fix: (1) append `object-src 'none'; base-uri
  'self'` to the CSP string in `config/app.yaml`; (2) add `referrer_policy` and
  `permissions_policy` fields to `HeadersConfig` with sensible defaults
  (`strict-origin-when-cross-origin` and `camera=(), microphone=(),
  geolocation=()`); (3) add a custom middleware in `internal/server/security.go`
  that sets these two response headers; (4) wire it into the middleware chain in
  `internal/server/middleware.go`. Add tests for the new header middleware.
  Files: `config/app.go`, `config/app.yaml`, `internal/server/security.go`,
  `internal/server/middleware.go`

- [x] **T0003** — Replace hardcoded route paths with named route reversal
  Severity: **Medium**
  Cover: two locations bypass the named-route single-source-of-truth.
  `internal/view/layout/base.go:113` hardcodes `"/signout"` instead of reversing
  `"signout.post"`, and `internal/handler/admin.go:55` hardcodes `"/"` instead of
  reversing `"home.show"`. The route reversal infrastructure already exists:
  `route.Reverse()` in `pkg/server/route/route.go:23` and the view layer exposes
  `Request.Path()` at `internal/view/request.go:81`. Both route names are
  registered in `internal/app/routes.go`. Fix: replace hardcoded strings with
  reversed named routes. Grep for any other hardcoded route paths in
  `internal/handler/` and `internal/view/` and fix those too.
  Files: `internal/view/layout/base.go`, `internal/handler/admin.go`

- [x] **T0004** — Rate limiter: log warning on unsupported backend instead of silent fallback
  Severity: **Medium**
  Cover: `internal/server/ratelimit.go:60` silently falls back to in-memory
  storage for any non-memory backend value — no log, no error, no indication to
  the operator. In multi-instance deployments the single-process, IP-only rate
  limiter becomes easy to evade. Fix: add `slog.Warn` in the `default` switch
  case of `newRateLimitStore` so operators are alerted that their configured
  backend is unsupported and memory fallback is in use. Add a code comment
  documenting the single-process limitation for future distributed-store work.
  Files: `internal/server/ratelimit.go`

### Phase 01: Echo v5 Runtime Hardening

Low-effort, high-value fixes to the existing runtime configuration.

- [ ] **T0101** — Configure `Echo.IPExtractor` for deployment model
  Cover: select the correct Echo v5 extractor (`ExtractIPDirect` for direct
  exposure, `ExtractIPFromXFFHeader` with trust options for reverse-proxy
  deployments) so `c.RealIP()` is trustworthy for rate limiting and request
  logging. Add a config key for deployment topology.
  Files: `internal/app/bootstrap.go`, `config/app.yaml`

- [ ] **T0102** — Enable Echo v5 `RequestLoggerWithConfig` field capture flags
  Cover: enable the `Log*` flags in `internal/server/middleware.go` (currently
  commented out at lines 39-45) so Echo v5 explicitly captures the fields the
  custom `LogValuesFunc` reads. Currently works by convention — make it
  intentional. Move logging policy into `pkg/server/logging` if reusable.
  Files: `internal/server/middleware.go`, optionally `pkg/server/logging/`

---

### Phase 02: Locale-Aware Formatting

Build on the existing `ParseAcceptLanguage()` in `pkg/server/headers/accept.go`.
Scope is formatting primitives only — no translation catalogs or message bundles.

- [ ] **T0201** — Add locale negotiation middleware
  Cover: create middleware that reads `Accept-Language` (using the existing
  parser), checks an optional cookie override, resolves to a preferred locale,
  and sets it on the request context. Provide configurable fallback. Package as
  reusable infrastructure under `pkg/`.
  Files: new `pkg/` locale package, middleware integration in `internal/server/`

- [ ] **T0202** — Add reusable number and time formatting helpers
  Cover: support number, currency, date, datetime, and relative-time formatting
  for a set of locales. Helpers must be usable from Gomponents views without
  coupling to HTTP concerns. Accept `time.Time` and locale identifier as inputs.
  Files: new formatting helpers in `pkg/` locale package

- [ ] **T0203** — Thread locale and timezone through the request/view layer
  Cover: expose the negotiated locale at render time via view request context so
  Gomponents views can format dates and numbers without importing HTTP packages.
  Add timezone awareness (derive from user preference or request header).
  Files: `internal/view/request.go`, view partials that render dates

---

### Phase 03: Account Lifecycle Expansion

Build on the existing email-TOTP auth system and email-change flow.

- [ ] **T0301** — Add self-service account settings page
  Cover: create an authenticated `/account` or `/account/settings` page where
  users can view their profile info, edit display name, and see their current
  email. The email-change flow already exists at `/account/email` — link to it.
  Files: `internal/handler/`, `internal/service/user.go`, `internal/view/page/`,
  `internal/app/routes.go`

- [ ] **T0302** — Add session listing and management
  Cover: let authenticated users see their active sessions and revoke individual
  sessions. Requires session store enumeration. If cookie-based sessions cannot
  enumerate, document the limitation and implement for Redis-backed sessions.
  Files: `internal/handler/`, session store interface, `internal/app/routes.go`

---

### Phase 04: Reference Workflow Surfaces

Demonstrate real application patterns beyond the existing admin user CRUD.

- [ ] **T0401** — Add a public search/browse/detail reference flow
  Cover: demonstrate query handling, filtering, result rendering, and detail
  pages as a reusable pattern for real products. The existing admin user list is
  a good internal example — this adds a public-facing equivalent (e.g., a
  resource catalog or directory).
  Files: `internal/handler/`, `internal/service/`, `internal/repository/`,
  `internal/view/page/`, `db/sql/`

- [ ] **T0402** — Expand admin patterns beyond user CRUD
  Cover: add list/detail/edit management patterns for at least one additional
  admin-managed resource so the starter demonstrates a true backoffice shape.
  Show how to reuse the existing pagination, inline-edit, and ETag-caching
  patterns.
  Files: `internal/handler/`, `internal/service/`, `internal/repository/`,
  `internal/view/`, `internal/app/routes.go`

- [ ] **T0403** — Standardize fragment and API endpoint conventions
  Cover: codify (in code and route naming) how full-page, HTMX-fragment, and
  programmatic JSON endpoints coexist. The existing `view.Render()` full/partial
  pattern is the foundation — define the naming and routing conventions so
  progressive enhancement remains predictable across all routes.
  Files: `internal/app/routes.go`, route naming conventions, `.decisions/`

---

### Phase 05: Regression Coverage & Package Extraction Safety

- [ ] **T0501** — Add middleware composition integration tests
  Cover: verify status codes, public messages, skipper behavior, and route-level
  middleware integration for the full middleware stack. Test the
  CSRF → CrossOriginGuard → Handler flow end-to-end. Compare pkg middleware
  behavior to Echo v5 expectations.
  Files: `internal/server/`, `internal/app/`, `pkg/security/` test files

- [ ] **T0502** — Add integration tests for IP extraction, rate limiting, and CSRF composition
  Cover: exercise the real starter middleware stack so the extracted packages are
  proven in the same way downstream adopters would use them. Test rate limiter
  under concurrent load. Test IP extraction with and without proxy headers.
  Files: integration test files in `internal/server/`, `internal/app/`

- [ ] **T0503** — Add cross-package regression coverage
  Cover: test the seams between config, app bootstrap, middleware, rendering,
  assets, and package-level abstractions so extraction stays safe over time.
  Focus on the boundaries between `pkg/` packages and `internal/` wiring.
  Files: test files across `internal/app/`, `internal/server/`

---

### Phase 06: Documentation & Deployment

- [ ] **T0601** — Add starter-level examples for new capabilities
  Cover: expose example surfaces that prove locale formatting, email delivery,
  font loading, and any new subsystems work together in the running application.
  Extend the existing `/_components` examples page.
  Files: `internal/handler/`, component examples, `internal/app/routes.go`

- [ ] **T0602** — Ensure each new package has end-to-end adoption docs
  Cover: document how an external project would consume each `pkg/` package and
  how the starter itself wires it into `internal/`. Review
  `pkg/security/README.md` and `pkg/server/README.md` for accuracy after any
  refactors.
  Files: `pkg/*/README.md` (only when explicitly requested)

- [ ] **T0603** — Complete production deployment automation
  Cover: GitHub Actions CI/CD and Docker multi-stage builds already exist. Add:
  documented rollback procedure, staging vs production environment separation
  strategy, and secrets rotation guidance. Ensure `make deploy` covers the full
  lifecycle (build → migrate → deploy → health-check → rollback-on-failure).
  Files: `.github/workflows/`, `Makefile`, `docker/`, deployment docs

---

### Future Considerations

The following capabilities are not on the active roadmap but are recorded here
for future evaluation:

- **OAuth/OIDC external-provider auth** — define Go-native support for
  OAuth/OIDC social sign-in patterns in `pkg/auth/` so the starter can move
  beyond email-only passwordless flows. Start with a single provider (e.g.,
  Google) as a reference implementation. Must coexist with existing email-TOTP.

- **Auth mode coexistence testing** — once external-provider auth is added,
  verify that email-TOTP, email change, and OAuth/OIDC flows coexist without
  breaking session and verification behavior.

---
