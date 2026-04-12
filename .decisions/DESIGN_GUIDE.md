---
title: Design Principles
description: Architecture, ownership boundaries, runtime assembly, and routing/rendering guidance for this application.
weight: 20
---

# Design Principles

> This guide is the authoritative source for the current architecture and
> ownership rules.
>
> Read this together with [`CLAUDE.md`](../CLAUDE.md),
> [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md), and
> [`UI_GUIDE.md`](./UI_GUIDE.md).

---

## 1. Purpose

This is a server-rendered Go web application built around:

- Echo v5 for HTTP transport
- Gomponents for HTML rendering
- HTMX for progressive enhancement
- PostgreSQL + pgx + sqlc for persistence
- Reusable external modules from `github.com/go-sum/*`

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

The application has one source zone:

- `internal/` for all code authored in this repository

External reusable modules from `github.com/go-sum/*` are ordinary Go
dependencies consumed via `go.mod`. They are not part of this repository.

Runtime assembly is centered in `internal/app`. The composition root wires:

- config loading
- logging
- asset registration
- security and middleware
- database pool and migrations
- sessions
- queue client and background services
- external auth, site, and storage modules wired in from `github.com/go-sum/*`
- app-owned feature modules and views

The current application is intentionally hybrid:

- some domains are provided by external modules and integrated into the app,
  such as auth, queue storage, sessions, senders, and site metadata handlers
- some domains are app-owned, such as contact flow, availability handling, and
  page composition

---

## 3. Ownership Model

All code authored in this repository is **app-owned** and lives in `internal/`.
Reusable functionality lives in the external `github.com/go-sum/*` modules and
is consumed as ordinary Go dependencies. There is no `pkg/` directory in this
repository.

The external modules this application depends on:

- `github.com/go-sum/auth`
- `github.com/go-sum/componentry`
- `github.com/go-sum/kv`
- `github.com/go-sum/queue`
- `github.com/go-sum/security`
- `github.com/go-sum/send`
- `github.com/go-sum/server`
- `github.com/go-sum/session`
- `github.com/go-sum/site`

### Type and error boundaries

Use external module types and error sentinels directly unless the app truly
needs different semantics or fields.

Use app-owned wrapper types only when:

- the app needs fields that do not belong in the external module
- multiple external modules feed the same app-level concept
- the app needs to enforce app-specific invariants in its own type

Do not create field-for-field mirror types or re-export external-module errors
in `internal/`.

### Interface composition

When the app depends on an external module store, the interface stays owned by
the external module. `internal/app` may define a combined wiring interface when
it needs to bundle multiple external-module interfaces together.

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
- external-module integration
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
- New background workers, route groups, and external-module integrations must be
  wired explicitly here.
- If an external module needs app configuration, pass it in from `internal/app`.
- Do not make external modules read `config.App` directly.

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
- calling services or external-module collaborators
- response rendering, redirects, and HTTP status mapping

Services own:

- app-specific orchestration
- business rules for app-owned features
- queue dispatch coordination

Views own:

- page composition
- partial composition
- request-aware route reversal and presentation state

Persistence for app-owned tables lives in `internal/repository/`. Persistence
for tables owned by external `github.com/go-sum/*` modules is owned by those
modules — do not mirror their schemas here.

