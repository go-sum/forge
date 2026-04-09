---
title: Patterns and Coding Principles
description: Coding standards, design patterns, and maintainability rules for Forge.
weight: 25
---

# Patterns and coding principles:

> This guide is the authoritative source for how new functions, types, packages,
> and features should be structured and maintained in Forge.
>
> It complements [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md), which defines
> architecture and ownership, and [`API_RULES.md`](./API_RULES.md), which
> defines Echo v5 transport rules.
>
> This guide is derived from:
>
> - Go best practices from Effective Go and Go Code Review Comments
> - recognized Go design-pattern guidance from Refactoring Guru
> - successful patterns already working in this codebase today

---

## 1. Purpose

Use this guide when creating or refactoring code to answer:

- how a new function should be shaped
- how dependencies and config should be owned
- when to add a pattern versus keeping code simple
- how to preserve maintainability, testability, and package boundaries

This guide is intentionally normative. When in doubt, follow the simplest
approach that keeps ownership clear and leaves the code easy to test.

---

## 2. What Already Works Well In Forge

The current codebase already demonstrates several patterns worth preserving:

- **Composition-root wiring** in `internal/app/`, where dependencies are
  assembled once and passed down explicitly.
- **Package-owned config and defaults** in packages such as `pkg/auth`,
  `pkg/session`, and `pkg/kv/redisstore`.
- **Zero-value fallback via `cmp.Or`** for comparable config fields where zero
  means "unset".
- **Factory/registry construction** in `pkg/send`, which keeps provider
  selection flexible without complicating callers.
- **Thin transport handlers** such as `internal/features/contact/handler.go`,
  where parsing, validation, service calls, and rendering remain separate.
- **Route naming and centralized registration** in `internal/app/routes.go`,
  which keeps route policy explicit and avoids path duplication.
- **Cross-cutting behavior via middleware chains**, not duplicated per handler.

The goal of this document is to make these working conventions explicit so new
code extends them instead of drifting away from them.

---

## 3. Core Programming Principles

### DRY

Reduce repetition in behavior, policies, and data mapping, but do not create a
shared abstraction until the duplication is real and stable.

### YAGNI

Be conservative. Do not add hooks, wrappers, config keys, or extension points
for hypothetical future use.

### Separation of Concerns

Split code by responsibility:

- transport concerns in handlers
- orchestration and business rules in services
- persistence in owning stores/repositories
- presentation in views
- assembly in the composition root

### SOLID, applied pragmatically

- **SRP**: each function, type, and file should have one primary reason to change.
- **ISP**: prefer narrow interfaces with a clear consumer.
- **DIP**: depend on abstractions only where that reduces coupling at a real
  boundary; do not create speculative interfaces.

### Favor composition and encapsulation

Prefer small collaborating types over inheritance-style layering or broad
utility packages. Expose the smallest public API that the next caller actually
needs.

---

## 4. Function And Type Design

When creating a new function:

- keep it focused on one job
- make the happy path obvious
- prefer early returns over nested branching
- keep side effects explicit

### Function signatures

- Pass `context.Context` as the first argument for any function that performs I/O or depends on cancellation.
- Parse and validate external input at the system boundary.
- Pass typed values, not raw strings, once input crosses the boundary.
- Return ordinary Go errors; do not use panics for expected failures.

### Error handling

- Wrap propagated errors with `%w` and enough context to identify the failing operation.
- Use `errors.Is` and `errors.As`, never string matching on error messages.
- Keep domain errors near the owning domain.
- Map domain errors to transport responses at the handler layer.

### Type design

- Prefer concrete types by default.
- Add an interface only when a consumer needs a seam for substitution or the
  package already has multiple valid implementations.
- Do not create field-for-field wrapper structs that add no semantics.
- Keep helper functions private unless another package truly needs them.

---

## 5. Package Structure And Layer Discipline

Choose the owner first by using [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md).

### App-owned code

Inside `internal/`:

- feature modules own app-specific handlers and orchestration
- `internal/repository/` owns only app-specific persistence
- `internal/view/` owns pages, partials, layout composition, and request-aware
  presentation state

### Package-owned code

Inside `pkg/`:

- each module should remain extractable
- each module should own its types, defaults, repository interfaces, and stores when it owns the domain
- packages must not reach into `internal/`

### Leaf-node rule

Treat `pkg/*` modules as leaf nodes relative to the application. They should be
portable and inward-facing, not coupled to app state or app-specific policies.

### Layer rule

Do not let a lower layer depend on a higher one:

- handlers should not own SQL
- services should not render HTML
- repositories should not decide redirects or HTTP status codes
- views should not own business rules or persistence

---

## 6. Configuration Ownership And Defaults

Where relevant, each package component should own its configuration struct and apply package defaults locally.

The standard shape is:

```go
type Config struct {
	// package-owned settings
}

var defaultConfig = Config{
	// package-owned defaults
}
```

### Rules

