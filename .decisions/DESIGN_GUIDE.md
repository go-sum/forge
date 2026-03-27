---
title: Application Design Principles
description: Layering, feature development rules, and architectural decision guidance for Forge.
weight: 20
---

# DESIGN_GUIDE.md — Application Design Principles

> This guide prescribes **how to think and build** when using this starter.
> It covers architecture, feature development process, layer patterns, error design,
> HTMX rendering strategy, form handling, and testing.
>
> It reinforces [CLAUDE.md](../CLAUDE.md) from a process perspective — read both.
> Unless a section explicitly says otherwise, this guide describes the current
> implemented architecture of the starter, not aspirational future structure.
>
> **This guide does not document `pkg/` features.** Refer to the package source and
> its tests for the API surface.
>
> **For visual and UI design decisions,** refer to [UI_GUIDE.md](./UI_GUIDE.md).
> This guide covers architectural UI decisions (where code lives, how state flows)
> but not visual hierarchy, spacing, typography, or component selection.

---

## 1. Philosophy

This starter is designed for high-performance modern web applications that favor
server rendering, progressive enhancement, and reusable supporting packages.
`internal/` contains app-specific code; `pkg/` contains reusable packages with
strict dependency boundaries.

### The core constraint: layers only talk to the layer directly below them

Every architecture decision in this project flows from a single constraint: each layer
may only depend on the one immediately below it. This is not a style preference — it is
what makes the system testable without a database, changeable without ripple effects,
and readable without context-switching across concerns.

Violating this constraint is always locally convenient and always globally costly.
Before importing a package, ask: is this the right layer for this dependency?

### Three questions before writing code

1. **Which layer does this belong to?**
   Transport (HTTP parsing, rendering), Service (business rules), or Repository
   (data access)? If you cannot answer immediately, the code is probably in the
   wrong place.

2. **Does this already exist?**
   Check `internal/model/` for domain types, `internal/routes/` for URL patterns,
   `pkg/components/` for UI primitives. Duplication is always worse than a dependency.

3. **What is the minimum needed right now?**
   YAGNI is a hard rule here. No parameters, fields, or abstractions for hypothetical
   future callers. The right structure is what current callers require — nothing more.

---

## 2. Architecture: The Four-Layer Rule

```
Transport   internal/handler/          HTTP in, HTML/JSON out. Knows Echo.
Service     internal/service/          Business logic. Knows nothing about HTTP.
Repository  internal/repository/       Data access. Knows nothing about services.
Database    db/schema/ (generated)     sqlc output. Do not edit.
```

Each layer communicates through types defined in `internal/model/`. No layer may
import from a layer above it. No layer may skip a layer below it.

### What each layer owns

**Transport layer** — `internal/handler/`
- Parse path params, query params, and form bodies
- Validate input using the validator
- Call exactly one service method (or a small coordinated sequence)
- Render the response using `view.Render` or `render.Fragment`
- Translate service errors to HTTP responses via `apperr`
- Own nothing stateful beyond request lifetime

**Service layer** — `internal/service/`
- Own all business rules (defaults, caps, invariants)
- Orchestrate repository calls, including transactions
- Return domain types and domain errors (`model.Err*`)
- Never inspect HTTP headers, context keys, or response writers
- Accept only domain types as input — never Echo types

**Repository layer** — `internal/repository/`
- Wrap sqlc-generated queries
- Map `db.*` structs to `model.*` structs (the `toXxxModel` function pattern)
- Translate database errors to domain errors (e.g., `pgError 23505` → `model.ErrEmailTaken`)
- Expose clean interfaces; hide implementation details from callers
- Never apply business rules — that belongs to the service

**Domain model** — `internal/model/`
- Define Go structs for domain entities (`User`, `Password`)
- Define input structs with `form:` and `validate:` struct tags
- Define sentinel error variables (`ErrUserNotFound`, `ErrEmailTaken`)
- Define shared constants (`RoleUser`, `RoleAdmin`)
- Import nothing from `internal/`

### The `pkg/` boundary

Infrastructure packages (`pkg/auth`, `pkg/database`, `pkg/validate`, etc.) are
strict leaf nodes — they must not import from `internal/`. They are designed to
be extractable as standalone modules.

Component packages (`pkg/components/`) follow a tiered DAG. Imports are only
allowed downward through the tiers (Tier 0 → Tier 1 → Tier 2 → Tier 3).
See [CLAUDE.md](../CLAUDE.md) for the full tier specification.

