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

## TASKS

Status markers: `[ ]` outstanding · `[x]` completed · `[~]` in progress

Task ID format: `TXXYY` where `XX` = phase, `YY` = task number within that
phase. This allows inserting new tasks within any phase without renumbering
downstream task numbers.

---

### Phase 01: Immediate Application Hardening

- [ ] **T0101** — Wire an explicit admin route group into the application router
  Cover: add an `admin` group in `internal/app/routes.go`, apply `RequireAdmin`,
  move admin-only surfaces under that group, and add route/middleware coverage
  proving public vs protected vs admin behavior.

- [ ] **T0102** — Automate the production deployment path described by the docs
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

## Current Defaults and Assumptions

- Reusable package design takes priority over app-only implementation shortcuts.
- The starter remains server-rendered and HTMX-first; no SPA runtime is planned.
- The existing asset pipeline remains Go-driven and Node-free.
- Theme switching, toast rendering, asset hashing, sprite generation, and core
  security middleware are considered existing capabilities, not missing features.
- Email starts with development-safe providers first; external provider support
  is a follow-up, not a prerequisite.
- Locale work initially targets formatting primitives only, not translated copy.
