# PROJECT_PLAN.md — Implementation Roadmap & Progress Tracker

> This is the living roadmap for the **starter** project.
> It tracks what to build, what's done, and what's next.
>
> For architectural guidelines and conventions, see [`CLAUDE.md`](./CLAUDE.md).

---

## Context

**starter** is a fullstack Go web application demonstrating a modern, production-ready stack. It serves as a reusable template for future Go projects, drawing from three reference implementations:

- **`examples/starter`** — authoritative layered architecture, Echo v5, sqlc+pgschema, pkg/ boundary rules
- **`examples/pagoga`** — richer pkg/ toolkit: HTMX helpers, flash messages, pagination, redirect builder, form interface, type-safe context keys
- **`examples/aadmin`** — fine-grained permission system, audit logging patterns, Django-style form field concepts

The architecture follows a strict layered pattern: **Transport → Service → Repository → Database**, with reusable library packages in `pkg/` and application-specific code in `internal/`.

**Current state:** Phase 01 complete. Phases 02, 04, 05, 06, and 11 complete. Phase 07 partially done (T0701, T0702). Phase 09 partially done (T0904). Phase 12 partially done (T1201, T1203). T0106 requires a running `app_data` container — run `make db-apply` from the host (it auto-starts `app_data` via the `db` profile).

---

## Key Design Decisions

| Decision | Source |
|----------|--------|
| Echo v5 handler signatures (`*echo.Context`) | `API_RULES.md` |
| Layered architecture: Transport → Service → Repository → Database | `examples/starter` |
| sqlc for type-safe query codegen (no ORM) | `examples/starter` |
| pgschema for declarative schema migrations (no migration files) | replaced Atlas |
| Koanf for layered config (YAML + `CTX_` env prefix) | `examples/starter` |
| Gorilla sessions + JWT for auth | `examples/starter` |
| Gomponents for type-safe HTML (no text templates) | `examples/starter` + `examples/pagoga` |
| HTMX request/response helpers in `pkg/htmx` | `examples/pagoga` |
| Flash message system in `pkg/flash` | `examples/pagoga` |
| Pagination helper in `pkg/pager` | `examples/pagoga` |
| Smart HTTP/HTMX redirect builder in `pkg/redirect` | `examples/pagoga` |
| Form submission interface in `pkg/form` | `examples/pagoga` |
| Type-safe context keys in `pkg/ctxkeys` | `examples/pagoga` |
| `pkg/` packages are leaf nodes (no `internal/` imports) | `CLAUDE.md` |
| Test utilities in `pkg/testutil` | `examples/pagoga` |

---

## TASKS

Status markers: `[ ]` outstanding · `[x]` completed · `[~]` in progress

Task ID format: `TXXYY` where `XX` = phase, `YY` = task number within that phase.
This allows inserting new tasks within any phase without renumbering downstream task numbers.

---

### Phase 01: Foundation — Go Module & Database Layer

> Establish the Go module, database schema, sqlc-generated code, and pgschema declarative migrations.
> Everything else builds on this layer.

- [x] **T0101** — `go.mod` + dependencies
  - Files: `go.mod`
  - Cover: Declare module `starter`, Go `1.26.0`. Direct deps: `github.com/labstack/echo/v5`, `github.com/jackc/pgx/v5`, `github.com/knadh/koanf/v2` (+ yaml/env/file providers), `github.com/go-playground/validator/v10`, `github.com/golang-jwt/jwt/v5`, `github.com/gorilla/sessions`, `github.com/google/uuid`, `maragu.dev/gomponents`, `golang.org/x/crypto`. Run `go mod tidy` to produce `go.sum`.

- [x] **T0102** — Database init scripts
  - Files: `db/init/01-extensions.sql`, `db/init-plan/01-extensions.sql`, `db/init-test/01-extensions.sql`
  - Cover: Install `pgcrypto` (for `gen_random_uuid()`) and `citext` (for case-insensitive email) extensions in each database service via `docker-entrypoint-initdb.d`. Plan database isolation is handled by the separate `schema_data` Docker service (ephemeral, tmpfs).

- [x] **T0103** — Database schema
  - Files: `db/sql/schema.sql`
  - Cover: `update_updated_at()` trigger function. `users` table: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`, `email CITEXT NOT NULL UNIQUE`, `display_name VARCHAR(255) NOT NULL`, `role VARCHAR(50) NOT NULL DEFAULT 'user'`, `created_at / updated_at TIMESTAMPTZ`. Trigger on users for `updated_at`. Index on `role`. `passwords` table: `id UUID PRIMARY KEY`, `user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `hash VARCHAR(255) NOT NULL`, `created_at TIMESTAMPTZ` (append-only — no `updated_at`). Index `idx_passwords_user_id` on `passwords(user_id)`.