For function-level coding rules inside these layers, use
[`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md).

---

## 6. External Module Boundary

The `github.com/go-sum/*` modules are external Go dependencies, not part of
this repository. They are consumed via `go.mod`.

### Rules

- Their public APIs accept configuration and collaborators from this app at the
  composition root.
- To change behavior provided by an external module, file a change in the
  upstream repository — do not vendor or fork inside `internal/`.
- External module HTTP handlers are valid when the module intentionally exposes
  an HTTP surface, as `github.com/go-sum/auth` and `github.com/go-sum/site`
  do — register them from `internal/app/routes.go`.
- External module persistence is valid when the module intentionally owns a
  reusable domain and schema, as `github.com/go-sum/auth` and
  `github.com/go-sum/queue` do.
- App-specific page composition still belongs in `internal/view/`.

---

## 7. Persistence Ownership Model

This application uses a single app-owned schema file plus external-module-owned
schemas.

### Canonical ownership

| Domain | Canonical schema | Owner |
|--------|-----------------|-------|
| App-owned tables | `db/sql/schema.sql` | `internal/` |
| Users and auth data | owned by `github.com/go-sum/auth` upstream | external module |
| Queue jobs | owned by `github.com/go-sum/queue` upstream | external module |

`db/sql/schemas.yaml` is the schema composition registry used for migration
diffing.

### Query ownership

All queries for app-owned tables live in `db/sql/queries/*.sql`. The root
`.sqlc.yaml` generates output for app-owned tables only.

External modules own their own sqlc configuration and generated code — do not
add their queries here.

### Rules

- App-owned tables belong in `db/sql/schema.sql`.
- If the app needs new behavior from an external-module-owned table, add the
  capability in the upstream module rather than mirroring that schema or query
  set here.
- Every new schema or query surface must have an explicit owner (this app or an
  external module) before implementation starts.

---

## 8. Routing Model

Route registration is orchestrated from `internal/app/routes.go`.

There is no separate route-constant package or independent routing composition
layer in the current app. Route policies are assembled inline using
`github.com/go-sum/server/route` helpers such as `route.Register`,
`route.Layout`, and `route.Group`.

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

- `github.com/go-sum/server/route.Reverse(...)`
- `github.com/go-sum/server/route.ReverseWithQuery(...)`
- `view.Request.Path(...)`
- `view.Request.PathWithQuery(...)`

### Rules

- Register every application route with a stable route name.
- Do not hardcode application URLs when a named route already exists.
- Route-policy composition belongs in `internal/app/routes.go`.
- External-module handlers may be registered directly by the composition root.

---

## 9. Rendering Model

The application supports multiple HTML response modes without splitting into
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

- Use `github.com/go-sum/security` and `github.com/go-sum/server` primitives
  rather than hand-rolling security or middleware logic.
- Apply cross-origin protections at unsafe route boundaries.
- Keep app-specific middleware composition in `internal/server/`.
- Keep reusable middleware implementations in the upstream external modules when
  they are extractable.

---

## 11. Background Services and Startup Degradation

The application has first-class background-service support.

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

The application is currently **Go-struct-first**, not YAML-driven.

### Current model

- `config/config.go` defines the root `Config` type and `Load(appEnv string)`.
- `productionDefault()` returns a fully populated root config in Go.
- environment overlays are applied by ordered functions such as
  `developmentConfig`.
- `github.com/go-sum/server/config.Load(...)` runs defaults, overlays, and
  validation.
- `RegisterValidationRules` composes cross-field validation at the config
  boundary.

### Ownership

- app runtime config structs live in `config/`
- reusable config loading and validation mechanics live in
  `github.com/go-sum/server/config`
- external-module config defaults remain owned by the upstream module that
  defines them

### Rules

- Keep app-specific config composition in `config/`.
- Keep reusable config mechanics in `github.com/go-sum/server/config`.
- If a config struct changes, trace all callers and all mappings that depend on
  it.
- Use [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md) for detailed rules on
  config types, defaults, `cmp.Or`, and validation registration.

---

## 13. How The Guides Fit Together

Use the decision docs this way:

- [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md): where code belongs and how the app is
  assembled
- [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md): how new code should be
  structured and maintained
- [`UI_GUIDE.md`](./UI_GUIDE.md): visual and UI composition guidance

---

## 14. Quick Reference

### Ownership checklist

- [ ] Is all new code in `internal/`?
- [ ] Does schema ownership match the owning layer (app or external module)?
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
| auth persistence | `github.com/go-sum/auth` (external module) |
| queue persistence | `github.com/go-sum/queue` (external module) |

### Avoid these stale assumptions

- There is no `pkg/` directory in this repository.
- Config is Go-struct-first, not YAML-file-first.
- `github.com/go-sum/componentry` is the shared component library.
- `internal/repository` is not the center of all persistence; external-module
  tables are persisted by their respective upstream modules.
- External-module transport and service layers are valid parts of the
  architecture when the module intentionally exposes them.
