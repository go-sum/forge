---
title: Design Principles
description: Current project architecture, composition rules, persistence ownership, and implementation guidance for Forge.
weight: 20
---

# Design Principles

> This guide is the authoritative source for Forge's project-specific design
> rules. It documents the architecture that is implemented in this repo today,
> with transitional notes called out explicitly where the current state is still
> evolving.
>
> Read this together with [`CLAUDE.md`](../CLAUDE.md),
> [`UI_GUIDE.md`](./UI_GUIDE.md), and [`API_RULES.md`](./API_RULES.md).

---

## 1. Purpose

Forge is a server-rendered Go web application starter built around:

- Echo v5 for HTTP transport
- Gomponents for HTML rendering
- HTMX for progressive enhancement
- PostgreSQL + pgx + sqlc for persistence
- reusable extractable packages under `pkg/`

This guide answers:

- where code belongs
- who owns data and routes
- how the app is assembled at runtime
- how full-page and fragment rendering coexist
- how to extend the repo without fighting its boundaries

This guide does **not** document every exported package API. For package
surfaces, read the package source and tests.

---

## 2. Current Architecture Overview

The repo has two distinct design zones:

- `internal/` for application-specific composition and app-owned behavior
- `pkg/` for extractable modules that can outlive this application

Current runtime assembly is centered in `internal/app`. That composition root
wires:

- config loading
- logging
- assets
- security and middleware
- database pool
- sessions
- queue client and background services
- package-owned auth and site handlers
- app-owned handlers and services

The current app is a hybrid:

- some domains are package-owned and integrated into the app, such as auth and queue storage
- some behavior is app-owned, such as contact flow, availability handling, page composition, and route assembly

Do not document Forge as if every domain is already fully app-owned. That is a
future design direction, not the complete current state.

### Package-Owned Domains vs App-Owned Domains

Use this distinction before deciding where code belongs:

- package-owned domains keep their reusable business logic, persistence, and any
  package-owned transport helpers inside the package
- app-owned domains keep their orchestration, transport, rendering, and
  app-specific persistence in `internal/`

Current examples:

- package-owned: auth, queue storage, site handler generation
- app-owned: contact workflow, docs surface, availability handling, app layout and pages

---

## 3. Composition Root

`internal/app` is the composition root. It is the only place that should know
how the full application is assembled.

### `internal/app` owns

- bootstrap order and startup wiring
- dependency construction
- package integration
- middleware registration
- route registration
- startup degradation behavior
- background service lifecycle

### `internal/app` currently consists of

- `app.go`: app construction, route registration, lifecycle start/shutdown
- `bootstrap.go`: config, logger, assets, DB, auth, queue, session, validator, services
- `container.go`: shared dependency container and background-service lifecycle
- `routes.go`: route groups and route registration, including degraded startup routes
- `authadapter.go`: adapters between app-local dependencies and `pkg/auth` interfaces
- `docs.go`: docs static-file handler for `/docs`

### Rules

- Do not make feature packages reach up into the composition root.
- Do not hide runtime registration behind implicit magic.
- New background workers, route groups, and package integrations must be wired explicitly here.
- If a package needs app configuration, pass it in from `internal/app`; do not make the package read `config.App` directly.

---

## 4. Layer Rules Inside `internal/`

The application-specific part of the repo still follows a layered model, but
it is narrower than older docs implied.

```
Transport   internal/handler/         App-owned HTTP handlers
Service     internal/service/         App-owned business orchestration
Repository  internal/repository/      App-owned persistence helpers
View        internal/view/            Request state, pages, partials, layouts
Model       internal/model/           App-owned domain errors and shared types
```

### Transport layer — `internal/handler/`

Handlers own:

- request parsing
- form binding and validation
- calling the next layer down
- response rendering and redirects
- mapping domain failures to transport outcomes

Handlers must not:

- own business policy
- construct SQL
- directly own infrastructure lifecycle

### Service layer — `internal/service/`

Services own:

- app-specific orchestration
- business rules for app-owned features
- pagination and caps
- queue dispatch orchestration for app workflows

Today this layer is mixed in an important way:

