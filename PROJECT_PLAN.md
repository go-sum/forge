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
| pgschema for declarative schema apply (no numbered migration files) | `CLAUDE.md` |
| Koanf-based layered config with `CTX_` env overrides | `config/config.go` |
| Gomponents for type-safe HTML rendering | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| HTMX for partial-page interactions | `CLAUDE.md`, `.decisions/DESIGN_GUIDE.md` |
| Reusable supporting packages under `pkg/` | `CLAUDE.md`, `go.work` |
| Tailwind standalone CLI + Go-driven asset pipeline | `CLAUDE.md`, `cli/build-assets.go` |

---

## Adoption Notes From `../examples/r3-cf/`

The review of `../examples/r3-cf/` surfaced several useful patterns worth
adopting in this repository:

- Strong package-level documentation and explicit public API boundaries.
- A build-time markdown documentation system with runtime navigation.
- Reusable site utilities for `robots.txt` and `sitemap.xml`.
- A predictable font-loading subsystem for remote and self-hosted fonts.
- Production-ready structured logging with field redaction.
- Provider-agnostic email delivery with reusable template components.
- Locale-aware formatting primitives as infrastructure, before full i18n.

The review also surfaced patterns that should **not** be copied directly:

- Cloudflare runtime integration, Worker bindings, and D1-specific plumbing.
- Remix-specific middleware composition and route contracts.
- TypeScript-specific Result APIs as a replacement for idiomatic Go errors.
- Client-island architecture where the current server-rendered + HTMX approach
  is already sufficient.

Several ideas from `r3-cf` are already represented here and therefore do not
need separate adoption tasks:

- Content-hashed asset URLs.
- SVG sprite generation.
- Theme switching.
- Flash/toast rendering.
- CSRF, origin, fetch metadata, and rate-limit protections.
- Problem-details style JSON error responses.

---

## Roadmap Priorities

The roadmap prioritizes **reusable packages first**, then app-level integration
that demonstrates those packages in the starter itself.

When a new capability can live under `pkg/` without breaking the dependency
rules, it should be extracted there first. Application routes, config wiring,
and examples in `internal/` should follow.

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
- Theme switching, toast rendering, asset hashing, sprite generation, and core
  security middleware are considered existing capabilities, not missing features.
- Email starts with development-safe providers first; external provider support
  is a follow-up, not a prerequisite.
- Locale work initially targets formatting primitives only, not translated copy.

---

## TASKS

Status markers: `[ ]` outstanding · `[x]` completed · `[~]` in progress

Task ID format: `TXXYY` where `XX` = phase, `YY` = task number within that
phase. This allows inserting new tasks within any phase without renumbering
downstream task numbers.

---

### Phase 01: Immediate Application Hardening

- [x] **T0101** — Harden the production docker container
  Cover: Research and implement hardening of the environment for production-ready
  docker containers. Implemented: read_only filesystems, cap_drop ALL,
  no-new-privileges on all services, pids_limit, shm_size for PG, log rotation,
  app healthcheck binary for distroless image, .dockerignore, USER nonroot,
  pinned dev tool versions, fixed deploy.sh to use dev_toolchain for pgschema.

- [ ] **T0102** — Wire an explicit admin route group into the application router
  Cover: add an `admin` group in `internal/app/routes.go`, apply `RequireAdmin`,
  move admin-only surfaces under that group, and add route/middleware coverage
  proving public vs protected vs admin behavior.

- [ ] **T0103** — Automate the production deployment path described by the docs
  Cover: script or pipeline the production image build, `pgschema apply`, and
  service restart flow so the documented production path is executable rather
  than manual.

---

### Phase 02: Package Surface and Monorepo Discipline

- [ ] **T0201** — Define the supported public API surface for each extractable package
  Cover: document which packages and symbols are intended for downstream use in
  `pkg/auth`, `pkg/componentry`, `pkg/security`, and `pkg/server`; separate
  examples and internals from the promised public surface.

- [ ] **T0202** — Add package-level READMEs for each extractable module
  Cover: add or expand READMEs covering purpose, dependency rules, install/use
  patterns, configuration, and examples for each published package subtree.

