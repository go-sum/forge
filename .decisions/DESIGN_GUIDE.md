---
title: Design Principles
description: Current project architecture, ownership boundaries, runtime assembly, and routing/rendering guidance for Forge.
weight: 20
---

# Design Principles

> This guide is the authoritative source for Forge's current architecture and
> ownership rules.
>
> Read this together with [`CLAUDE.md`](../CLAUDE.md),
> [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md),
> [`UI_GUIDE.md`](./UI_GUIDE.md), and [`API_RULES.md`](./API_RULES.md).

---

## 1. Purpose

Forge is a server-rendered Go web application starter built around:

- Echo v5 for HTTP transport
- Gomponents for HTML rendering
- HTMX for progressive enhancement
- PostgreSQL + pgx + sqlc for persistence
- extractable packages under `pkg/`

This guide answers:

- where code belongs
- which layer owns a domain
- how the application is assembled at runtime
- how routing, rendering, persistence, and startup behavior work today

This guide does **not** define low-level coding style, function structure, or
general design-pattern usage. Those rules now live in
[`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md).

---

## 2. Current Architecture Overview

The repo has two design zones:

- `internal/` for application-specific behavior and composition
- `pkg/` for extractable, reusable modules

Runtime assembly is centered in `internal/app`. The composition root wires:

- config loading
- logging
- asset registration
- security and middleware
- database pool and migrations
- sessions
- queue client and background services
- package-owned auth, site, and storage modules
- app-owned feature modules and views

The current application is intentionally hybrid:

- some domains are package-owned and integrated into the app, such as auth,
  queue storage, sessions, senders, and site metadata handlers
- some domains are app-owned, such as contact flow, availability handling, and page composition

Use the ownership rules in [3](#3-domain-ownership-decision-framework) before adding a new feature.

---

## 3. Domain Ownership Decision Framework

Every feature in Forge is either **package-owned** or **app-owned**. That
decision drives where code lives, who owns schema and types, and which layer
holds the main business rules.

### 3.1 The two ownership models

**Package-owned** means the domain lives in `pkg/<name>/` as a self-contained,
extractable module. The package owns its schema, queries, generated code,
repository interfaces, service logic, and any package-level transport helpers.
The app integrates the package from `internal/app`.

**App-owned** means the domain lives in `internal/`. The app owns its types,
error sentinels, persistence, service orchestration, handlers, and views.

### 3.2 Decision criteria

Answer these questions in order. The first "yes" decides ownership.

| # | Question | If yes -> |
|---|----------|-----------|
| 1 | Could this domain deploy cleanly in a different Go application with no `internal/` dependency? | **Package-owned** |
| 2 | Does this domain define a schema contract that other apps would consume the same way? | **Package-owned** |
| 3 | Does the capability already exist in a `pkg/` module? | **Package-owned** |
| 4 | Is the behavior specific to this app's product logic, policies, or page composition? | **App-owned** |
| 5 | Does the feature exist mainly to orchestrate package-owned capabilities? | **App-owned** |

If ownership is still unclear, default to **app-owned** and extract later if
real reuse demand appears.

### 3.3 Current ownership map

| Concern | Owner | Current location |
|---------|-------|------------------|
| app wiring and runtime | app-owned | `internal/app/` |
| public pages and fragments | app-owned | `internal/view/` |
| contact workflow | app-owned | `internal/features/contact/` |
| availability / degraded startup | app-owned | `internal/features/availability/` |
| docs/examples/sessions feature modules | app-owned | `internal/features/*` |
| auth domain | package-owned | `pkg/auth/` |
| queue domain and store | package-owned | `pkg/queue/` |
| session engine | package-owned | `pkg/session/` |
| site metadata handlers | package-owned | `pkg/site/handlers/` |
| generic server/security helpers | package-owned | `pkg/server/`, `pkg/security/` |

### 3.4 Type and error boundaries

For package-owned domains, the app should use package-owned model types and
error sentinels directly unless the app truly needs different semantics or
fields.

Use app-owned wrapper types only when:

- the app needs fields that do not belong in the package
- multiple packages feed the same app-level concept
- the app needs to enforce app-specific invariants in its own type

Do not create field-for-field mirror types or re-export package errors in
`internal/`.

### 3.5 Interface ownership

When the app depends on a package store, the interface stays owned by the
package. `internal/app` may define a combined wiring interface when it needs to
bundle multiple package-owned interfaces together.

Current example:

```go
type AuthStore interface {
	authrepo.UserStore
	authrepo.AdminStore
}
```

This is composition glue, not a transfer of domain ownership.

---

## 4. Composition Root

`internal/app` is the composition root. It is the only place that should know
how the full application is assembled.

### `internal/app` owns

- bootstrap order
- dependency construction
- package integration
- middleware registration
- route registration
- startup degradation behavior
- background service lifecycle

### Current file roles

| File | Current role |
|------|--------------|
| `internal/app/app.go` | builds the runtime, constructs auth handlers, registers routes, starts and stops the app |
| `internal/app/runtime.go` | long-lived runtime resource container |
| `internal/app/runtime_foundation.go` | config, logger, sender, assets, server, and middleware bootstrap |
| `internal/app/runtime_persistence.go` | database, migrations, repository startup checks, queue, and KV bootstrap |
| `internal/app/runtime_application.go` | sessions, auth services, passkey services, and validator bootstrap |
| `internal/app/routes.go` | route-policy composition and concrete route registration |

### Rules

- Feature packages must not reach up into the composition root.
- New background workers, route groups, and package integrations must be wired
  explicitly here.
- If a package needs app configuration, pass it in from `internal/app`.
- Do not make reusable packages read `config.App` directly.

---

## 5. Application Feature Modules

App-owned feature code is organized feature-first under `internal/features/`.

Current feature modules:

- `availability/`
- `contact/`
- `docs/`
- `examples/`
- `public/`
- `sessions/`

### Feature package shape

Each feature package may contain:

- `module.go` for dependency wiring and handler exposure
- `handler.go` for HTTP transport logic
- `service.go` for app-owned orchestration
- feature-local tests and small helpers

Route registration still happens from `internal/app/routes.go`. Feature modules
expose handlers; they do not self-register into Echo.

### Layer responsibilities

Handlers own:

- request parsing
- validation
- calling services or package-owned collaborators
- response rendering, redirects, and HTTP status mapping

Services own:

- app-specific orchestration
- business rules for app-owned features
- queue dispatch coordination

Views own:

- page composition
- partial composition
- request-aware route reversal and presentation state

Persistence for app-owned domains lives in `internal/repository/` only when the
domain is truly app-owned.

For function-level coding rules inside these layers, use
[`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md) and
[`API_RULES.md`](./API_RULES.md).

---

## 6. `pkg/` Module Boundary

Top-level reusable packages under `pkg/` are extractable modules. They must not
import `internal/`.

Current module families include:

- `pkg/auth`
- `pkg/componentry`
- `pkg/kv`
- `pkg/queue`
- `pkg/security`
- `pkg/send`
- `pkg/server`
- `pkg/session`
- `pkg/site`

### Rules

- `pkg/` modules are leaf modules relative to the application.
- Package APIs must accept configuration and collaborators from the host app.
- Package-owned handlers are valid when the package intentionally exposes an
  HTTP surface, as `pkg/auth` and `pkg/site/handlers` do.
- Package-owned persistence is valid when the package intentionally owns a
  reusable domain and schema, as `pkg/auth/pgstore` and `pkg/queue/pgstore` do.
- App-specific page composition still belongs in `internal/view/`.

For config/default ownership and code-structure discipline inside packages, use
[`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md).

---

## 7. Persistence Ownership Model

Forge uses distributed schema ownership.

### Canonical ownership

| Domain | Canonical schema file | Owner |
|--------|------------------------|-------|
| app-owned shared DB objects | `db/sql/schema.sql` | `internal/` |
| users and auth data | `pkg/auth/pgstore/sql/schema.sql` | `pkg/auth` |
| queue jobs | `pkg/queue/pgstore/sql/schema.sql` | `pkg/queue` |

`db/sql/schemas.yaml` is the schema composition registry used for migration
diffing.

### Query ownership

Each package-owned store owns its own sqlc config and generated code:

- `pkg/auth/pgstore/.sqlc.yaml` -> `pkg/auth/pgstore/db/`
- `pkg/queue/pgstore/.sqlc.yaml` -> `pkg/queue/pgstore/db/`

The root `.sqlc.yaml` is reserved for app-owned tables and queries.

### Rules

- App-owned tables belong in `db/sql/schema.sql`.
- Package-owned tables belong beside the owning package.
- If the app needs new behavior from a package-owned table, add it in the
  owning package rather than mirroring that schema or query set in the root.
- Every new schema or query surface must have an explicit owner before
  implementation starts.

---

## 8. Routing Model

Route registration is orchestrated from `internal/app/routes.go`.

There is no separate route-constant package or independent routing composition
layer in the current app. Route policies are assembled inline using
`pkg/server/route` helpers such as `route.Register`, `route.Layout`, and
`route.Group`.

### Current route-policy families

| Policy family | Current use |
|---------------|-------------|
| public GET | home, docs, robots, sitemap, contact form, health |
| public mutation with cross-origin protection and auth rate limit | signin, signup, verify, contact submit, passkey authenticate |
| authenticated read | profile pages, session listing, examples |
| authenticated mutation with cross-origin protection | signout, email change, session revoke, passkey management |
| admin read | elevate page, user list/edit/row |
| admin mutation with cross-origin protection | elevate, user update/delete |

### Route reversal

Use named route reversal through:

- `pkg/server/route.Reverse(...)`
- `pkg/server/route.ReverseWithQuery(...)`
- `view.Request.Path(...)`
- `view.Request.PathWithQuery(...)`

### Rules

- Register every application route with a stable route name.
- Do not hardcode application URLs when a named route already exists.
- Route-policy composition belongs in `internal/app/routes.go`.
- Package-owned handlers may be registered directly by the composition root.

---

## 9. Rendering Model

Forge supports multiple HTML response modes without splitting the app into
separate rendering stacks.

### Canonical rendering modes

| Mode | Handler pattern |
|------|-----------------|
| full page + HTMX partial | `view.Render(c, req, fullPage, partial)` |
| fragment-only | `render.Fragment(c, node)` or `render.FragmentWithStatus(c, status, node)` |
| HTMX removal | `c.String(http.StatusOK, "")` |
| JSON/problem | selected by the global error handler based on request headers |
| redirect | HTMX-aware redirect helpers |

### Rules

- Use `view.NewRequest(...)` to build request-scoped presentation state.
- Use `view.Render(...)` when one endpoint serves both full-page and HTMX
  partial modes.
- Use `render.Fragment(...)` only when the endpoint exists purely for fragment
  swapping.
- Let the global error handler decide between HTML, HTMX fragment, and problem
  JSON responses.

For handler-shape and Echo-specific transport rules, use
[`API_RULES.md`](./API_RULES.md).

---

## 10. Security and Middleware Model

App middleware composition lives in `internal/server/`.

Current responsibilities include:

- secure headers
- CSRF protection
- cross-origin and fetch-metadata protection
- request logging and recovery
- rate limiting
- static asset caching and fragment caching where needed

### Rules

- Use `pkg/security` and `pkg/server` primitives rather than hand-rolling
  security or middleware logic.
- Apply cross-origin protections at unsafe route boundaries.
- Keep app-specific middleware composition in `internal/server/`.
- Keep reusable middleware implementations in `pkg/server` or `pkg/security`
  when they are extractable.

---

## 11. Background Services and Startup Degradation

Forge has first-class background-service support.

### Background service lifecycle

`internal/app/runtime.go` defines the background-service contract:

- services register through `Runtime.AddBackground(...)`
- `App.Start()` starts them before the HTTP server begins serving
- shutdown stops them in reverse registration order

Current example:

- the queue client

### Startup degradation

The app intentionally supports degraded startup when required dependencies are
not ready.

Current flow:

- `initDatabase()` records `StartupError` instead of panicking on every failure
- `App.New(...)` switches to `RegisterStartupRoutes(...)` when full startup
  cannot complete
- `availability.Handler` serves `/health` and degraded responses

### Rules

- Infrastructure that can fail at startup must report clearly through startup
  checks.
- Degraded startup routes must stay minimal and safe.
- Do not assume the app either fully boots or fully crashes; degraded mode is a
  first-class runtime state.

---

## 12. Configuration Architecture

Forge is currently **Go-struct-first**, not YAML-driven.

### Current model

- `config/config.go` defines the root `Config` type and `Load(appEnv string)`.
- `productionDefault()` returns a fully populated root config in Go.
- environment overlays are applied by ordered functions such as
  `developmentConfig`.
- `pkg/server/config.Load(...)` runs defaults, overlays, and validation.
- `RegisterValidationRules` composes cross-field validation at the config
  boundary.

### Ownership

- app runtime config structs live in `config/`
- reusable config loading and validation mechanics live in `pkg/server/config`
- package-specific defaults remain owned by the package that defines them

### Rules

- Keep app-specific config composition in `config/`.
- Keep reusable config mechanics in `pkg/server/config`.
- If a config struct changes, trace all callers and all mappings that depend on
  it.
- Use [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md) for detailed rules on
  package-owned `Config` types, defaults, `cmp.Or`, and validation registration.

---

## 13. How The Guides Fit Together

Use the decision docs this way:

- [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md): where code belongs and how the app is
  assembled
- [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md): how new code should be
  structured and maintained
- [`API_RULES.md`](./API_RULES.md): Echo v5 transport specifics
- [`UI_GUIDE.md`](./UI_GUIDE.md): visual and UI composition guidance

---

## 14. Quick Reference

### Ownership checklist

- [ ] Is this app-owned or package-owned?
- [ ] Does schema ownership match code ownership?
- [ ] Does the route get wired from `internal/app/routes.go`?
- [ ] Is the URL generated from a route name instead of a hardcoded string?
- [ ] Is the rendering path full-page, dual-mode, or fragment-only?
- [ ] Is the behavior tested at the layer that owns it?

### Current source-of-truth map

| Concern | Current source of truth |
|---------|--------------------------|
| app assembly | `internal/app/` |
| route registration | `internal/app/routes.go` |
| app middleware stack | `internal/server/` |
| error handling | `internal/server/error.go` |
| view request state | `internal/view/request.go` |
| app-owned schema | `db/sql/schema.sql` |
| schema composition registry | `db/sql/schemas.yaml` |
| package-owned auth persistence | `pkg/auth/pgstore/` |
| package-owned queue persistence | `pkg/queue/pgstore/` |

### Avoid these stale assumptions

- There is no current `internal/routing/` package.
- Config is currently Go-struct-first, not YAML-file-first.
- `pkg/componentry` is the current shared component library, not `pkg/components`.
- `internal/repository` is not the center of all persistence.
- Package-owned transport and service layers are valid parts of the current
  architecture.