---

## 3. Feature Development Process

Follow this sequence when adding any new feature. Each step should be completable
and reviewable independently.

### Step 1: Define the domain model

Start in `internal/model/`. Add or extend:
- Entity structs for new domain types
- Input structs for new operations (with `form:` and `validate:` tags)
- Sentinel error variables for expected failure modes
- Role or status constants if the feature introduces new named values

Do not touch any other layer yet. The model defines the contract — everything
else is implementation.

### Step 2: Design the SQL

Write SQL in `db/sql/`:
- Schema changes go in `db/sql/schema.sql`
- New query definitions go in `db/sql/queries/`
- Run `make db-plan` to preview schema changes before applying
- Run `make db-apply` to apply, then `make db-gen` to regenerate Go code

Never edit generated files in `db/schema/*.go`.

### Step 3: Implement the repository

In `internal/repository/`:
- Add the new operation to the appropriate `Repository` interface
- Implement it on the private `*repository` struct
- Write a `toXxxModel` mapper if the feature introduces a new entity
- Map database constraint errors to domain errors in the mapper

Test by examining the interface contract: does it accept and return only domain
types? Does it map all known database errors to sentinel errors?

### Step 4: Implement the service

In `internal/service/`:
- Add the new operation to the `*Service` struct
- Apply business rules (defaults, validation beyond struct tags, caps, ordering)
- Wrap repository calls in a transaction when atomicity is required (use
  `Repositories.WithTx`)
- Return domain types and domain errors — never wrap HTTP context into service
  return values

Write unit tests against a fake repository. Service tests should never touch
a real database.

### Step 5: Define the route

In `internal/routes/routes.go`:
- Add a new path constant
- Add a URL builder function if the path has dynamic segments

In `internal/server/routes.go`:
- Add the handler method to the `Handlers` interface
- Register the route in `RegisterRoutes`, in the correct active group (`public`
  or `protected`). If the route is truly admin-only, track that work against the
  planned admin route-group task in `PROJECT_PLAN.md` until the admin group is
  wired into the router.

Both files must be updated together. A route registered without a `Handlers`
method, or vice versa, will not compile.

### Step 6: Implement the handler

In `internal/handler/`:
- Add a method to `*Handler`
- Parse input (path params, form body, query string)
- Submit forms through `pkgform.New(h.validator.Validate())`
- Check `sub.IsValid()` before calling the service
- Translate service errors via `apperr.Resolve(err)` or explicit `errors.Is`
  checks for domain-specific responses
- Render using `view.Render(c, req, fullPage, region)` — pass `nil` as
  `region` only when there is no meaningful partial version

### Step 7: Build the views

In `internal/view/page/` for full-page constructors, or
`internal/view/partial/` for HTMX-swappable regions:
- Accept `view.Request` as the first parameter for pages
- Use `req.Page(title, children...)` to wrap content in the base layout
- Use `view.FormError(messages)` for top-of-form error banners
- Use component library primitives from `pkg/components/`

For visual design decisions at this step, consult [UI_GUIDE.md](./UI_GUIDE.md).

### Step 8: Write tests

- Handler tests: fake the service interfaces, test each HTTP outcome
- Service tests: fake the repository interfaces, test business rules
- Middleware tests: test the middleware function directly with `echo.New()`
- View tests: assert rendered HTML contains expected strings

Tests are complete when every distinct outcome path (validation error, domain
error, success, HTMX vs. full-page) has a corresponding test case.

---

## 4. Layer Implementation Patterns

### Handler pattern

Every handler follows the same structure:

```
1. Build req := h.request(c)
2. Parse path/query params — return apperr.BadRequest on parse failure
3. [For mutations] Bind and validate form input via pkgform
4. [If invalid] Re-render form with validation errors, status 422
5. Call exactly one service method (or a small coordinated sequence)
6. [On domain error] Return the appropriate apperr or re-render with errors
7. [On success] Render response or redirect
```

Handlers may not contain business logic. A handler that calls two unrelated
services is a signal that the service layer needs a new method that composes them.

A handler that applies a business rule (e.g., "cap per_page at 100") is a
signal that rule belongs in the service.

### Service pattern

Services own rules. Every business invariant that is not enforced by the
database schema or struct validation tags belongs here:

- Default values (`if role == "" { role = model.RoleUser }`)
- Cross-field constraints not expressible in struct tags
- Pagination arithmetic (offset = (page-1) * perPage)
- Caps and limits (perPage capped at 100)
- Transactional boundaries (create user + create password atomically)