- [x] **T0104** — sqlc query definitions
  - Files: `db/sql/queries/user.sql`, `db/sql/queries/password.sql`
  - Cover: **user.sql** — `-- name: CreateUser :one` (INSERT email, display_name, role RETURNING *). `-- name: GetUserByID :one`. `-- name: GetUserByEmail :one`. `-- name: ListUsers :many` (ORDER BY created_at DESC LIMIT/OFFSET). `-- name: UpdateUser :one` (COALESCE/NULLIF pattern for partial update, RETURNING *). `-- name: DeleteUser :exec`. `-- name: CountUsers :one` (SELECT COUNT(*)). **password.sql** — `-- name: CreatePassword :one` (INSERT user_id, hash RETURNING *). `-- name: GetCurrentPasswordByUserID :one` (ORDER BY created_at DESC LIMIT 1). `-- name: GetCurrentPasswordByEmail :one` (JOIN users, ORDER BY created_at DESC LIMIT 1 — used for login). `-- name: ListPasswordsByUserID :many` (full history, newest first).

- [x] **T0105** — sqlc code generation
  - Files: `db/schema/db.go`, `db/schema/models.go`, `db/schema/user.sql.go`, `db/schema/password.sql.go` (all generated — do not edit)
  - Cover: Run `sqlc generate`. Verify `pgx/v5` driver output, `uuid.UUID` type override for UUID columns, `pgtype.Timestamptz` → `time.Time` override. `db.Queries` struct with `WithTx()`. `User` struct has no `PasswordHash` field. New `Password` struct: `ID`, `UserID`, `Hash`, `CreatedAt`. `password.sql.go` contains typed functions for all 4 password queries.

- [~] **T0106** — Initial schema apply
  - Files: n/a (pgschema is declarative — no migration files)
  - Cover: With `app_data` running, `make db-apply` computes the diff between `db/sql/schema.sql` and the empty `starter` database and applies it. Verify with `make db-dump` to confirm the schema matches.

---

### Phase 02: Core Infrastructure pkg/

> Reusable leaf packages for database connectivity, Echo server setup, static asset
> versioning, and gomponents rendering. No `internal/` imports anywhere in `pkg/`.

- [x] **T0201** — `pkg/database` — PostgreSQL connection pool
  - Files: `pkg/database/database.go`
  - Cover: `Config{DSN string}`. `Connect(ctx, cfg) (*pgxpool.Pool, error)` — creates pool, calls `pool.Ping()`. `HealthCheck(ctx, pool) error`. `Close(pool)`. Returns descriptive errors wrapping pgx errors.

- [x] **T0202** — `pkg/server` — Echo v5 server factory
  - Files: `pkg/server/server.go`, `pkg/server/middleware.go`
  - Cover: `Config{Host, Port, Debug, GracefulTimeout, CookieSecure, CSP, CSRFCookieName string}`. `New(cfg) *echo.Echo` — creates Echo instance and applies full middleware stack in order: `RemoveTrailingSlash` (pre-routing), `Recover`, `RequestID`, `Secure` (HSTS + CSP + X-Frame-Options + X-Content-Type-Options), `RequestLogger` (structured slog output), `CSRF` (double-submit cookie, reads `CTX_` config), `staticCacheControl` (1-year TTL for `/public/` paths with `?v=` param, no-cache otherwise). `Start(e, cfg)` — listens on `cfg.Host:cfg.Port`, handles `SIGINT`/`SIGTERM` with `echo.Shutdown(ctx)` and the configured `GracefulTimeout`.

- [x] **T0203** — `pkg/assets` — Static file cache-busting
  - Files: `pkg/assets/assets.go`
  - Cover: `Init(staticDir, urlPrefix string, dev bool) error` — walks `staticDir`, computes 8-char SHA-256 hash of each file, stores `filename → ?v=<hash>` map. `MustInit(...)` — panics on error. `Path(name string) string` — returns `urlPrefix + name` in dev mode (no hash), `urlPrefix + name + "?v=" + hash` in production. Thread-safe after `Init`.

- [x] **T0204** — `pkg/render` — Gomponents HTTP renderer
  - Files: `pkg/render/render.go`
  - Cover: `Component(c *echo.Context, node g.Node) error` — sets `Content-Type: text/html; charset=utf-8`, status 200, renders node to response. `ComponentWithStatus(c *echo.Context, status int, node g.Node) error`. `Fragment(c *echo.Context, node g.Node) error` — same as Component but signals HTMX partial (used by handlers to render partials without layout). All functions render directly to `c.Response()` using `node.Render(w)`.

---

### Phase 03: Security pkg/

> JWT token lifecycle and Gorilla session management. Input validation wrapper.

- [x] **T0301** — `pkg/validate` — Input validation wrapper
  - Files: `pkg/validate/validate.go`
  - Cover: `Validator` struct wrapping `*validator.Validate`. `New() *Validator`. `Struct(s any) error`. `Var(field any, tag string) error`.