- The package that defines a behavior should define that behavior's `Config` type.
- The package should apply its own defaults instead of forcing callers to duplicate them.
- Application config in `config/` should compose package config and override
  only the values that are app-specific.
- Cross-field config validation should be registered through `RegisterValidationRules`.
- Config defaults should be applied at the boundary before consumers start reading fields.

### `cmp.Or` rule

Always use `cmp.Or` for zero-value fallback when all of the following are true:

- the field type is comparable
- zero means "unset"
- the fallback is local and obvious

Do not force `cmp.Or` onto slices, maps, or cases where zero is a meaningful configured value rather than "unset".

---

## 7. Dependency Construction

Dependencies should be explicit and constructor-injected.

### Rules

- Constructors should accept the dependencies they need.
- The composition root should assemble concrete implementations.
- Feature code should receive collaborators through constructors, not by reading package globals.
- Hidden global state should be the exception, not the default.

### Singletons

Use the Singleton pattern sparingly and intentionally.

It is appropriate when:

- the process should own exactly one shared registry or coordinator
- the lifetime is truly application-wide
- the singleton does not hide request-specific or feature-specific state

Current acceptable example:

- `send.DefaultRegistry`, which acts as a process-wide provider registry

Prefer a single runtime-owned instance over scattered globals when the resource
is application-specific.

---

## 8. Recommended Design Patterns

Patterns are tools, not goals. Use them where they simplify the code that
exists today.

### Factory / registry

Use a factory or registry when protocol or provider selection is a real
requirement.

This is the preferred pattern for:

- sender/provider selection
- transport or backend selection
- constructing different implementations behind one small entry point

Current example:

- `pkg/send`

### Chain of Responsibility

Use a chain when the request or operation should pass through a sequence of
orthogonal behaviors.

This is the preferred pattern for:

- middleware stacks
- request guards
- layered cross-cutting policies such as rate limits, auth checks, and origin
  protection

Current examples:

- Echo middleware composition
- route policies assembled with `route.Layout(...)`

### Adapter

Use adapters when a package-owned interface must be satisfied by app-owned
rendering, session, form, or redirect behavior.

Current examples:

- `internal/adapters/auth/`
- `internal/adapters/kvsession/`

### Resource-owner singleton

Use a single shared instance when a resource should be created once and reused,
but keep ownership explicit.

This is appropriate for:

- registries
- connection pools
- long-lived runtime-managed services

Avoid using the Singleton pattern as an excuse to hide dependencies.

---

## 9. Concurrency And Lifecycle

Apply concurrency deliberately, not casually.

### Default stance

Prefer synchronous code first. Add goroutines, workers, or async dispatch only
when there is a clear correctness, latency, or throughput reason.

### Rules

- Every goroutine must have a clear owner.
- Every long-lived background activity must have a shutdown path.
- Cancellation should flow from a passed-in context.
- Do not create fire-and-forget goroutines from handlers or services.
- If work must outlive the request, hand it to an owned background subsystem
  such as the queue.

Current good pattern:

- `Runtime.AddBackground(...)` and queue lifecycle ownership in `internal/app`

---

## 10. Testing And Refactoring Discipline

Code should be easy to test at the boundary that owns behavior.

### Rules

- Test behavior, not internal implementation details.
- Prefer fakes over mocks.
- Cover both happy paths and failure paths.
- Keep HTML assertions exact where practical, and account for HTML-encoded
  entities.
- When refactoring config structs or mappings, trace every caller and every
  mapping that depends on them.

### Refactor threshold

Refactor when code becomes:

- duplicated in a stable way
- difficult to test
- unclear about ownership
- too broad for one type or file to explain cleanly

Do not refactor just to increase abstraction count.

---

## 11. Anti-Patterns To Avoid

- speculative abstractions
- package globals used as hidden request-time dependencies
- duplicated defaults across app and package layers
- field-for-field wrapper types with no semantic change
- mirrored schemas or query code for a package-owned domain
- transport code that owns SQL or business rules
- business logic embedded in views
- unbounded goroutines or background work with no lifecycle
- hardcoded route paths where named routes already exist

If a simpler function, struct, or constructor solves the problem clearly, use
that instead of reaching for a pattern.

---

## 12. Review Checklist For New Functions And Features

Before merging a new function or feature, verify:

- ownership is clear: app-owned or package-owned
- the code lives in the correct layer
- configuration is owned by the package/component that uses it
- package defaults are local to the package
- `cmp.Or` is used for comparable zero-value fallback where appropriate
- external input is parsed and validated at the boundary
- domain errors are returned from the owning layer and mapped at the transport
  boundary
- dependencies are constructor-injected
- routes are named and reversed, not hardcoded
- concurrency has explicit ownership, cancellation, and shutdown behavior
- tests cover both primary behavior and failure cases

---

## 13. Sources

- Effective Go: <https://go.dev/doc/effective_go>
- Go Code Review Comments: <https://go.dev/wiki/CodeReviewComments>
- Refactoring Guru, design patterns in Go: <https://refactoring.guru/design-patterns/go>