Services must never accept or return HTTP types. The service signature is a
contract with the business domain, not with HTTP.

### Repository pattern

The repository's one job is to translate between the persistence layer and the
domain layer. The interface must be expressible in terms of domain types only.

**The `toXxxModel` mapper function** is the boundary. Everything below it is
database-coupled. Everything above it is domain-clean.

Error mapping belongs in the repository, not in the service:

```go
// In repository: translate database constraint to domain error
func mapUserErr(err error) error {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
        return model.ErrEmailTaken
    }
    return err
}
```

The service should never inspect `pgconn.PgError`. If it does, the repository
has leaked its implementation detail upward.

### Transaction pattern

When an operation must be atomic, use `Repositories.WithTx`:

```go
tx, err := s.pool.Begin(ctx)
defer tx.Rollback(ctx) // always deferred; no-op after Commit
txRepos := s.factory.WithTx(tx)
// use txRepos.User, txRepos.Password — they share the transaction
if err := tx.Commit(ctx); err != nil { ... }
```

The `defer tx.Rollback` pattern is idiomatic and safe: it is a no-op if
`Commit` already succeeded. Never omit it.

---

## 5. Error Design

### The error taxonomy

There are three error categories in this application:

| Category | Lives in | Purpose |
|---|---|---|
| Domain errors | `internal/model/errors.go` | Named sentinel errors for expected failure modes |
| Application errors | `internal/apperr/` | Transport-facing errors with HTTP status, code, and safe message |
| Infrastructure errors | returned raw | Unexpected failures (DB timeout, etc.) |

### Domain errors express expected failure modes

When a user is not found, a signin fails, or an email is already taken — these
are expected outcomes, not exceptional conditions. Name them:

```go
var (
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailTaken         = errors.New("email already in use")
    ErrInvalidCredentials = errors.New("invalid credentials")
)
```

Services return them. Handlers inspect them with `errors.Is`.

### Application errors express HTTP responses

`apperr` constructors build transport-facing errors with an HTTP status code,
a machine-readable code, a title, and a safe public message:

```go
apperr.NotFound("The requested user could not be found.")
apperr.Unauthorized("Your session is invalid. Please sign in again.")
apperr.Unavailable("Unable to load users right now.", err)
apperr.Internal(err) // always shows a generic message — never the raw error
```

Use `apperr.Resolve(err)` as the default error return in handlers. It maps
known domain errors to application errors and falls back to `Internal` for
anything unrecognised.

### Never expose infrastructure errors to users

`apperr.Internal` wraps the cause for logging but always returns the generic
message: "Something went wrong on our side. Please try again." The raw cause
is only visible in debug mode and in server logs.

### The error handler is the final safety net

The global error handler (`internal/server/error_handler.go`) provides
tri-mode dispatch:

- **HTMX partial request** → out-of-band toast fragment
- **JSON Accept header** → RFC 7807 `application/problem+json`
- **HTML request** → rendered error page

Do not attempt to replicate this logic in handlers. Return an error; let the
error handler render it.

---

## 6. HTMX Rendering Strategy

### The fundamental duality

Every user-facing route that returns HTML has two response modes:

- **Full-page response**: The base layout wraps the page content. Used on
  direct navigation, browser refresh, and first-load.
- **Partial response**: Only the changed region is returned. Used when HTMX
  drives the request.

The `view.Render(c, req, full, partial)` call handles both. When `partial` is
`nil`, the full-page component is used for both modes.

### Design regions for independent replacement

When a page contains content that HTMX will update, extract that content into
a named region function:

```go
// Full page — includes layout
func UserListPage(req view.Request, data UserListData) g.Node {
    return req.Page("Users",
        h.H1(...),
        UserListRegion(data),  // reused below
    )
}

// HTMX region — returned for partial requests
func UserListRegion(data UserListData) g.Node {
    return h.Div(h.ID("users-list-region"), ...)
}
```

The handler then passes both to `view.Render`:

```go
return view.Render(c, req, page.UserListPage(req, data), page.UserListRegion(data))
```

Name regions with the `Region` suffix. Give them a stable HTML `id`. The `id`
is the contract between the server and the HTMX `hx-target` attribute.

### HTMX attributes belong in the view, not in handlers

Never construct HTMX attributes in a handler. They belong in the component
that renders the element. The handler knows nothing about `hx-get`, `hx-target`,
or `hx-swap`.