- [ ] **T0302** — `pkg/auth` — JWT + session management
  - Files: `pkg/auth/jwt.go`, `pkg/auth/session.go`
  - Cover: **JWT** — `JWTConfig{Secret, Issuer string, TokenDuration time.Duration}`. `Claims{UserID uuid.UUID, Email, Role string, jwt.RegisteredClaims}`. `GenerateToken(cfg, userID, email, role) (string, error)`. `ValidateToken(cfg, tokenString) (*Claims, error)`. **Session** — `SessionConfig{Name, AuthKey, EncryptKey string, MaxAge int, Secure bool}`. `SessionManager` wrapping `gorilla/sessions.Store`. `NewSessionStore(cfg) *SessionManager`. `SetUserID(w, r, userID string) error`. `GetUserID(r) (string, error)`. `SetFlash(w, r, key, value string) error`. `GetFlashes(w, r, key string) ([]string, error)`. `Clear(w, r) error`. Cookies: `HttpOnly`, `SameSite=Lax`, `Secure` from config.

---

### Phase 04: UX/Request Helpers pkg/

> Request introspection and response utilities for HTMX, flash messages, pagination,
> and redirect handling. Pure UX behaviors — no HTML rendering. All are leaf nodes.

- [x] **T0401** — `pkg/ctxkeys` — Type-safe context key constants
  - Files: `pkg/ctxkeys/ctxkeys.go`
  - Cover: Unexported `type ctxKey string`. Exported constants: `UserID`, `UserEmail`, `UserRole`, `IsAuthenticated`, `RequestID`, `Logger`, `CSRF`, `Config`.

- [x] **T0402** — `pkg/htmx` — HTMX request/response helpers
  - Files: `pkg/htmx/htmx.go`
  - Cover: **Request** — `IsRequest`, `IsBoosted`, `GetTrigger`, `GetTarget`, `GetTriggerName`, `GetCurrentURL`. **Response** — `SetRedirect`, `SetRefresh`, `SetPushURL`, `SetReplaceURL`, `SetTrigger`, `SetTriggerAfterSettle`, `SetRetarget`, `SetReswap`.

- [x] **T0403** — `pkg/flash` — Flash message system
  - Files: `pkg/flash/flash.go`
  - Cover: `Type` string constants `TypeSuccess = "success"`, `TypeInfo = "info"`, `TypeWarning = "warning"`, `TypeError = "error"`. Values are aligned to daisyUI modifier suffixes for zero-cost type conversion in layout. `Message{Type Type, Text string}`. `Set`, `GetAll` (reads and clears cookie), `Success`, `Info`, `Warning`, `Error`. Errors returned to callers.

- [x] **T0404** — `pkg/pager` — Pagination helper
  - Files: `pkg/pager/pager.go`
  - Cover: `Pager{Page, PerPage, TotalItems, TotalPages int}`. `New(r, defaultPerPage) Pager`. `SetTotal(total int)`. `Offset() int`. `IsFirst()`, `IsLast()`, `PrevPage()`, `NextPage()`.

- [x] **T0405** — `pkg/redirect` — Smart HTTP/HTMX redirect builder
  - Files: `pkg/redirect/redirect.go`
  - Cover: `New(c) *Builder`. `To(url) *Builder`. `StatusCode(code) *Builder` (default 303). `Go() error` — inlines `HX-Request`/`HX-Boosted` header reads (no `pkg/htmx` import) to preserve leaf-node status.

---

### Phase 05: Form Handling pkg/

> Form submission interface with binding, validation, and per-field error tracking.
> `pkg/ui/form` components accept `[]string` (plain slices) — callers extract errors
> from Submission before passing to UI components, maintaining the leaf-node boundary.

- [x] **T0501** — `pkg/form` — Form submission interface + implementation
  - Files: `pkg/form/form.go`, `pkg/form/submission.go`
  - Cover: **`Form` interface** — `IsSubmitted()`, `IsValid()`, `IsDone()`, `FieldHasErrors(field)`, `GetFieldErrors(field) []string`, `SetFieldError(field, msg)`, `GetErrors() map[string][]string`. **`Submission` struct** — `New(v *validate.Validator) *Submission`. `Submit(c, dest)` binds via Echo's `Bind`, runs `validate.Validator.Struct`, populates per-field error map from `validator.ValidationErrors`.

---

### Phase 06: Composable UI Primitives pkg/

> daisyUI v5 + Tailwind CSS v4 + Gomponents. Organized as sub-packages under `pkg/ui/`.
> Each sub-package is independently importable and imports only stdlib + gomponents.
> Components accept plain Go types (`string`, `[]string`, `bool`, `[]g.Node`) — never domain types.