- [ ] **T0203** — Align extraction and sync workflows with the documented API surface
  Cover: verify the sync/release automation matches the intended package
  boundaries and does not export files that should remain internal to the
  monorepo.

- [ ] **T0204** — Add package-health checks for extractable modules
  Cover: ensure each extractable package can be tested, linted, and documented
  independently without depending on application-only internals.

---

### Phase 03: Documentation System

- [x] **T0301** — Adopt Hugo as the documentation compiler
  Cover: add a Hugo-based build step that compiles markdown documentation into
  static output under `public/doc` rather than building a custom Go-native
  markdown pipeline.

- [x] **T0302** — Standardize docs on Hugo-native content conventions
  Cover: ensure docs such as `README.md` and other published markdown sources
  follow Hugo front matter and content structure conventions so Hugo-native
  metadata controls titles, slugs, ordering, sections, and navigation.

- [x] **T0303** — Serve generated docs statically from the application
  Cover: expose the compiled Hugo output through the existing app static-file
  serving path instead of introducing a second runtime markdown renderer or a
  Gomponents-based docs route.

- [x] **T0304** — Use Hugo Chroma highlighting for CSP-safe rendered docs
  Cover: configure Hugo to render syntax-highlighted code blocks with Chroma
  only, keeping docs fully static and CSP-safe without Shiki or client-side
  rendering requirements.

- [x] **T0305** — Publish architecture and package guides through the Hugo docs site
  Cover: compile and serve `.decisions/*`, package guides, `README.md`, and
  other selected operational docs through the Hugo-generated docs site using
  Hugo-native metadata only; defer search and versioning to a later phase.

---

### Phase 04: Site Metadata and SEO Utilities

- [x] **T0401** — Create a reusable `pkg/site` package for site metadata outputs
  Cover: add reusable helpers to generate `robots.txt` and `sitemap.xml` from
  declarative configuration plus route metadata.

- [x] **T0402** — Extend `config/site.yaml` to drive canonical metadata
  Cover: support site origin, robots directives, sitemap defaults, and static
  page inclusion without hard-coding these values in handlers.

- [x] **T0403** — Expose `/robots.txt` and `/sitemap.xml` from the starter app
  Cover: add transport wiring, caching headers where appropriate, and tests for
  output shape and content type.

- [x] **T0404** — Generate sitemap entries from named routes and config-owned pages
  Cover: build a route-to-URL strategy that uses existing route names and app
  config while allowing extra static entries from `site.yaml`.

---

### Phase 05: Font Loading and Asset Ergonomics

- [x] **T0501** — Add a reusable `pkg/componentry/font` subsystem
  Cover: provide Gomponents helpers for remote font providers such as Google,
  Bunny, and Adobe plus self-hosted font preload/link generation.

- [x] **T0502** — Integrate self-hosted fonts with the asset manifest
  Cover: ensure local font files resolve through the same content-hashed public
  asset path strategy already used for CSS, JS, and sprites.

- [x] **T0503** — Add layout-level font configuration hooks
  Cover: let the starter choose fonts declaratively and render the required
  preload and stylesheet nodes from the root layout.

- [x] **T0504** — Document Tailwind token integration for custom fonts
  Cover: define how the font subsystem maps to CSS variables and theme tokens so
  new apps can swap type systems predictably.

---

### Phase 06: Email Delivery and Templating

- [x] **T0601** — Introduce a reusable `pkg/send` package
  Cover: define provider-agnostic message types, sender interfaces, and a small
  service layer that can be reused across starter applications.