The component knows its own swap target (`"closest tr"`) because it renders
both the element and the context it lives in. The handler only knows the URL.

### Fragment returns for HTMX-only endpoints

Some endpoints exist solely to serve HTMX swaps — they have no meaningful
full-page version. Use `render.Fragment` directly:

```go
return render.Fragment(c, userpartial.UserRow(props))
```

These are typically sub-resource endpoints: a row after editing, a form
before editing, a paginated region after navigation.

### Redirects from HTMX forms

When an HTMX form submission should navigate the browser (e.g., signin
success), use `pkg/components/patterns/redirect`. It detects whether the
request is an HTMX partial and emits either `HX-Redirect` (for HTMX) or a
standard `303 See Other` redirect automatically.

---

## 7. Form Handling

### Validation is a two-step process

**Step 1: Struct-tag validation** — runs first via `pkgform.Submission.Submit`.
This catches format errors (email format, min/max length, required fields).
Return a `422 Unprocessable Entity` response with the form re-rendered.

**Step 2: Domain validation** — runs when the service is called. Business
rules that require data (is this email already taken?) cannot be expressed in
struct tags. The handler must check for the relevant domain error and attach it
to the form via `sub.SetFieldError` or `sub.SetFormError`.

### Always re-render the form on failure

On any validation failure — whether from struct tags or domain rules — re-render
the form, not a redirect. The user should see their input preserved and the error
in context. Use `view.RenderWithStatus` with the appropriate HTTP status code:

- `422` for input validation failures (format, required fields)
- `409` for conflict errors (email already taken)
- `401` for credential failures (wrong password)

### CSRF injection

CSRF tokens are injected via two paths:
- **Form submissions**: an `<input type="hidden" name="_csrf">` with `req.CSRFToken`
- **HTMX requests**: the `hx-headers` attribute on `<body>` injects `X-CSRF-Token`
  for all HTMX requests automatically

Do not add a CSRF hidden field to forms that will be submitted exclusively via HTMX.
Do add it to standard HTML form submissions.

### Validation tag contracts

Input struct tags define the shape of valid data:

```go
type SignupInput struct {
    Email       string `form:"email"        validate:"required,email,max=255"`
    DisplayName string `form:"display_name" validate:"required,min=1,max=255"`
    Password    string `form:"password"     validate:"required,min=8"`
    Role        string `form:"role"         validate:"omitempty,oneof=user admin"`
}
```

The `form:` tag maps form field names to struct fields. The `validate:` tag
defines constraints. Struct tag validate strings cannot reference Go constants —
keep the literals in sync with the constants in `internal/model/`.

---

## 8. Route Design

### Routes are a shared contract

`internal/routes/routes.go` is the single source of truth for all URL patterns.
Both the router registration and the view layer import from it. This means:

- A URL change requires editing exactly one file
- A view that links to `/users/123/edit` cannot diverge from the route the
  router knows about

**Rule:** Never hardcode a URL path string outside of `internal/routes/`.

### Use builder functions for parameterised URLs

Route constants use Echo's parameter syntax (`:id`). View code uses builder
functions that return concrete URLs:

```go
// Route constant — for router registration
UserEdit = "/users/:id/edit"

// Builder function — for view href attributes
func UserEditPath(id string) string { return "/users/" + id + "/edit" }
```

### Group routes by authorization requirement

Routes in today's `RegisterRoutes` implementation are organized into two active
groups:

| Group | Middleware | Who can access |
|---|---|---|
| Public | `LoadSession` only | Anyone |
| `protected` | `RequireAuth` + `LoadUserContext` | Authenticated users |

Place a new route in the most restrictive implemented group appropriate for it.
Routes that need authentication belong in `protected`.

`RequireAdmin` middleware already exists, but an explicit `admin` route group is
not yet wired into `RegisterRoutes`. That work is tracked in
`PROJECT_PLAN.md`. Until it lands, do not document the router as if the admin
group already exists.

---

## 9. Testing Strategy

### Test at the boundary, not the implementation

Tests verify observable behavior, not implementation details. A handler test
should not assert that a specific service method was called — it should assert
that the response has the correct status code and contains the expected content.

### Use fakes, not mocks

Interface fakes with optional function fields are the preferred test double:

```go
type fakeUserService struct {
    listFn  func(context.Context, int, int) ([]model.User, error)
    getByID func(context.Context, uuid.UUID) (model.User, error)
    // ... one field per method
}
```