- [x] **T0601** — Tooling: daisyUI prebuilt CSS
  - Files: `Makefile`, `Dockerfile`, `static/css/tailwind.css`, `CLAUDE.md`
  - Cover: `DAISYUI_VERSION := 5.0.14`. `make css` downloads `daisyui.min.css` via curl before Tailwind compile. Dockerfile `assets` stage downloads it alongside htmx/alpine. `tailwind.css` adds `@source "../../pkg/ui/**/*.go"`. `CLAUDE.md` technology stack table updated.

- [x] **T0602** — `pkg/ui/core` — Button
  - Files: `pkg/ui/core/button.go`
  - Cover: `ButtonVariant` (`VariantPrimary`, `VariantSecondary`, `VariantError`, `VariantGhost`, `VariantNeutral`). `ButtonSize` (`SizeDefault`, `SizeSm`, `SizeLg`). `ButtonProps`. `Button(p) g.Node`. `LinkButton(p, href) g.Node`.

- [x] **T0603** — `pkg/ui/form` — Form field components
  - Files: `pkg/ui/form/input.go`, `select.go`, `checkbox.go`, `textarea.go`
  - Cover: All wrap inputs in daisyUI `fieldset` / `legend` pattern. Errors render as `<p class="fieldset-label text-error">`. `components.Classes` used for conditional error class. `Input`, `Select` (with `SelectOption`), `Checkbox`, `Textarea`.

- [x] **T0604** — `pkg/ui/feedback` — Alert, Badge, Toast
  - Files: `pkg/ui/feedback/alert.go`, `badge.go`, `toast.go`
  - Cover: `AlertType` constants (`"success"`, `"info"`, `"warning"`, `"error"`) aligned to `flash.Type` values. `Alert` uses Alpine `dismissible` component. `AlertList(types, texts []string) g.Node`. `BadgeVariant` constants. `Badge`. `ToastItem`. `Toast(items, position) g.Node`.

- [x] **T0605** — `pkg/ui/data` — Table + Card
  - Files: `pkg/ui/data/table.go`, `card.go`
  - Cover: `TableProps{Headers, Rows, Zebra, ID}`. `Table`, `Row`, `Cell`, `HeaderCell`, `ActionsCell`. `CardProps{Title, Children, Compact}`. `Card`.

- [x] **T0606** — `pkg/ui/layout` — Generic shells
  - Files: `pkg/ui/layout/navbar.go`, `sidebar.go`
  - Cover: `NavbarProps{Brand, StartItems, EndItems}`. `Navbar` — daisyUI navbar-start/center/end. `SidebarProps{Nav, Content}`. `Sidebar` — daisyUI drawer with Alpine `sidebarDrawer`.

---

### Phase 07: Domain Layer

> Koanf-based configuration loading and domain model definitions.
> The domain types are the shared language between layers.

- [x] **T0701** — Config — Koanf configuration
  - Files: `config/config.go`, `pkg/config/loader.go`, `internal/server/router.go`
  - Cover: **`config/config.go`** — `Config` struct with sub-structs (`AppConfig`, `ServerConfig`, `DatabaseConfig`, `AuthConfig`, `JWTConfig`, `SessionConfig`, `LogConfig`, `SiteConfig`). All fields tagged `koanf:` + `validate:`. `var App *Config` singleton. `Init(baseDir string) error` calls `pkg/config.Load`. Helpers: `IsDevelopment()`, `IsProduction()`, `DSN()`. **`pkg/config/loader.go`** — `Options{EnvPrefix, BaseDir, EnvKey, SiteFile string}`. `Load(target any, opts Options) error` — koanf load order: base config.yaml (required) → env overlay (optional) → CTX_-prefixed env vars via smart `transformKey` → site.yaml (optional, wins over env vars for content). `transformKey` tries candidates from all-dots to all-underscores, returning the first form where `k.Exists()` is true. **`pkg/server.New()`** stripped to bare echo + error handler. **`internal/server/router.go`** — `Setup(e *echo.Echo, cfg pkgserver.Config)` wires app middleware stack.

- [~] **T0702** — `internal/model` — Domain models
  - Files: `internal/model/model.go`, `internal/model/errors.go` (errors.go pending)
  - Cover: **`model.go`** — `User{ID uuid.UUID, Email, DisplayName, Role string, CreatedAt, UpdatedAt time.Time}`. **Pending** — `CreateUserInput`, `UpdateUserInput`, `LoginInput` structs with validate tags. **`errors.go`** — sentinel errors: `ErrUserNotFound`, `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrForbidden`.

---

### Phase 08: Repository Layer

> Data access layer wrapping sqlc-generated code. Maps `db.*` structs to domain `model.*`
> types. Translates database errors into domain sentinel errors.

