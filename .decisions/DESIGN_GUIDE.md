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

See [3](#3-domain-ownership-decision-framework) for the decision framework that governs which pattern to use.

---

## 3. Domain Ownership Decision Framework

Every feature in Forge is either **package-owned** or **app-owned**. This
distinction drives where code lives, who owns the schema, how types flow
through the system, and how the feature is tested.

### 3.1 The two ownership models

**Package-owned** — the domain lives in `pkg/<name>/` as a self-contained,
extractable module. The package owns its schema, queries, generated code, model
types, error sentinels, store interfaces, and any package-specific service or
transport code. The host app integrates the package through its public API and
wires it in `internal/app`.

**App-owned** — the domain lives in `internal/`. The app owns its model types,
error sentinels, persistence, service orchestration, handlers, and views.

### 3.2 Decision criteria

Answer these questions in order. The first "yes" determines ownership.

| # | Question | If yes → |
|---|----------|----------|
| 1 | Could this domain deploy cleanly in a different Go application with zero `internal/` dependencies? | **Package-owned** |
| 2 | Does this domain define a schema contract that other apps would consume identically? | **Package-owned** |
| 3 | Does the domain already exist as a `pkg/` module? | **Package-owned** — extend it rather than duplicating in `internal/` |
| 4 | Is this domain specific to this application's product logic, policies, or workflows? | **App-owned** |
| 5 | Does the domain exist only to orchestrate or compose package-owned capabilities? | **App-owned** |

If none of the above resolves clearly, default to **app-owned** and extract
later if reuse demand materializes. Premature extraction into `pkg/` is harder
to undo than a later promotion from `internal/`.

### 3.3 What each owner controls

| Concern | Package-owned | App-owned |
|---------|--------------|-----------|
| Domain types (structs, enums) | `pkg/<name>/model/` | `internal/model/` |
| Error sentinels | `pkg/<name>/model/` | `internal/model/` |
| SQL schema | `pkg/<name>/pgstore/sql/schema.sql` | `db/sql/schema.sql` |
| SQL queries | `pkg/<name>/pgstore/sql/queries.sql` | `db/sql/queries/*.sql` |
| sqlc config + generated code | `pkg/<name>/pgstore/.sqlc.yaml` → `pkg/<name>/pgstore/db/` | `.sqlc.yaml` → `db/schema/` |
| Store interfaces | `pkg/<name>/repository/` | `internal/repository/` |
| Store implementation | `pkg/<name>/pgstore/` | `internal/repository/` |
| Business logic | `pkg/<name>/service/` (if any) | `internal/service/` |
| HTTP handlers | `pkg/<name>/` (if package exposes transport) | `internal/handler/` |
| Views / UI | n/a (app always owns page composition) | `internal/view/` |
| Route registration | wired by `internal/app/routes.go` | wired by `internal/app/routes.go` |

### 3.4 Type boundaries between `pkg/` and `internal/`

When the app consumes a package-owned domain, there are two valid patterns for
how the package's types flow through app layers. Choose based on the criteria
below — neither is the universal default.

#### Pattern A: Direct use of package types

App layers (handlers, services, views) import and use the package's model types
directly. No app-layer wrapper or conversion function exists.

```go
// internal/service/user.go
import authmodel "github.com/go-sum/auth/model"

func (s *UserService) List(ctx context.Context, ...) ([]authmodel.User, error)
```

**Use when:**

- the app adds no fields, computed properties, or semantic differences to the package type
- the package type is stable and the app is unlikely to diverge from it
- the domain is fundamentally owned by the package — the app is a consumer, not a co-owner

**Current example:** user management — `internal/handler`, `internal/service`,
and `internal/view` all use `authmodel.User` directly.

#### Pattern B: App-owned wrapper type with boundary conversion

The app defines its own domain type in `internal/model/` and maps at the
boundary where package types enter the app.

```go
// internal/model/
type Project struct { ... }   // app-owned, not a copy of any package type

// internal/service/
func toProject(p somepkg.RawProject) model.Project { ... }
```

**Use when:**

- the app needs fields, computed values, or semantics the package type does not provide
- the domain has app-specific invariants or policies that should be expressed in the type
- the app is a co-owner of the domain's shape — it may diverge from the package over time
- multiple packages contribute data to the same app-level concept

**Practical example:** if the app ever adds `LastLoginAt`, `Preferences`, or
`Subscription` fields to its notion of a user, those belong in an app-owned
type — not in `pkg/auth/model.User`.

#### When to convert from A to B

Start with Pattern A. Introduce Pattern B when the app first needs a field or
semantic that does not belong in the package. Do not pre-create wrapper types
in anticipation of future divergence.

### 3.5 Error sentinel ownership

Error sentinels follow the same ownership rule as domain types:

- package-owned domains define their sentinels in `pkg/<name>/model/`
- app-owned domains define their sentinels in `internal/model/`

When using Pattern A (direct package types), app code imports package error
sentinels directly. Do not re-export package sentinels in `internal/model/`
unless the app needs to add a distinct error that does not belong in the
package.

**Current example:** `authmodel.ErrUserNotFound` is used directly in handlers.
`model.ErrAdminExists` is app-owned because admin elevation policy belongs to
the app, not to `pkg/auth`.

### 3.6 Interface ownership for cross-boundary consumption

When the app depends on a package store, the interface is defined in the
package's `repository/` sub-package. The app should not redefine the interface
in `internal/`.

The composition root in `internal/app` may define a **combined interface** when
it needs to expose multiple package-defined interfaces through a single
container field:

```go
// internal/app/container.go
type AuthStore interface {
    authrepo.UserStore   // auth flows
    authrepo.AdminStore  // admin CRUD
}
```

This combined interface is wiring glue, not domain ownership. The individual
interfaces remain owned by the package.

### 3.7 Anti-patterns

- **Mirror schemas.** Do not maintain a copy of a package-owned table in
  `db/sql/sqlc_schema.sql` for root code generation. If the app needs to query
  a package-owned table, add the query to the package.
- **Re-exported sentinels.** Do not alias package errors in `internal/model/`
  (`var ErrUserNotFound = authmodel.ErrUserNotFound`). Import them directly.
- **Duplicate mappers.** Do not write a second `toUserModel()` in
  `internal/repository` that converts the same DB row the package already
  converts. Extend the package's store interface instead.
- **Premature wrapper types.** Do not create `internal/model.User` as a
  field-for-field copy of `authmodel.User` with only a conversion function.
  Wait until the app actually needs different fields or semantics.

---

## 4. Composition Root

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
- `bootstrap_infra.go`: config, logger, assets, sender, web/Echo initialization
- `bootstrap_data.go`: database, auth store, queue, KV store initialization
- `bootstrap_services.go`: session, validator, and domain services initialization
- `container.go`: shared dependency container and background-service lifecycle
- `routes.go`: route group definitions and orchestration (delegates to feature files)
- `routes_public.go`: unauthenticated read-only routes (home, docs, health, site files)
- `routes_auth.go`: auth flows, session management, account routes, contact POST
- `routes_admin.go`: admin-only user management routes with ETag caching
- `docs.go`: docs static-file handler for `/docs`

Auth-domain adapters live in `internal/adapters/auth/` (package `authadapter`):

- `adapters.go`: session, form, flash, and redirect adapters for `pkg/auth` interfaces
- `notifier.go`: email verification notifier
- `view.go`: auth page renderer using componentry

### Rules

- Do not make feature packages reach up into the composition root.
- Do not hide runtime registration behind implicit magic.
- New background workers, route groups, and package integrations must be wired explicitly here.
- If a package needs app configuration, pass it in from `internal/app`; do not make the package read `config.App` directly.

---

## 5. Layer Rules Inside `internal/`

The application-specific part of the repo still follows a layered model, but
it is narrower than older docs implied.

```
Transport   internal/handler/         App-owned HTTP handlers
Service     internal/service/         App-owned business orchestration
Repository  internal/repository/      App-owned persistence helpers
View        internal/view/            Request state, pages, partials, layouts
Model       internal/model/           App-owned domain errors and shared types
```

### Feature-oriented file naming convention

Within each flat package, one file per feature domain. File names must match the
feature they serve: `user.go`, `contact.go`, `admin.go`, `account.go`, etc.

| Layer | File naming | Example |
|-------|------------|---------|
| `internal/handler/` | `{feature}.go` | `user.go`, `contact.go` |
| `internal/service/` | `{feature}.go` | `user.go`, `contact.go` |
| `internal/view/page/` | `{feature}.go` | `users.go`, `contact.go` |
| `internal/view/partial/` | `{feature}partial/` | `userpartial/`, `contactpartial/` |

**Split triggers** — introduce a second file for a feature when:
- The file grows beyond ~250 lines: split into `{feature}.go` + `{feature}_helpers.go`
- The feature needs private types that could collide with other features

**Subpackage trigger** — introduce a subpackage (e.g., `handler/billing/`) only when
a feature has 3+ files or requires package-private types that cannot be shared.

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

`internal/model` owns:

- app-owned sentinel errors (e.g. `ErrAdminExists`, `ErrForbidden`)
- app-owned domain structs (e.g. `ContactInput`)

For package-owned domains, the app uses the package's model types directly
(Pattern A in [3.4](#34-type-boundaries-between-pkg-and-internal)) unless the app needs its own fields or semantics (Pattern
B). See [3](#3-domain-ownership-decision-framework) for the decision criteria.

### View layer — `internal/view/`

`internal/view` owns:

- request-scoped presentation state via `view.Request`
- full-page constructors
- partial constructors
- shared layout composition
- request-aware route reversal in the view layer

It is the app-owned presentation layer above `pkg/componentry`.

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

## 7. Persistence Ownership Model

Forge uses distributed schema ownership.

### Canonical ownership

| Domain | Canonical schema file | Owner |
|-------|------------------------|-------|
| app-owned shared DB objects | `db/sql/schema.sql` | `internal/` |
| users | `pkg/auth/pgstore/sql/schema.sql` | `pkg/auth` |
| queue jobs | `pkg/queue/pgstore/sql/schema.sql` | `pkg/queue` |

`db/sql/schemas.yaml` is the composition registry used for migration diffing.

### Query ownership

Each package-owned store has its own sqlc config and generated code:

- `pkg/auth/pgstore/.sqlc.yaml` → `pkg/auth/pgstore/db/`
- `pkg/queue/pgstore/.sqlc.yaml` → `pkg/queue/pgstore/db/`

A root sqlc config (`.sqlc.yaml`) is reserved for future app-owned tables. Currently no root queries exist.

### Rules

- App-owned tables belong in `db/sql/schema.sql`.
- Package-owned tables belong beside the owning package.
- App-owned query code belongs in root sqlc inputs and outputs when the domain is internal.
- Package-owned query code belongs inside the owning package and must not cross module boundaries.
- Every new table or query surface must declare its owner before implementation: `internal` or a specific package.
- Do not add runtime schema-install hooks in stores. Schema changes flow through migrations and `db/sql/schemas.yaml`.

### Choosing between `internal/repository` and `pkg/*/pgstore`

Apply the decision criteria in [3.2](#32-decision-criteria). In short: if the domain is extractable
and reusable, its persistence belongs in `pkg/*/pgstore`. If the domain is
app-specific, its persistence belongs in `internal/repository`.

---

## 8. Routing Model

Route registration currently lives in `internal/app/routes.go`.

There is no separate `internal/routes/routes.go` source of truth at present.
The durable contract is instead:

- routes are registered with names in `internal/app/routes.go`
- links and redirects are generated by reversing those names against the live route table

### Route file structure

Route registration is split across feature files within `internal/app/`:

- `routes.go` — group definitions, middleware wiring, orchestration
- `routes_public.go` — unauthenticated read-only pages, health, docs, site files
- `routes_auth.go` — auth flows, session management, account routes
- `routes_admin.go` — admin-only user management with ETag caching

### Current route groups

| Group | Middleware stack | Use for |
|-------|-----------------|---------|
| `public` (global `c.Web`) | (global middleware only) | Unauthenticated read-only pages |
| `publicPost` | CrossOriginGuard + RateLimit("auth") | Unauthenticated mutations (signin, signup, contact) |
| `authGuarded` | RateLimit("server") + RequireAuthPath | Authenticated read pages |
| `authGuardedPost` | (authGuarded) + CrossOriginGuard | Authenticated mutations (signout, account changes) |
| `adminGuarded` | (authGuarded) + LoadUserRole + RequireAdmin | Admin read pages |
| `adminGuardedPost` | (adminGuarded) + CrossOriginGuard | Admin mutations |

Future groups (not yet registered):

| Group | Purpose |
|-------|---------|
| `apiV1` | Bearer token auth, JSON responses, rate-limited |
| `webhookInbound` | Signature verification, no session, rate-limited |

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

## 9. Rendering Model

Forge supports multiple HTML response modes without splitting the app into
separate rendering stacks.

### Canonical rendering modes

| Mode | Handler pattern | Error behavior |
|------|----------------|----------------|
| Full page | `view.Render(c, req, fullPage, partial)` — `partial` may be `nil` | HTML error page |
| HTMX fragment | `view.Render(c, req, fullPage, partial)` — auto-detected via `HX-Request` header | OOB toast fragment |
| Fragment-only | `render.Fragment(c, node)` or `render.FragmentWithStatus(c, status, node)` | OOB toast fragment |
| JSON/problem | Automatic — error handler detects `Accept: application/json` | RFC 7807 problem JSON |
| Redirect | `redirect.New(w, r).To(url).Go()` | HTMX-aware redirect |

**Decision rule:** call `view.Render` for dual-mode endpoints (same handler serves
both full-page and HTMX requests), and `render.Fragment` for endpoints whose sole
purpose is a fragment swap. Never manually branch on HTMX headers for error
responses — the global error handler in `internal/server/error.go` routes those
automatically.

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

## 10. Security and Middleware Model

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

## 11. Background Services and Startup Degradation

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

## 12. Configuration Architecture

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

## 13. Feature Development Workflow

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

Before writing code, apply the ownership criteria in [3.2](#32-decision-criteria). If ownership is
unclear, stop and resolve it before implementation.

---

## 14. Operational Readiness and Test Modes

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

## 15. Quick Reference

### Ownership checklist

- [ ] Is this app-owned or package-owned? ([3.2](#32-decision-criteria))
- [ ] Does the schema owner match the package/code owner?
- [ ] Are package types used directly (Pattern A) or wrapped (Pattern B)? ([3.4](#34-type-boundaries-between-pkg-and-internal))
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