Unset function fields should return `errors.New("unexpected X call")`. This
makes test failures clearly describe which operation was unexpectedly invoked,
without requiring a mock library.

### What each layer tests

**Handler tests** (`internal/handler/*_test.go`):
- Use `newTestHandler(fakeSvc)` to build the handler with fakes
- Use `newFormContext` / `newRequestContext` to build the request
- Assert: status code, redirect location, presence of content in body
- Cover: validation failure (422), each domain error path, success

**Service tests** (`internal/service/*_test.go`):
- Inject fake repositories
- Assert: correct values passed to the repo, correct model returned
- Cover: business rule enforcement (caps, defaults, transaction commit/rollback)

**Middleware tests** (`internal/middleware/*_test.go`):
- Invoke the middleware function directly with a test context
- Assert: context values set, next handler called or not called
- Cover: unauthenticated, invalid session, HTMX vs non-HTMX redirect

**View tests** (`internal/view/*_test.go`, `internal/view/page/*_test.go`):
- Render the component to a string
- Assert presence of expected HTML content (text, attributes, structure)
- Do not test visual styling — test semantics (the `id` is there, the form
  action is correct)

### Test table format for multi-outcome scenarios

When a function has several distinct outcome paths, use table-driven tests:

```go
tests := []struct {
    name       string
    input      SomeInput
    mockSetup  func() fakeRepo
    wantErr    error
    wantResult model.User
}{...}
for _, tc := range tests {
    t.Run(tc.name, func(t *testing.T) { ... })
}
```

Name each case for its outcome, not its input: `"user not found"`, `"email
taken"`, `"success"` — not `"test case 1"`.

---

## 10. Code Style Rules

### Single Responsibility

Each function does one thing. Distinct concerns are extracted into named functions.
`main()` orchestrates — it does not implement. A function whose name cannot capture
what it does in a single verb phrase is doing too much.

### DRY

Values that appear in more than one place become constants. Logic that appears in
more than one place becomes a function. String literals used as identifiers (URL
prefixes, role names, header values) are defined once at the narrowest applicable
scope — never scattered across files.

### YAGNI

No parameters, fields, or abstractions for hypothetical future callers. No
single-field wrapper structs. The right amount of structure is what current callers
require — nothing more.

### Naming

- **Functions**: verb-first. `initLogger`, `parseLogLevel`, `toUserModel`.
- **Variables**: noun describing content. `dev`, `pool`, `opts`, `srvCfg`.
- **Constants**: descriptive noun phrase. `staticPrefix`, `cacheImmutable`, `hstsOneYear`.
- **Interfaces**: named for their capability, not their implementor. `userContextLoader`,
  not `UserServiceInterface`.
- **Constructors**: `New` for public, `new` for package-private.
  `NewUserService`, `newUserRepository`.
- **Fakes in tests**: prefix with `fake`. `fakeUserService`, `fakeAuthRepo`.

### Magic values

Any string or number that has a non-obvious meaning, or that appears more than once,
becomes a named constant scoped to the narrowest applicable level.

Role strings (`"admin"`, `"user"`) that appear in middleware, services, and views
belong in `internal/model/` as `RoleAdmin` and `RoleUser`. URL prefixes that appear
in multiple middleware calls belong as `const` in the file that uses them.

### Readability

Repeated expressions become named variables rather than inline calls:

```go
dev := cfg.IsDevelopment()  // evaluated once, used in multiple branches
```

Shared struct literals are assigned once, not duplicated across `if` branches:

```go
opts := &slog.HandlerOptions{Level: level}
// then: slog.NewTextHandler(os.Stderr, opts) or slog.NewJSONHandler(os.Stdout, opts)
```

### Consistency

New code follows established layer patterns exactly. When a pattern is insufficient,
update it everywhere — not just where you need it. A second way to do the same thing
compounds confusion with every file added.

### Comments

Comments explain what the code cannot say for itself. They are not required
everywhere — only where behavior is non-obvious, where a constraint would surprise a
reader, or where an unusual pattern needs justification.

**Add a comment when it reveals something the name or signature does not:**

- `// Port and GracefulTimeout are int (YAML integers); callers convert as needed` —
  the type alone does not reveal the conversion contract
- `// ContentFiles are loaded after env vars` — a precedence rule invisible in the
  field type
- `// Missing files are silently skipped` — a surprising absence of an error return
  warrants a note
- `// defer tx.Rollback` would look like a bug without a note — the comment
  clarifies the deliberate no-op-after-commit pattern

**Omit a comment when the name says it:**