- [ ] **T0801** — `internal/repository` — Interfaces + container
  - Files: `internal/repository/repository.go`
  - Cover: `UserRepository` interface: `Create(ctx, email, displayName, role string) (model.User, error)`. `GetByID(ctx, uuid.UUID) (model.User, error)`. `GetByEmail(ctx, string) (model.User, error)`. `List(ctx, limit, offset int32) ([]model.User, error)`. `Update(ctx, id uuid.UUID, email, displayName, role string) (model.User, error)`. `Delete(ctx, uuid.UUID) error`. `Count(ctx) (int64, error)`. `PasswordRepository` interface: `Create(ctx, userID uuid.UUID, hash string) (model.Password, error)`. `GetCurrentByUserID(ctx, uuid.UUID) (model.Password, error)`. `GetCurrentByEmail(ctx, string) (model.Password, error)`. `ListByUserID(ctx, uuid.UUID) ([]model.Password, error)`. `Repositories{User UserRepository, Password PasswordRepository}` container. `NewRepositories(pool *pgxpool.Pool) *Repositories`.

- [ ] **T0802** — `internal/repository/user` — Repository implementations
  - Files: `internal/repository/user.go`, `internal/repository/password.go`
  - Cover: **user.go** — `userRepository{q *db.Queries}` implementing `UserRepository`. `toModel(u db.User) model.User` — maps all fields (no `PasswordHash`). `Create` passes email, display_name, role (3 params). Error mapping: `pgx.ErrNoRows` → `model.ErrUserNotFound`. PostgreSQL error code `23505` (unique_violation) → `model.ErrEmailTaken`. **password.go** — `passwordRepository{q *db.Queries}` implementing `PasswordRepository`. `toModel(p db.Password) model.Password`. Wraps `db.CreatePassword`, `db.GetCurrentPasswordByUserID`, `db.GetCurrentPasswordByEmail`, `db.ListPasswordsByUserID`. `pgx.ErrNoRows` → appropriate domain error.

---

### Phase 09: Service Layer

> Business logic that orchestrates repository operations and applies domain rules.
> Never imports `internal/handler` or any HTTP package.

- [ ] **T0901** — `internal/service` — Services container
  - Files: `internal/service/service.go`
  - Cover: `Services{Auth *AuthService, User *UserService}`. `NewServices(repos *repository.Repositories, pool *pgxpool.Pool, jwtCfg auth.JWTConfig, sessions *auth.SessionManager) *Services`. `repos` now includes `Password PasswordRepository`; `pool` is passed so `AuthService` can open transactions.

- [ ] **T0902** — `internal/service/auth` — Authentication service
  - Files: `internal/service/auth.go`
  - Cover: `AuthService{userRepo repository.UserRepository, passwordRepo repository.PasswordRepository, pool *pgxpool.Pool, jwt auth.JWTConfig}`. `Register(ctx, input model.CreateUserInput) (model.User, error)` — opens `pgxpool` transaction; within it: hash password with `bcrypt.GenerateFromPassword(cost=10)`, call `userRepo.CreateWithTx(ctx, tx, ...)` then `passwordRepo.CreateWithTx(ctx, tx, userID, hash)`. Commit on success, rollback on error. Propagate `model.ErrEmailTaken` unchanged. Repositories accept a `db.DBTX` interface (consistent with `db.New(dbtx)`), exposing `WithTx` variants. `Login(ctx, input model.LoginInput) (model.User, string, error)` — call `passwordRepo.GetCurrentByEmail`, on any error return `model.ErrInvalidCredentials` (no user enumeration). Compare hash with `bcrypt.CompareHashAndPassword`. Fetch user via `userRepo.GetByEmail`. On success, call `auth.GenerateToken`, return `(user, token, nil)`.

- [ ] **T0903** — `internal/service/user` — User management service
  - Files: `internal/service/user.go`
  - Cover: `UserService{repo repository.UserRepository}`. `List(ctx, page, perPage int) ([]model.User, error)` — cap `perPage` at 100, compute offset, call `repo.List`. `GetByID(ctx, id uuid.UUID) (model.User, error)`. `Update(ctx, id uuid.UUID, input model.UpdateUserInput) (model.User, error)` — pass input fields directly (empty string = COALESCE skips update in SQL). `Delete(ctx, id uuid.UUID) error`. `Count(ctx) (int64, error)`.