- app-owned services exist in `internal/service`
- package-owned services also exist, for example `pkg/auth/service`

That means `internal/service` is not the only place business logic can live.
Document and implement based on ownership, not on an outdated rule that every
service must be internal.

### Repository layer — `internal/repository/`

`internal/repository` is currently small and app-specific. It is not the
primary home of all persistence in this repo.

Use it for:

- app-owned persistence code
- app-owned DB checks and app-specific query helpers
- mapping infrastructure failures to app-owned errors where the owning domain is internal

Do not pretend all database access flows through this layer today. Package-owned
stores exist and are first-class.

### Model layer — `internal/model/`

`internal/model` currently owns:

- app-owned sentinel errors
- small shared app-domain structs such as contact inputs

It does **not** yet fully own every domain type used by the app. For example,
the user-management surface still relies on `pkg/auth/model` types. Treat that
as current state, not as proof that the app-domain boundary is complete.

### View layer — `internal/view/`

`internal/view` owns:

- request-scoped presentation state via `view.Request`
- full-page constructors
- partial constructors
- shared layout composition
- request-aware route reversal in the view layer

It is the app-owned presentation layer above `pkg/componentry`.

---

## 5. `pkg/` Module Boundary

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

- `pkg/` modules are leaf modules relative to `internal/`.
- Package APIs must accept configuration and collaborators from the host app rather than reaching back into app state.
- Package-owned handlers are allowed when the package intentionally exposes an HTTP surface, such as `pkg/auth` and `pkg/site/handlers`.
- Package-owned persistence is allowed when the package intentionally owns a reusable domain and its schema, such as `pkg/auth/pgstore` and `pkg/queue/pgstore`.

### `pkg/componentry`

`pkg/componentry` is the shared UI/component library. Use it instead of the old
`pkg/components` name referenced in stale docs.

Its import hierarchy is governed by the component DAG documented in code and
package docs. App-specific page composition belongs in `internal/view`, not in
the component package itself.

---

## 6. Persistence Ownership Model

Forge uses distributed schema ownership.

### Canonical ownership

| Domain | Canonical schema file | Owner |
|-------|------------------------|-------|
| app-owned shared DB objects | `db/sql/schema.sql` | `internal/` |
| users | `pkg/auth/pgstore/sql/schema.sql` | `pkg/auth` |
| queue jobs | `pkg/queue/pgstore/sql/schema.sql` | `pkg/queue` |

`db/sql/schemas.yaml` is the composition registry used for migration diffing.

### Query ownership

There are two query-generation patterns in the repo:

- root sqlc config in `.sqlc.yaml`, which targets `db/sql/queries/` and emits to `db/schema`
- package-local sqlc configs in `pkg/*/pgstore/.sqlc.yaml`, which emit generated code inside the owning package

Current active package-owned stores:

- `pkg/auth/pgstore`
- `pkg/queue/pgstore`

### Rules

- App-owned tables belong in `db/sql/schema.sql`.
- Package-owned tables belong beside the owning package.
- App-owned query code belongs in root sqlc inputs and outputs when the domain is internal.
- Package-owned query code belongs inside the owning package and must not cross module boundaries.
- Every new table or query surface must declare its owner before implementation: `internal` or a specific package.
- Do not add runtime schema-install hooks in stores. Schema changes flow through migrations and `db/sql/schemas.yaml`.

### When to use `internal/repository` vs `pkg/*/pgstore`

Use `internal/repository` when:

- the domain is app-specific
- the schema is app-owned
- the query shape is not intended for extraction

Use `pkg/*/pgstore` when:

- the domain is intentionally package-owned
- the package owns the schema contract
- the package is meant to be reusable across apps

### Current reality note

The repo currently leans heavily on package-owned persistence for auth and queue
domains. Do not write guidance that assumes every feature must first create an
internal repository.

---

## 7. Routing Model

Route registration currently lives in `internal/app/routes.go`.

There is no separate `internal/routes/routes.go` source of truth at present.
The durable contract is instead:

- routes are registered with names in `internal/app/routes.go`
- links and redirects are generated by reversing those names against the live route table

### Current route groups

`RegisterRoutes` currently assembles these groups:

- static assets under the public prefix
- `publicPost` for cross-origin-guarded public POST routes
- `authGuarded` for authenticated routes
- `authGuardedPost` for authenticated unsafe methods
- `adminGuarded` for authenticated admin-only routes
- `adminGuardedPost` for authenticated admin-only unsafe methods

This admin grouping is already implemented. Do not describe admin grouping as a
future-only concept.

### Route reversal

Use named route reversal through:

- `pkg/server/route.Reverse(...)`
- `pkg/server/route.ReverseWithQuery(...)`
- `view.Request.Path(...)`
- `view.Request.PathWithQuery(...)`

### Rules

- Do not hardcode application route paths when a named route already exists.
- Register every app route with a stable route name.
- Resolve URLs from route names at the point of use.
- Package-owned handlers may be registered directly by the composition root.

---

## 8. Rendering Model

Forge supports multiple HTML response modes without splitting the app into
separate rendering stacks.

### Full page and fragment rendering

The standard pattern is:

- build `req := view.NewRequest(...)`
- render with `view.Render(...)` or `view.RenderWithStatus(...)`
- pass both a full page and, when needed, a fragment region

`view.Render` chooses between:

- full-page rendering for normal requests
- fragment rendering for HTMX partial requests

If there is no meaningful fragment variant, pass `nil` for the partial and the
full component will be used for both.

### HTMX-only fragment endpoints

Use `render.Fragment(...)` or `render.FragmentWithStatus(...)` for endpoints
whose only purpose is a fragment swap.

Current examples include row/form fragment handlers in the user-management flow.

### Redirect behavior

Use the established HTMX-aware redirect helpers where the flow needs to support
both boosted/partial and normal navigation.

### Error rendering

The global error handler lives in `internal/server/error.go`.

It currently selects among:

- RFC 7807 problem JSON for JSON/problem requests
- out-of-band HTMX flash/toast fragments for partial requests
- rendered HTML error pages for normal browser requests

Handlers should return classified errors rather than reimplementing this branching.

---

## 9. Security and Middleware Model

Middleware registration currently lives in `internal/server/middleware.go`.

Security composition currently lives in `internal/server/security.go` and includes:

- secure headers
- CSRF protection
- CORS helpers for opt-in groups
- origin and fetch-metadata guards
- request logging and recovery

### Rules

- Use `pkg/security` primitives rather than hand-rolling security checks.
- Apply cross-origin protections at unsafe route boundaries.
- Keep app-specific middleware composition in `internal/server`, not in reusable packages.
- Keep generic middleware implementations in `pkg/server` or `pkg/security` when they are extractable.

---

## 10. Background Services and Startup Degradation

Forge has first-class background-service support.

### Background service lifecycle

`internal/app/container.go` defines the background-service contract:

- services self-register through `Container.AddBackground(...)`
- `App.Start()` calls `StartBackground(...)` before the HTTP server starts
- shutdown stops background services in reverse registration order

Current background service example:

- the queue client, when configured in async mode

### Startup degradation

The app intentionally supports degraded startup when required dependencies are
not ready.

Current flow:

- `initDatabase()` records `StartupError` instead of always panicking
- `App.New(...)` selects `RegisterStartupRoutes(...)` when startup cannot fully wire the app
- `AvailabilityHandler` serves `/health` and degraded responses

### Rules

- Infrastructure that can fail at startup must report clearly through startup checks.
- Degraded startup routes must stay minimal and safe.
- Do not assume the app either fully boots or fully crashes; degraded mode is part of the design.

---

## 11. Configuration Architecture

Config is loaded from `config/` through `config.Load(...)`, which is built on
`pkg/server/config`.

### Current file roles

| File | Purpose | Required |
|------|---------|----------|
| `config/app.yaml` | base app runtime config | Yes |
| `config/app.development.yaml` | development overlay | No |
| `config/site.yaml` | site metadata, fonts, robots, sitemap | Yes |
| `config/nav.yaml` | navigation config | No |
| `config/service.yaml` | service/provider config | Yes |

### Rules