- [x] **T0602** — Ship development-safe and production-ready providers
  Cover: implement `noop`, in-memory test capture, and HTTP API providers
  for [`Resend`](https://resend.com/) and [`Mailchannels`](https://www.mailchannels.com/).

- [x] **T0603** — Add Gomponents-based email templates
  Cover: leverage @pkg/componentry to enable reusable HTML email layout components and a plain-text render
  path so email composition stays server-side and type-safe.

- [x] **T0604** — Demonstrate the package with one real workflow
  Cover: the contact page, handler, service, `config/service.yaml`, and `config/service.go`
  are wired; submitting the form sends one message to `send_to` and one confirmation
  message to the submitter from the configured `send_from` address.

- [x] **T0605** — Add mail testing and development inspection support
  Cover: make email behavior testable without third-party services and easy to
  inspect locally during development.

---

### Phase 07: Locale-Aware Formatting

- [ ] **T0701** — Add locale negotiation infrastructure
  Cover: create a reusable locale package that derives a preferred locale from
  `Accept-Language` with configurable fallback behavior.

- [ ] **T0702** — Add reusable number and time formatting helpers
  Cover: support number, currency, date, datetime, and relative-time formatting
  for use in views without coupling business logic to transport concerns.

- [ ] **T0703** — Thread locale and timezone through the request/view layer
  Cover: expose locale-aware formatting at render time without pushing HTTP
  concerns into services or repositories.

- [ ] **T0704** — Keep the initial scope to formatting, not full translation catalogs
  Cover: deliver locale-aware presentation first and defer message translation,
  pluralization catalogs, and content localization to a later roadmap update.

---

### Phase 09: Integration and Example Surfaces

- [ ] **T0901** — Add starter-level examples for new reusable capabilities
  Cover: expose example surfaces or small application workflows that prove docs,
  site metadata, fonts, logging, email, and locale formatting work together.

- [ ] **T0902** — Ensure each new package has end-to-end adoption docs
  Cover: document how an external project would consume the package and how the
  starter itself wires it into `internal/`.

- [ ] **T0903** — Add regression coverage around cross-package interactions
  Cover: test the seams between config, app bootstrap, middleware, rendering,
  assets, and package-level abstractions so extraction stays safe over time.

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

### Phase 05: CORS Middleware and HTTP Method Override

**Why:** Forge's `pkg/security` correctly rejects unsafe cross-origin requests for
browser-initiated state mutations. However, it has no CORS support for legitimate
cross-origin consumers (REST API clients, webhooks, future mobile apps). Separately,
HTML forms only support `GET` and `POST` — without a method-override middleware,
HTMX `DELETE`/`PUT`/`PATCH` interactions require JavaScript-driven requests and
cannot degrade gracefully to plain HTML forms.

- [x] **T0501** — Add CORS middleware to `pkg/security/cors`
  Cover: implement configurable allowed-origins (exact, wildcard, function), preflight
  `OPTIONS` handling, `Access-Control-Allow-Credentials`, and `Vary: Origin` when
  reflecting the request origin. Support a static list, a regex set, and a dynamic
  function policy. Wire into `internal/app/routes.go` as an opt-in group middleware
  for API route groups. Add pass/fail tests for each origin policy mode.

- [x] **T0502** — Add HTTP method override middleware to `pkg/server/middleware/methodoverride`
  Cover: read `_method` from the parsed POST body and promote it to the effective
  HTTP method before routing. Support `PUT`, `PATCH`, and `DELETE` only (reject
  arbitrary values). Add a configurable field name (default `_method`). Wire into
  the application middleware stack after body parsing. Add handler and middleware
  tests verifying correct method promotion and rejection of disallowed overrides.

- [x] **T0503** — Wire CORS into the starter and demonstrate method override in the UI
  Cover: configure CORS for a `/api/` route group in `internal/app/routes.go`.
  Add at least one form interaction in the starter that uses `_method=DELETE` to
  exercise method override through the full stack.

---

### Phase 06: Pluggable Server-Side Session Storage

**Why:** `pkg/auth/session` was hardwired to `gorilla/sessions.CookieStore`. Cookie
sessions are limited to ~4 KB, cannot be enumerated or invalidated server-side, and
carry all session data to the client on every request. Production applications that
need forced logout, concurrent-session controls, or large session payloads require
server-side storage.

- [x] **T0601** — Extract a `Store` interface in `pkg/auth/session`
  Cover: `session.Store` interface defined in `pkg/auth/session/store.go`.
  `SessionManager` (CookieStore) implements it with compile-time check.
  All callers (handler, middleware, container) updated to use `session.Store`.
  `NewSessionStore` returns concrete type; interface enforced at caller sites.

- [x] **T0602** — Implement a PostgreSQL-backed session store using hstore
  Cover: enabled hstore extension in all DB init scripts. Added `sessions` table
  with hstore `data` column, expiry index, and update trigger to `db/sql/schema.sql`.
  Implemented `internal/adapters/sessionpg.PGStore` satisfying `session.Store`. Stores
  session ID in a minimal cookie; all data persists in PostgreSQL hstore. Supports
  `user_id`, `display_name`, and `pending_flow` (JSON-in-hstore) keys. Includes
  `DeleteExpired()` for session cleanup. Tests against test\_data container.

- [x] **T0603** — Implement a Redis-backed session store via `kv.Store` injection
  Cover: implemented `internal/adapters/sessionkv.KVStore` satisfying `session.Store`.
  Accepts `kv.Store` interface via constructor (dependency injection from `internal/app`).
  Stores session data as JSON, keyed by `{prefix}{sessionID}` with TTL matching MaxAge.
  Session ID stored in minimal cookie. Reuses the existing `pkg/kv` connection pool
  (Dragonfly). No cross-module import — `pkg/auth` leaf-node rule preserved.

- [x] **T0604** — Wire configurable session backend into app bootstrap
  Cover: added `session.backend` config key (`cookie` | `postgres` | `redis`) with
  default `cookie`. Updated `initAuth()` with backend switch to construct the correct
  store. `initKV()` runs before `initAuth()` so `c.KV` is available for redis backend.
  Panics with clear message if `backend=redis` but `kv.enabled=false`. CookieStore
  remains default — zero config change for existing deployments.

---

### Phase 07: Typed HTTP Header Utilities

**Why:** Go's `net/http` represents all headers as raw strings. Parsing
`Accept-Language` for locale negotiation (Phase 03), reading `Cache-Control`
directives, or building correct `Vary` and `ETag` headers all require manual string
parsing today. Remix 3's `headers` package demonstrates the value of purpose-built
classes for common headers that parse, manipulate, and round-trip back to strings.
A Go equivalent in `pkg/server/headers` would directly unblock Phase 03 locale work
and make future caching work safe.

- [x] **T0701** — Add `pkg/server/headers` with `Accept-Language` and `Accept` parsers
  Cover: implement `AcceptLanguage` and `Accept` types that parse the raw header
  string into a quality-weighted slice, expose `Preferred(candidates []string) string`,
  and implement `String()` for round-trip serialization. Add table-driven tests
  covering empty headers, single values, multiple values with q-weights, and
  malformed input. This directly feeds T0301 in Phase 03.

- [x] **T0702** — Add `Cache-Control` and `Vary` utilities to `pkg/server/headers`
  Cover: implement a `CacheControl` type with directive accessors (`MaxAge()`,
  `NoStore()`, `Private()`, `Public()`, `MustRevalidate()`, etc.) and a builder for
  constructing directives programmatically. Add `Vary` helper to safely append values
  without duplicates. Add round-trip tests for common directives.

- [x] **T0703** — Integrate `Accept-Language` parser into Phase 03 locale negotiation
  Cover: replace any raw `Accept-Language` string parsing in the locale package
  (T0301) with `pkg/server/headers.AcceptLanguage`. This makes the locale package
  depend only on the headers package, not on raw HTTP parsing logic.

---

### Phase 08: HTTP Response Caching

**Why:** HTMX partial-page requests are a natural fit for conditional GET caching.
Without ETags or `Last-Modified` support on HTML fragment endpoints, every HTMX
poll or refetch pays full render cost regardless of whether the content changed.
Remix 3's `response` package provides ETag generation, conditional request handling
(`If-None-Match`, `If-Modified-Since`), and Cache-Control helpers as first-class
primitives. A Go analogue in `pkg/server/cache` would reduce server load and
improve perceived performance for data-heavy pages.

- [x] **T0801** — Add ETag generation utilities to `pkg/server/cache`
  Cover: implement `WeakETag(content []byte) string` (size + hash-based) and
  `StrongETag(content []byte) string` (full SHA-256 hash). Add helpers to write
  `ETag` and `Last-Modified` response headers and to check `If-None-Match` /
  `If-Modified-Since` request headers. Return a boolean indicating whether a
  `304 Not Modified` response is appropriate. Add tests for both ETag flavors
  and for the conditional-request decision logic.

- [x] **T0802** — Add a conditional-response middleware to `pkg/server/middleware`
  Cover: implement middleware that computes a weak ETag over the rendered response
  body, compares it against `If-None-Match`, and short-circuits with `304` when
  the content is unchanged. Designed for opt-in use on individual route groups
  (e.g., HTMX fragment endpoints). Add tests verifying 200 on first request,
  304 on matching ETag, and 200 on changed content.

- [x] **T0803** — Demonstrate response caching on at least one HTMX fragment endpoint
  Cover: apply the conditional-response middleware to one fragment route in the
  starter and add an integration test that drives the full 200 → 304 → 200 cycle.

---

  ## Gap Analysis 2 — Remix 3 Feature Review - (codex review)

  ### Phase 09: Authentication and Identity Expansion

  - [ ] T0901 — Add external-provider auth capability to the reusable auth surface
    Cover: define Go-native support for OAuth/OIDC/social sign-in patterns so Forge moves beyond email-only passwordless flows.
  - [ ] T0902 — Add password recovery and account recovery flows
    Cover: introduce forgot-password/reset-password style recovery paths, token lifecycle rules, and starter UI wiring for recovery journeys.
  - [ ] T0903 — Expand authenticated account lifecycle surfaces
    Cover: add a real account area pattern beyond email change, including profile/settings/session-oriented flows expected in a production app.
  - [ ] T0904 — Verify coexistence of auth modes
    Cover: add tests proving email-based auth, recovery, and future external-provider auth can coexist without breaking current session and verification behavior.

  ### Phase 11: World-Class Reference Workflow Surfaces

  - [ ] T1101 — Add a search/browse/detail reference flow to the starter
    Cover: demonstrate query handling, filtering, result rendering, and detail pages as a reusable pattern for real products.
  - [ ] T1102 — Add a fuller authenticated account area
    Cover: introduce account dashboard/settings/history-style routes that demonstrate protected layouts and user-centric navigation patterns.
  - [ ] T1103 — Expand admin reference patterns beyond current user CRUD
    Cover: add list/detail/edit management patterns for at least one additional admin-managed resource so the starter demonstrates a true backoffice shape.
  - [ ] T1104 — Standardize fragment and API endpoint conventions
    Cover: define how full-page, fragment, and programmatic endpoints coexist in the starter so progressive enhancement remains predictable.

  ### Phase 12: Realtime and Navigation Interaction Patterns

  - [ ] T1201 — Add SSE-based live update support
    Cover: provide a reusable server-sent events pattern and a starter example for notifications, progress, or live status updates.
  - [ ] T1202 — Add frame-like partial navigation patterns using HTMX
    Cover: adapt the conceptual benefits of Remix frame navigation into Go/Echo/HTMX shell, section, and partial-refresh conventions.
  - [ ] T1203 — Add regression coverage for streaming and partial navigation
    Cover: verify route behavior, fallback rendering, and progressive enhancement across full-page, fragment, and live-update interactions.

  ### Phase 13: KV Store

  Why: The roadmap requires stateful infrastructure for server-side sessions, caching,
  and lightweight app-local coordination. Dragonfly (Redis-compatible) was selected as
  the primary KV provider, superseding the earlier Badger v4 decision. Dragonfly provides
  Redis protocol compatibility, multi-threaded performance, and a single Docker service
  for both dev and production. `pkg/kv` defines a backend-agnostic `Store` interface;
  future backends (Badger, etcd) can implement it without breaking consumers.

  - [x] T1301 — Record the provider decision and scope boundary
    Cover: Dragonfly (Redis-compatible) selected as the KV provider via
    `pkg/kv/redisstore`. Badger v4 decision superseded — Dragonfly provides
    both embedded-feel simplicity (single Docker service) and Redis protocol
    compatibility. pkg/kv defines a backend-agnostic Store interface; future
    backends (Badger, etcd) can implement it without breaking consumers.
  - [x] T1302 — Add a new extractable pkg/kv module and define the v1 API
    Cover: created standalone `pkg/kv` module with `kv.Store` interface (Ping,
    Get, Set, Delete, Exists, Close), `kv.Scanner` interface (Scan), `SetOptions`
    with TTL, and `ErrNotFound`/`ErrClosed` sentinels. Added to `go.work` and
    root `go.mod` replace directives. Uses string keys and `[]byte` values.
  - [x] T1303 — Implement pkg/kv/redisstore
    Cover: wrapped `github.com/redis/go-redis/v9` behind the `pkg/kv` interfaces.
    Maps `redis.Nil` to `kv.ErrNotFound`. Supports configurable pool size, timeouts,
    password, and database number. Cursor-based `Scan` implementation. Compile-time
    interface checks for both `kv.Store` and `kv.Scanner`.
  - [x] T1305 — Wire optional KV configuration and lifecycle into the starter
    Cover: added `app.kv` config in `config/app.yaml` with `enabled`, `backend`,
    and `redis` sub-config. Added `KVConfig`/`RedisKVConfig` structs to `config/app.go`.
    Added `initKV()` to bootstrap with Ping health check. Container holds `kv.Store`,
    closes on shutdown. Default state is disabled; development overlay enables it.
  - [x] T1306 — Add one starter-level reference consumer in internal/
    Cover: `internal/adapters/sessionkv` uses `kv.Store` for Redis-backed session
    storage. Accepts `kv.Store` via constructor injection from `internal/app`.
    Demonstrates the full lifecycle: bootstrap → inject → use → shutdown.
  - [x] T1307 — Add package, config, and integration coverage for the KV subsystem
    Cover: `pkg/kv/kv_test.go` verifies error sentinel identity. `pkg/kv/redisstore/
    redisstore_test.go` provides 17 integration tests covering CRUD, TTL expiry,
    prefix scans, binary values, empty values, multi-delete, existence checks, close
    behavior, and callback error propagation. Tests skip gracefully without infra.
  - [ ] T1304 — Add Dragonfly-specific maintenance behavior (optional future)
    Cover: background expired-key cleanup is handled natively by Dragonfly/Redis.
    If needed, add periodic `DeleteExpiredSessions` scheduling for the pgstore backend.

  ## Public APIs / Interfaces

  - Module: `pkg/kv` (`github.com/go-sum/kv`)
  - Backend package: `pkg/kv/redisstore`
  - V1 API shape:
      - `Store.Ping(ctx) error`
      - `Store.Get(ctx, key) ([]byte, error)`
      - `Store.Set(ctx, key, value, SetOptions) error`
      - `Store.Delete(ctx, keys...) error`
      - `Store.Exists(ctx, keys...) (int64, error)`
      - `Store.Close() error`
      - `Scanner.Scan(ctx, pattern, fn) error`
      - `SetOptions{TTL time.Duration}`
  - Explicitly out of scope for v1: transactions, watch streams, leases/locks/elections,
    replication, remote protocol exposure, and cross-pkg/ refactors.

  ## Test Plan

  - CRUD roundtrips with exact byte equality.
  - TTL expiry and post-expiry reads return `kv.ErrNotFound`.
  - Prefix scan returns matching keys; empty scan calls fn zero times.
  - Delete of non-existent keys is a no-op.
  - Close prevents further operations.
  - App config/bootstrap/shutdown wiring behaves correctly when KV is enabled and disabled.

  ## Assumptions

  - Dragonfly is the primary backend; Badger remains a documented future option.
  - `pkg/kv` is an extractable offering consumed by `internal/` via dependency injection.
  - Docker services (`app_kv`, `test_kv`) provide dev and test infrastructure.
  
## Echo v5 Package Alignment Review (2026-03-29) (codex review)

This review compared `pkg/server/` and `pkg/security/` against the local Echo v5 reference at [`../examples/echov5/`](../examples/echov5/) and the current application wiring under `internal/server/` and `internal/app/`.

### Review Conclusions

- Keep and preserve: `pkg/server/route`, `pkg/server/apperr`,
  `pkg/security/origin`, `pkg/security/fetchmeta`, `pkg/security/token`, and
  `pkg/security/middleware.CrossOriginGuard` all add value beyond raw Echo v5
  and should remain package-level features.
- Refactor to leverage Echo more directly: `pkg/security/ratelimit` currently
  reimplements Echo v5 rate limiting instead of composing
  `middleware.RateLimiterWithConfig`, and it drops native Echo features such as
  `Skipper`, `BeforeFunc`, `ToMiddleware`, and expiring in-memory visitor
  cleanup.
- Tighten standard Echo integration: `pkg/server/validate` should satisfy
  Echo's `Validator` interface so callers can use `e.Validator` and
  `c.Validate(...)` without adapters.
- Reduce thin pass-through surface: the root `pkg/server` package currently
  wraps `echo.New()` and only exposes a narrow subset of `echo.StartConfig`.
  It should either expose meaningful opinionated configuration over
  `echo.NewWithConfig` / `echo.StartConfig` or stop duplicating Echo's surface.
- Retain the HMAC CSRF improvement, but modernize its Echo integration:
  `pkg/security/csrf` adds real value over Echo's cookie-based CSRF middleware,
  but it should adopt Echo v5 middleware conventions (`ToMiddleware`,
  `Skipper`, explicit default config, richer token lookup/config validation)
  instead of remaining a bespoke wrapper.
- Move static cache behavior closer to Echo static routing:
  `pkg/server/middleware.StaticCacheControl` is useful, but it is currently
  path-prefix global middleware rather than a route-level composition with
  `Echo.Static(...)` / `Echo.StaticFS(...)`.
- Wire the app to Echo's trusted proxy/IP model explicitly: request logging and
  rate limiting depend on `c.RealIP()`, but the application currently does not
  configure `Echo.IPExtractor`.

### Echo v5 Alignment Phases

#### Phase 14: Preserve What Is Already Better Than Echo Defaults

- [ ] **T1401** — Keep `pkg/server/route` as the named-route registration and reversal layer
  Cover: retain `route.Add`, `route.Reverse`, and `route.SafeReverse` as the
  package-level policy around Echo v5 `AddRoute` / `Routes.Reverse`; only add
  tests or API polish if needed.

- [ ] **T1402** — Keep `pkg/server/apperr` as the transport-safe error envelope
  Cover: continue using Echo status-coder compatibility, but keep the package's
  separation between internal cause and public message rather than collapsing
  into raw `echo.HTTPError`.

- [ ] **T1403** — Keep the framework-agnostic security primitives
  Cover: retain `token`, `origin`, `fetchmeta`, and `headers` as independent
  reusable packages because Echo v5 does not provide equivalent primitives at
  this abstraction level.

#### Phase 15: Replace Reimplementation With Echo-Native Composition

- [x] **T1501** — Refactor `pkg/security/ratelimit` to wrap Echo v5 rate limiting instead of reimplementing it
  Cover: rebuild the package around `middleware.RateLimiterWithConfig`,
  preserve the current public-message behavior through custom `DenyHandler` /
  `ErrorHandler`, and stop maintaining a parallel middleware/store API where
  Echo already has one.

- [x] **T1502** — Restore expiring visitor cleanup and store configurability in rate limiting
  Cover: use Echo's expiring memory store directly or provide a thin adapter
  over it so the package no longer has unbounded in-process identifier growth.

- [x] **T1503** — Make `pkg/server/validate.Validator` implement Echo's `Validator` interface
  Cover: add a `Validate(any) error` method, keep direct access to the
  underlying go-playground validator for form packages, and verify that an
  `*echo.Echo` can register the package directly as `e.Validator`.

- [x] **T1504** — Reshape the root `pkg/server` surface around real Echo configuration points
  Cover: either replace `server.New()` with an opinionated wrapper over
  `echo.NewWithConfig` or remove the thin pass-through behavior; expose only
  the Echo v5 configuration hooks that the package intends to standardize.

- [x] **T1505** — Expand `pkg/server.Start` only where it adds value beyond Echo's `StartConfig`
  Cover: preserve graceful signal shutdown, but decide whether package callers
  need supported access to `BeforeServeFunc`, `ListenerAddrFunc`,
  `ListenerNetwork`, TLS startup, and banner/port controls instead of hard
  narrowing Echo's startup surface.

- [x] **T1506** — Remove or justify zero-policy pass-through helpers in `pkg/server/database`
  Cover: keep `Connect` and `IsUniqueViolation`, but either inline or expand
  `CheckHealth` / `Close` only if they will carry shared package policy beyond
  `pool.Ping(...)` and `pool.Close()`.

#### Phase 16: Keep the Improvements, but Adopt Echo v5 Middleware Conventions

- [x] **T1601** — Refactor `pkg/security/csrf` into an Echo-style configurable middleware
  Cover: keep the signed stateless token model, but add a default config,
  `ToMiddleware`, `Skipper`, config validation, and richer token lookup options
  so it behaves like a first-class Echo v5 middleware instead of a one-off
  helper.

- [x] **T1602** — Explicitly define the boundary between `csrf` and `CrossOriginGuard`
  Cover: document and test the intended composition of HMAC CSRF, Origin
  validation, and Fetch Metadata checks so the package set presents one
  coherent Echo-facing security story rather than overlapping middleware.

- [x] **T1603** — Give `pkg/security/middleware.CrossOriginGuard` an Echo-style config surface
  Cover: add a config type with `ToMiddleware` / `Skipper` support while
  retaining the current Origin + Fetch Metadata combination that Echo does not
  provide out of the box.

- [x] **T1604** — Rework static cache-control into route-level Echo composition
  Cover: prefer an API that is attached directly to `Echo.Static(...)` /
  `Echo.StaticFS(...)` registration or to a dedicated static-route helper so
  cache headers are scoped to actual static responses instead of a global path
  prefix check.

#### Phase 17: Adopt the Missing Echo v5 Runtime Features in the Starter

- [ ] **T1701** — Configure `Echo.IPExtractor` in starter bootstrap
  Cover: select the correct Echo v5 extractor for the deployment model
  (`ExtractIPDirect`, `ExtractIPFromRealIPHeader`, or
  `ExtractIPFromXFFHeader` with trust options) so `c.RealIP()` is trustworthy
  for rate limiting and request logging.

- [ ] **T1702** — Fix request logging to use Echo v5 `RequestLogger` field capture explicitly
  Cover: enable the required `Log*` flags in the application's
  `RequestLoggerWithConfig` usage or move that policy into `pkg/server/logging`
  so the current logger does not rely on unset Echo values.

- [ ] **T1703** — Register the package validator directly on Echo once `pkg/server/validate` is updated
  Cover: set `e.Validator`, use `c.Validate(...)` where it simplifies handlers,
  and keep existing form-layer access to the raw validator where needed.

#### Phase 18: Regression Coverage and Package Extraction Safety

- [ ] **T1801** — Add package-level tests that compare pkg middleware behavior to Echo v5 expectations
  Cover: verify status codes, public messages, skipper behavior, config
  validation, and route-level middleware integration for every refactored Echo
  wrapper.

- [ ] **T1802** — Add integration tests for trusted IP extraction, rate limiting, and CSRF/cross-origin composition
  Cover: exercise the real starter middleware stack so the extracted packages
  are proven in the same way downstream adopters would use them.

- [ ] **T1803** — Re-review package READMEs after refactors to remove claims of wrapping Echo where the code does not
  Cover: make the public package docs describe composition over Echo v5
  accurately, especially for `pkg/security/ratelimit`, `pkg/security/csrf`, and
  the root `pkg/server` package.

---