- [x] **T0904** — `internal/services` — Application service container
  - Files: `internal/services/container.go`, `cmd/server/main.go`
  - Cover: Mirrors `examples/pagoga/pkg/services/container.go` — a single container that owns and initialises **all** application services (infrastructure and domain alike), so `main()` reduces to a clean three-step composition root: create → register routes → start. **`Container` struct** — `Config *config.Config`, `DB *pgxpool.Pool`, `Assets *assets.Assets`, `Web *echo.Echo`, `ServerConfig pkgserver.Config`. **`NewContainer() *Container`** — panics on fatal startup failure (non-recoverable at init time); initialises in dependency order: (1) `config.Init("config")`, (2) structured logger setup (text/stderr in dev, JSON/stdout in prod — absorbs the `initLogger` helper currently in `cmd/server/main.go`), (3) `assets.MustInit(staticDir, staticPrefix, cfg.IsDevelopment())`, (4) `database.Connect(ctx, cfg.DSN())`, (5) build `pkgserver.Config` from cfg, (6) `pkgserver.New(serverCfg)` + `internalserver.Setup(e, serverCfg)`. **`Shutdown() error`** — `database.Close(pool)`. **`cmd/server/main.go`** — refactored to mirror pagoda's entry point: `c := services.NewContainer()`, `defer c.Shutdown()`, register static files + health check on `c.Web`, `pkgserver.Start(c.Web, c.ServerConfig)`. Domain layers (`Repos`, `Services`, `Handler`) are added to the Container struct in T1202; `main.go` gains no new orchestration logic at that point — only the container grows.

---

### Phase 10: Transport Layer

> HTTP handlers and middleware. Depends on services, sessions, and validator.
> Handlers never import repository. HTMX-aware throughout.

- [ ] **T1001** — `internal/middleware` — Auth middleware
  - Files: `internal/middleware/auth.go`
  - Cover: `RequireAuth(sessions *auth.SessionManager) echo.MiddlewareFunc`. Calls `sessions.GetUserID(r)`. If empty or error: HTMX request → `htmx.SetRedirect(c, "/login")` + return `401`; regular request → `c.Redirect(303, "/login")`. If valid: store `userID` in context via `ctxkeys.UserID`. Call `next(c)`.

- [ ] **T1002** — `internal/handler` — Handler struct
  - Files: `internal/handler/handler.go`
  - Cover: `Handler{services *service.Services, sessions *auth.SessionManager, validator *validate.Validator, pool *pgxpool.Pool, csrfFieldName string}`. `New(services, sessions, validator, pool, csrfFieldName) *Handler`. Helper method `csrfToken(c) string` reads CSRF token from context set by Echo middleware.

- [ ] **T1003** — `internal/handler/routes` — Route registration
  - Files: `internal/handler/routes.go`
  - Cover: `(h *Handler) RegisterRoutes(e *echo.Echo, sessions *auth.SessionManager)`. Public: `GET /` → `h.Home`. `GET /health` → `h.HealthCheck`. `GET /login` → `h.LoginPage`. `POST /login` → `h.Login`. `GET /register` → `h.RegisterPage`. `POST /register` → `h.Register`. Protected group with `custommw.RequireAuth(sessions)`: `POST /logout` → `h.Logout`. `GET /users` → `h.UserList`. `GET /users/:id/edit` → `h.UserEditForm`. `GET /users/:id/row` → `h.UserRow`. `PUT /users/:id` → `h.UserUpdate`. `DELETE /users/:id` → `h.UserDelete`. Static files: `e.Static("/public", "./public")`.

- [ ] **T1004** — `internal/handler/home` — Home + health handlers
  - Files: `internal/handler/home.go`
  - Cover: `(h *Handler) Home(c *echo.Context) error` — reads session for `userID`, resolves auth state, renders `view/page.Home(props)` via `render.Component`. `(h *Handler) HealthCheck(c *echo.Context) error` — calls `database.HealthCheck(ctx, pool)`, returns `200 {"status":"ok"}` or `503 {"status":"error"}`.

- [ ] **T1005** — `internal/handler/auth` — Auth handlers
  - Files: `internal/handler/auth.go`
  - Cover: `LoginPage(c)` — renders login form. `Login(c)` — bind `model.LoginInput` → validate → `services.Auth.Login` → on success `sessions.SetUserID` → `redirect.New(c).To("/").Go()`. On failure render form with error. `RegisterPage(c)` — renders register form. `Register(c)` — bind `model.CreateUserInput` → validate → `services.Auth.Register` → on success redirect to login with flash. On `ErrEmailTaken` return 409 with form error. `Logout(c)` — `sessions.Clear` → redirect to `/`. HTMX-aware redirects throughout.

- [ ] **T1006** — `internal/handler/user` — User CRUD handlers
  - Files: `internal/handler/user.go`
  - Cover: `UserList(c)` — `pager.New(c.Request(), 20)` → `services.User.Count` → `pager.SetItems` → `services.User.List` → `render.Component(view/page.UserList(props))`. `UserEditForm(c)` — parse `:id` → `services.User.GetByID` → `render.Fragment(view/partial.UserEditForm(props))`. `UserRow(c)` — parse `:id` → `services.User.GetByID` → `render.Fragment(view/partial.UserRow(props))`. `UserUpdate(c)` — bind `model.UpdateUserInput` → validate → `services.User.Update` → render updated row partial. `UserDelete(c)` — parse `:id` → `services.User.Delete` → `c.NoContent(200)`.