- `// ServerConfig holds HTTP server settings` — the name says this
- `// IsDevelopment reports whether the app is in development mode` — the signature
  says this
- `// toUserModel maps a db.User to a model.User` — obvious from the function name

Name things clearly enough that a comment restating the name is unnecessary. Prefer
a single precise sentence over a verbose paragraph. Exported symbols that are
self-describing need no doc comment.

### Early returns over nesting

Validate inputs and handle errors first. The happy path should be at the
leftmost indentation level:

```go
// preferred
id, err := uuid.Parse(c.Param("id"))
if err != nil {
    return apperr.BadRequest("...")
}
user, err := h.services.User.GetByID(ctx, id)
if err != nil {
    return apperr.Resolve(err)
}
return render.Fragment(c, userpartial.UserRow(...))

// avoid
if id, err := uuid.Parse(...); err == nil {
    if user, err := h.services.User.GetByID(...); err == nil {
        return render.Fragment(...)
    } else {
        return apperr.Resolve(err)
    }
}
```

---

## 11. Extending the Architecture

### Adding a new entity

1. Define the model struct, input structs, and sentinel errors in `internal/model/`
2. Write schema SQL, run `make db-apply`, run `make db-gen`
3. Add a `XxxRepository` interface and implementation in `internal/repository/`
4. Add the repo to `Repositories` and `NewRepositories`
5. Add a `XxxService` in `internal/service/` with `NewXxxService`
6. Add the service to `Services` and `NewServices`
7. Add the service interface to `handler.go`'s handler struct and `New` constructor
8. Register routes, implement handlers, build views

### Adding a new middleware

Middleware in `internal/middleware/` may import `internal/` packages.
Define a minimal interface for the dependency (not the concrete service type):

```go
type userContextLoader interface {
    GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
}
```

This keeps the middleware testable with a simple fake, independent of the
full service struct. Register the middleware in `internal/server/routes.go`
in the appropriate group.

### Adding a new pkg/ package

Infrastructure packages in `pkg/` must satisfy the leaf-node rule: no imports
from `internal/`, no imports from other `pkg/` packages. If the package needs
application configuration, accept it as a constructor parameter — do not reach
into `config.App`.

Component packages in `pkg/components/` must respect the tier hierarchy. A new
Tier 1 component may use Tier 0 primitives but not other Tier 1 components.

### Refactoring vs. extending

Prefer refactoring an existing pattern over introducing a parallel one. If two
handlers share identical validation logic, extract it into a helper in the
handler package. If two services share a business rule, define it once in the
service or model package.

When an existing pattern is insufficient, update the pattern everywhere rather
than introducing a second way to do the same thing. Inconsistency has a
compounding cost.

---

## Quick Reference

### New feature checklist

- [ ] Domain types and error sentinels defined in `internal/model/`
- [ ] SQL written, `make db-apply` run, `make db-gen` run
- [ ] Repository interface extended, implementation added, db errors mapped
- [ ] Service method added, business rules applied, transaction used if atomic
- [ ] Route constant added to `internal/routes/`
- [ ] `Handlers` interface and `RegisterRoutes` updated
- [ ] Handler implemented: parse → validate → service → render/redirect
- [ ] Views built: page constructor and region constructor (if HTMX-replaceable)
- [ ] Tests cover: each validation outcome, each domain error, success, HTMX path

### Layer dependency map

```
cmd/server/main.go
  - internal/app  (composition root: wires all layers)
    - internal/handler    → internal/service (interfaces)
                          → internal/view
                          → internal/model
    - internal/service    → internal/repository (interfaces)
                          → internal/model
    - internal/repository → db/schema (sqlc)
                          → internal/model
    - internal/server     → internal/handler (Handlers interface)
                          → internal/middleware
                          → internal/routes
```

### Error return guide

| Situation | Return |
|---|---|
| Bad URL parameter (parse failure) | `apperr.BadRequest("...")` |
| User typed invalid input (validation) | re-render form, `422` |
| Resource does not exist | `apperr.Resolve(err)` (maps `ErrNotFound`) |
| Email/unique conflict after service call | re-render form with field error, `409` |
| Auth required, not present | `apperr.Unauthorized("...")` |
| Admin required, user is not admin | `apperr.Forbidden("...")` |
| Service or database unavailable | `apperr.Unavailable("...", err)` |
| Unexpected infrastructure failure | `apperr.Internal(err)` |
| Any unclassified service error | `apperr.Resolve(err)` |