- Base YAML files contain structure, not secrets.
- Secrets enter through env expansion.
- Duration-valued YAML fields use integer seconds.
- Convert to `time.Duration` at the adapter boundary.
- If a config struct changes, trace every caller and every YAML mapping.

### Ownership

- app runtime config structs live in `config/`
- reusable config loading mechanics live in `pkg/server/config`

---

## 12. Feature Development Workflow

Follow the owning-domain path rather than a stale one-size-fits-all sequence.

### App-owned feature flow

1. Define or extend app-owned types in `internal/model` when the feature truly belongs to the app.
2. Decide schema ownership.
3. Add app-owned SQL and generated code if the persistence is app-owned.
4. Add or extend `internal/repository` only when the persistence contract is app-owned.
5. Add or extend `internal/service`.
6. Add handlers in `internal/handler`.
7. Add full pages and fragments in `internal/view`.
8. Register routes in `internal/app/routes.go`.
9. Add tests at the handler, service, and view boundaries.

### Package-owned feature flow

1. Confirm that the feature belongs in a reusable package.
2. Keep domain, persistence, and package-specific transport inside that package.
3. Expose a narrow package API to the app.
4. Wire the package into `internal/app`.
5. Add app-level route registration, config plumbing, and integration tests.

### Decision rule

Before writing code, answer this explicitly:

- Is the domain app-owned?
- Is the domain package-owned?

If that answer is unclear, stop and resolve ownership before implementation.

---

## 13. Operational Readiness and Test Modes

Operational readiness should be treated as a design concern, not just an ops
afterthought.

### Readiness modes

The app currently has multiple meaningful runtime modes:

- fully started
- degraded startup with availability-only routes
- sync queue mode
- async queue mode
- cookie-backed session mode
- server-backed session mode

Design changes should account for which of these modes they affect.

### Test modes

The repo also has multiple meaningful verification modes:

- pure unit tests for package and app logic
- route/middleware integration tests
- persistence-backed tests where SQL and stores are the thing being verified
- environment-dependent checks for assets, docs, or external build tooling

### Testing strategy

Test at the boundary that owns behavior.

### Handler tests

Use fakes and assert:

- status code
- redirect target
- rendered body content
- HTMX vs non-HTMX behavior where relevant

### Service tests

Use fakes and assert:

- business rules
- pagination and caps
- queue dispatch orchestration
- domain error mapping

### View tests

Render nodes and assert:

- exact content
- expected action/href values
- presence of structural IDs/attributes that matter to behavior

Remember:

- account for HTML-encoded output in assertions
- prefer exact-match assertions over substring-only assertions where feasible

### Integration tests

Use integration tests for:

- route registration behavior
- startup/degraded mode behavior
- auth/session integration boundaries
- persistence-backed package behavior where unit tests are insufficient

---

## 14. Quick Reference

### Ownership checklist

- [ ] Is this app-owned or package-owned?
- [ ] Does the schema owner match the package/code owner?
- [ ] Does the route belong in `internal/app/routes.go`?
- [ ] Is the URL generated from a route name instead of a hardcoded string?
- [ ] Is the rendering path page-only, dual-mode, or fragment-only?
- [ ] Is the behavior covered at the correct test boundary?

### Current source-of-truth map

| Concern | Current source of truth |
|--------|--------------------------|
| app assembly | `internal/app/` |
| route registration | `internal/app/routes.go` |
| app middleware stack | `internal/server/middleware.go` |
| app security composition | `internal/server/security.go` |
| error handling | `internal/server/error.go` |
| view request state and page/fragment switching | `internal/view/request.go` |
| migration schema composition | `db/sql/schemas.yaml` |
| app-owned schema | `db/sql/schema.sql` |
| package-owned auth schema and queries | `pkg/auth/pgstore/` |
| package-owned queue schema and queries | `pkg/queue/pgstore/` |

### Avoid these stale assumptions

- There is no current `internal/routes/routes.go` route-constant layer.
- `pkg/componentry` is the current component package family, not `pkg/components`.
- `internal/repository` is not currently the center of all persistence.
- Admin route grouping already exists in `internal/app/routes.go`.
- Package-owned transport and service layers are valid parts of the current architecture.