---

### Phase 11: View Layer

> Type-safe HTML using gomponents + daisyUI. No `html/template`. Components are Go
> functions returning `g.Node`. All pages compose `pkg/ui/*` primitives.

- [x] **T1101** — Alpine.js registry expansion
  - Files: `static/js/app.js`
  - Cover: Added `confirmDelete` (window.confirm for delete buttons), `sidebarDrawer` (isOpen/toggle/close for drawer), `dropdown` (isOpen/toggle/close for future navbar menus) alongside existing `dismissible`.

- [x] **T1102** — `internal/view/layout` — Base HTML shell + app navbar
  - Files: `internal/view/layout/base.go`, `internal/view/layout/navbar.go`
  - Cover: **`base.go`** — `Props{Title, CSRFToken string, IsAuthenticated bool, UserName string, Flash []flash.Message, Children []g.Node}`. `Page(p Props) g.Node` — `<html data-theme="light">`, links `daisyui.min.css` then `app.css`, `hx-headers` on `<body>` for CSRF, flash alerts via `flashAlerts()` helper, toast container div, scripts deferred (app.js → htmx → alpine). **`navbar.go`** — `AppNavbar(p Props) g.Node` — brand link, auth-conditional end items (Login/Register vs UserName/Logout form).

- [x] **T1103** — `internal/view/page/auth` — Login + Register pages
  - Files: `internal/view/page/auth.go`
  - Cover: `LoginProps{Form *pkgform.Submission, CSRFToken, ErrorMsg string}`. `LoginPage`. `RegisterProps{Form *pkgform.Submission, CSRFToken string}`. `RegisterPage`. Both use `uidata.Card` + `uiform.Input` fields with per-field error slices from `p.Form.GetFieldErrors(...)`.

- [x] **T1104** — `internal/view/page/users` — User list page
  - Files: `internal/view/page/users.go`
  - Cover: `UserListProps{Users []model.User, Pager pager.Pager, CSRFToken string, IsAuthenticated bool, UserName string, Flash []flash.Message}`. `UserListPage`. `uidata.Table` with zebra rows + HTMX-enhanced pagination (`hx-get`/`hx-target="#users-table"`/`hx-swap="outerHTML"`).

- [x] **T1105** — `internal/view/partial/userpartial` — Inline edit form
  - Files: `internal/view/partial/userpartial/user_form.go`
  - Cover: `UserFormProps{User model.User, CSRFToken string, Errors map[string][]string}`. `UserEditForm` renders `<tr>` with `hx-put` form. `uiform.Input` for email/display_name, `uiform.Select` for role. Save (primary submit) + Cancel (`hx-get` to restore row).

- [x] **T1106** — `internal/view/partial/userpartial` — User table row
  - Files: `internal/view/partial/userpartial/user_row.go`
  - Cover: `UserRowProps{User model.User, CSRFToken string}`. `UserRow` renders `<tr id="user-{id}">` with `feedback.Badge` for role, Edit button (`hx-get`/`hx-swap`), Delete button (`x-data="confirmDelete"`, `@click.prevent` Alpine confirm, `hx-delete`, `hx-swap="outerHTML swap:500ms"`). `roleVariant` maps admin→BadgePrimary, user→BadgeNeutral.

---

### Phase 12: Server Entry Point

> Wire all layers together into a running application. This is the composition root —
> the only place where concrete types are instantiated and injected.

- [x] **T1201** — `cmd/server/main` — Minimal runnable bootstrap
  - Files: `cmd/server/main.go`
  - Cover: Wires only the layers implemented so far, producing a server that starts, connects to the database, and answers requests. Verifies config + database + middleware stack end-to-end before any application logic exists.
    1. `config.Init("config")` — load layered config; `log.Fatal` on error
    2. Inline slog setup: `slog.NewTextHandler(os.Stderr, LevelDebug)` in dev, `slog.NewJSONHandler(os.Stdout, LevelInfo)` in prod; `slog.SetDefault`. (Extracted to `SetupLogger` in T1203.)
    3. `assets.MustInit("./public", "/public", cfg.IsDevelopment())` — hash static files
    4. `database.Connect(ctx, database.Config{DSN: cfg.DSN()})` — pgxpool; `log.Fatal` on error; `defer database.Close(pool)`
    5. Build `pkgserver.Config` from cfg: `Port: strconv.Itoa(cfg.Server.Port)`, `GracefulTimeout: time.Duration(cfg.Server.GracefulTimeout) * time.Second`, `CookieSecure: cfg.Auth.Session.Secure`, `Debug: cfg.IsDevelopment()`
    6. `server.New(serverCfg)` — bare Echo instance
    7. `internalserver.Setup(e, serverCfg)` — wire full middleware stack
    8. `e.Static("/public", "./public")` — serve static assets
    9. Inline `GET /health` → `{"status":"ok","env":cfg.App.Env}` (replaced by `h.HealthCheck` in T1202)
    10. `server.Start(e, serverCfg)` — blocking listen + graceful shutdown on signal

- [ ] **T1202** — `cmd/server/main` — Complete bootstrap
  - Files: `cmd/server/main.go`
  - Cover: Replaces the T1201 skeleton with the full application stack once all layers exist. Remove inline logger (use `SetupLogger`), remove inline routes (use `h.RegisterRoutes`).
    1. `config.Init("config")` — unchanged
    2. `SetupLogger(cfg.IsDevelopment())` — replaces inline slog setup
    3. `assets.MustInit(...)` — unchanged
    4. `database.Connect(...)` — unchanged
    5. `auth.NewSessionStore(cfg.Auth.Session)` — Gorilla session store
    6. `validate.New()` — validator instance
    7. `repository.NewRepositories(pool)` — data access layer
    8. `service.NewServices(repos, cfg.Auth.JWT, sessions)` — business logic layer
    9. `handler.New(services, sessions, validator, pool, cfg.Server.CSRFCookieName)` — transport layer
    10. `server.New(serverCfg)` + `internalserver.Setup(e, serverCfg)` — unchanged
    11. `h.RegisterRoutes(e, sessions)` — replaces inline routes
    12. `server.Start(e, serverCfg)` — unchanged

- [x] **T1203** — `internal/services` — Config-driven log level
  - Files: `internal/services/container.go`
  - Cover: Enhances `container.initLogger()` to read `config.Log.Level` via `parseLogLevel(s string) slog.Level` instead of inferring level from `IsDevelopment()`. Handler format remains environment-driven (text/stderr in dev, JSON/stdout in prod). `parseLogLevel` maps "debug"/"info"/"warn"/"error" to `slog.Level`; the `LogConfig.Level` validate tag (`oneof=debug info warn error`) makes the default branch unreachable. No separate `cmd/server/logging.go` file — the container paradigm established in T0904 owns all infrastructure initialization, so extracting to a standalone function would be an unnecessary indirection.

---

### Phase 13: Testing Infrastructure

> Test utilities and test suites covering all layers. Integration tests use a live
> PostgreSQL instance (same Docker stack). Unit tests use interface mocks.

- [ ] **T1301** — `pkg/testutil` — Test helpers
  - Files: `pkg/testutil/testutil.go`, `pkg/testutil/fixtures.go`
  - Cover: **`testutil.go`** — `NewContext(method, path string, body io.Reader) (*echo.Context, *httptest.ResponseRecorder)` — creates Echo context backed by `httptest.NewRecorder()`. `SetFormBody(c, values url.Values)`. `AssertStatus(t, rec, expected int)`. `AssertContains(t, rec, substr string)`. `ExecuteHandler(t, c, handler echo.HandlerFunc) *httptest.ResponseRecorder`. **`fixtures.go`** — `RandomEmail() string`. `CreateUser(t, pool, email, displayName, role string) model.User` — inserts directly via `db.CreateUser` (no password_hash). `CreatePassword(t, pool, userID uuid.UUID, hash string) model.Password` — inserts via `db.CreatePassword`; call after `CreateUser` to seed a usable credential.

- [ ] **T1302** — Handler tests
  - Files: `internal/handler/home_test.go`, `internal/handler/auth_test.go`, `internal/handler/user_test.go`
  - Cover: Table-driven tests for each handler method. Mock `*service.Services` via interface. Test HTMX path (adds `HX-Request: true` header) vs regular path. Test validation error → 422. Test auth redirect. Test successful render → check `text/html` body. Use `testutil.NewContext` and `testutil.AssertStatus`.

- [ ] **T1303** — Service tests
  - Files: `internal/service/auth_test.go`, `internal/service/user_test.go`
  - Cover: Unit tests with `mockUserRepository` implementing `repository.UserRepository`. Auth: test `Register` hashes password and propagates `ErrEmailTaken`. Test `Login` returns `ErrInvalidCredentials` for wrong email AND wrong password (same error, no enumeration). Test `Login` returns valid JWT on success. User: test `List` caps `perPage` at 100. Test `Update` passes empty strings unchanged.

- [ ] **T1304** — Repository integration tests
  - Files: `internal/repository/user_test.go`
  - Cover: `TestMain` connects to PostgreSQL via `DATABASE_URL` env var (set in `docker-compose.yml`). Truncates `users` table before each test. Tests: `TestCreateUser`, `TestGetUserByID_NotFound`, `TestGetUserByEmail_DuplicateEmail` (assert `ErrEmailTaken`), `TestListUsers_Pagination`, `TestUpdateUser_PartialFields` (empty string preserves existing value), `TestDeleteUser`. Uses `testutil.CreateUser` fixture helper.
