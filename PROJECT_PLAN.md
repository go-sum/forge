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

**Current state:** Phase 01 complete. `go.mod`, `go.sum`, database schema, sqlc-generated code, and query definitions are in place. T0106 requires a running `app_data` container — run `make db-apply` from the host (it auto-starts `app_data` via the `db` profile).

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

- [ ] **T0201** — `pkg/database` — PostgreSQL connection pool
  - Files: `pkg/database/database.go`
  - Cover: `Config{DSN string}`. `Connect(ctx, cfg) (*pgxpool.Pool, error)` — creates pool, calls `pool.Ping()`. `HealthCheck(ctx, pool) error`. `Close(pool)`. Returns descriptive errors wrapping pgx errors.

- [ ] **T0202** — `pkg/server` — Echo v5 server factory
  - Files: `pkg/server/server.go`, `pkg/server/middleware.go`
  - Cover: `Config{Host, Port, Debug, GracefulTimeout, CookieSecure, CSP, CSRFCookieName string}`. `New(cfg) *echo.Echo` — creates Echo instance and applies full middleware stack in order: `RemoveTrailingSlash` (pre-routing), `Recover`, `RequestID`, `Secure` (HSTS + CSP + X-Frame-Options + X-Content-Type-Options), `RequestLogger` (structured slog output), `CSRF` (double-submit cookie, reads `CTX_` config), `staticCacheControl` (1-year TTL for `/public/` paths with `?v=` param, no-cache otherwise). `Start(e, cfg)` — listens on `cfg.Host:cfg.Port`, handles `SIGINT`/`SIGTERM` with `echo.Shutdown(ctx)` and the configured `GracefulTimeout`.

- [ ] **T0203** — `pkg/assets` — Static file cache-busting
  - Files: `pkg/assets/assets.go`
  - Cover: `Init(staticDir, urlPrefix string, dev bool) error` — walks `staticDir`, computes 8-char SHA-256 hash of each file, stores `filename → ?v=<hash>` map. `MustInit(...)` — panics on error. `Path(name string) string` — returns `urlPrefix + name` in dev mode (no hash), `urlPrefix + name + "?v=" + hash` in production. Thread-safe after `Init`.

- [ ] **T0204** — `pkg/render` — Gomponents HTTP renderer
  - Files: `pkg/render/render.go`
  - Cover: `Component(c *echo.Context, node g.Node) error` — sets `Content-Type: text/html; charset=utf-8`, status 200, renders node to response. `ComponentWithStatus(c *echo.Context, status int, node g.Node) error`. `Fragment(c *echo.Context, node g.Node) error` — same as Component but signals HTMX partial (used by handlers to render partials without layout). All functions render directly to `c.Response()` using `node.Render(w)`.

---

### Phase 03: Security pkg/

> JWT token lifecycle and Gorilla session management. Input validation wrapper.

- [ ] **T0301** — `pkg/validate` — Input validation wrapper
  - Files: `pkg/validate/validate.go`
  - Cover: `Validator` struct wrapping `*validator.Validate`. `New() *Validator` — registers custom tag name func to use `form` tag before `json` tag for field name lookup. `Validate(s any) ValidationErrors` where `ValidationErrors` is `map[string]string`. Returns nil if valid. `(e ValidationErrors) ForField(name string) string`. `(e ValidationErrors) HasField(name string) bool`. `(e ValidationErrors) IsEmpty() bool`.

- [ ] **T0302** — `pkg/auth` — JWT + session management
  - Files: `pkg/auth/jwt.go`, `pkg/auth/session.go`
  - Cover: **JWT** — `JWTConfig{Secret, Issuer string, TokenDuration time.Duration}`. `Claims{UserID uuid.UUID, Email, Role string, jwt.RegisteredClaims}`. `GenerateToken(cfg, userID, email, role) (string, error)`. `ValidateToken(cfg, tokenString) (*Claims, error)`. **Session** — `SessionConfig{Name, AuthKey, EncryptKey string, MaxAge int, Secure bool}`. `SessionManager` wrapping `gorilla/sessions.Store`. `NewSessionStore(cfg) *SessionManager`. `SetUserID(w, r, userID string) error`. `GetUserID(r) (string, error)`. `SetFlash(w, r, key, value string) error`. `GetFlashes(w, r, key string) ([]string, error)`. `Clear(w, r) error`. Cookies: `HttpOnly`, `SameSite=Lax`, `Secure` from config.

---

### Phase 04: UX/Request Helpers pkg/

> Request introspection and response utilities for HTMX, flash messages, pagination,
> and redirect handling. Draws heavily from `examples/pagoga/pkg/`.

- [ ] **T0401** — `pkg/ctxkeys` — Type-safe context key constants
  - Files: `pkg/ctxkeys/ctxkeys.go`
  - Cover: Unexported type `type ctxKey string`. Exported constants: `UserID`, `UserEmail`, `UserRole`, `IsAuthenticated`, `RequestID`, `Logger`, `CSRF`, `Config`. Using a distinct type prevents accidental collisions with third-party packages using bare string keys. Each constant is a `ctxKey` value.

- [ ] **T0402** — `pkg/htmx` — HTMX request/response helpers
  - Files: `pkg/htmx/htmx.go`
  - Cover: **Request inspection** — `IsRequest(c) bool` (checks `HX-Request: true`). `IsBoosted(c) bool` (`HX-Boosted`). `GetTrigger(c) string` (`HX-Trigger`). `GetTarget(c) string` (`HX-Target`). `GetTriggerName(c) string` (`HX-Trigger-Name`). `GetCurrentURL(c) string` (`HX-Current-URL`). **Response** — `SetRedirect(c, url)` (sets `HX-Redirect`). `SetRefresh(c)` (sets `HX-Refresh: true`). `SetPushURL(c, url)` (`HX-Push-Url`). `SetReplaceURL(c, url)` (`HX-Replace-Url`). `SetTrigger(c, event)` (`HX-Trigger`). `SetTriggerAfterSettle(c, event)` (`HX-Trigger-After-Settle`). `SetRetarget(c, selector)` (`HX-Retarget`). `SetReswap(c, strategy)` (`HX-Reswap`).

- [ ] **T0403** — `pkg/flash` — Flash message system
  - Files: `pkg/flash/flash.go`
  - Cover: `Type` string type with constants `TypeSuccess`, `TypeInfo`, `TypeWarning`, `TypeError`. `Message{Type, Text string}`. `Set(w http.ResponseWriter, r *http.Request, store sessions.Store, t Type, text string) error` — stores in session. `GetAll(w, r, store) ([]Message, error)` — reads and clears all flash messages. Convenience helpers: `Success(...)`, `Info(...)`, `Warning(...)`, `Error(...)`. Errors are logged but not bubbled (flash is non-critical).

- [ ] **T0404** — `pkg/pager` — Pagination helper
  - Files: `pkg/pager/pager.go`
  - Cover: `Pager{Page, PerPage, TotalItems, TotalPages int}`. `New(r *http.Request, defaultPerPage int) Pager` — reads `?page` and `?per_page` query params, clamps page ≥ 1. `(p *Pager) SetItems(total int)` — calculates `TotalPages`. `(p Pager) GetOffset() int` — returns `(Page-1) * PerPage`. `IsBeginning() bool`, `IsEnd() bool`, `HasPrev() bool`, `HasNext() bool`. `PrevPage() int`, `NextPage() int`.

- [ ] **T0405** — `pkg/redirect` — Smart HTTP/HTMX redirect builder
  - Files: `pkg/redirect/redirect.go`
  - Cover: `New(c *echo.Context) *Builder`. `(b *Builder) To(url string) *Builder`. `(b *Builder) StatusCode(code int) *Builder` (default 303). `(b *Builder) Go() error` — detects HTMX request: if `HX-Request` and not `HX-Boosted`, sets `HX-Redirect` header + returns `204 No Content`; if `HX-Boosted` or not HTMX, uses standard `c.Redirect()`.

---

### Phase 05: Form Handling pkg/

> Form submission interface with binding, validation, and per-field error tracking.
> Inspired by `examples/pagoga/pkg/form/`.

- [ ] **T0501** — `pkg/form` — Form submission interface + implementation
  - Files: `pkg/form/form.go`, `pkg/form/submission.go`
  - Cover: **`Form` interface** — `Submit(c *echo.Context, dest any) error`. `IsSubmitted() bool`. `IsValid() bool`. `IsDone() bool` (submitted AND valid). `FieldHasErrors(field string) bool`. `GetFieldErrors(field string) []string`. `SetFieldError(field, msg string)`. `GetErrors() map[string][]string`. **`Submission` struct** — implements `Form`. `Submit` binds `c.Request()` into `dest` using `echo.BindBody`, then calls validator. Populates `errors map[string][]string` from `validate.ValidationErrors`. `New() *Submission` factory.

---

### Phase 06: UI Components pkg/

> Reusable gomponents primitives that views compose into pages. Tailwind v4 utility classes.
> All components accept `g.Node` children for composability.

- [ ] **T0601** — `pkg/ui` — Common UI components
  - Files: `pkg/ui/ui.go`, `pkg/ui/button.go`, `pkg/ui/alert.go`, `pkg/ui/badge.go`, `pkg/ui/field.go`, `pkg/ui/table.go`, `pkg/ui/card.go`
  - Cover:
    - **button.go**: `Button(variant, label string, attrs ...g.Node) g.Node`. Variants: `primary` (blue), `secondary` (gray), `danger` (red), `ghost` (transparent). Applies Tailwind classes per variant.
    - **alert.go**: `Alert(t flash.Type, text string) g.Node` — dismissible alert with Alpine.js `x-data / x-show` pattern. Styled per flash type (green/blue/yellow/red).
    - **badge.go**: `Badge(label, color string) g.Node` — small inline status pill.
    - **field.go**: `Field(label, name, inputType, value, errMsg string, attrs ...g.Node) g.Node` — wraps `<label>` + `<input>` + error message `<p>`. Adds red border class when `errMsg` is non-empty.
    - **table.go**: `Table(headers []string, rows []g.Node) g.Node` — responsive table with `<thead>` + `<tbody>`. `TableRow(cells []g.Node) g.Node`. `TableCell(content g.Node) g.Node`. `TableActions(nodes ...g.Node) g.Node` (right-aligned last column).
    - **card.go**: `Card(title string, children ...g.Node) g.Node` — rounded card with shadow, optional title in header.

---

### Phase 07: Domain Layer

> Koanf-based configuration loading and domain model definitions.
> The domain types are the shared language between layers.

- [ ] **T0701** — `internal/config` — Koanf configuration
  - Files: `internal/config/config.go`, `internal/config/loader.go`
  - Cover: **`Config` struct** with nested `App{Env, Name string}`, `Server{Host, Port, GracefulTimeout, CookieSecure bool, CSP, CSRFCookieName string}`, `Database{URL string}`, `Auth{JWT{Secret, Issuer string, TokenDuration time.Duration}, Session{Name, AuthKey, EncryptKey string, MaxAge int, Secure bool}}`, `Log{Level string}`, `Site{Title, Description, LogoPath, FaviconPath, MetaKeywords, OGImage string}`. All fields tagged with `validate:` rules (required, min length, etc.). **`loader.go`** — `Init() error` uses koanf: load `config/config.yaml` → `config/config.{env}.yaml` (silently skip if missing) → `config/site.yaml` → `CTX_`-prefixed env vars (strip prefix, lowercase, `_` → `.`). Unmarshal into global `App *Config`. Call `validate.New().Validate(App)`. **Helper methods on `*Config`**: `IsDevelopment() bool`, `IsProduction() bool`, `DSN() string`, `Addr() string`.

- [ ] **T0702** — `internal/model` — Domain models
  - Files: `internal/model/user.go`, `internal/model/errors.go`
  - Cover: **`user.go`** — `User{ID uuid.UUID, Email, DisplayName, Role string, CreatedAt, UpdatedAt time.Time}` (no `PasswordHash` — enforced security boundary). `CreateUserInput{Email, DisplayName, Password string}` with `form:` and `validate:` tags (email required+email format, display_name min=2 max=255, password required min=8 max=72). `UpdateUserInput{Email, DisplayName, Role string}` — all `omitempty` (empty string = do not change). `LoginInput{Email, Password string}` — both required. **`errors.go`** — `var ErrUserNotFound`, `ErrEmailTaken`, `ErrInvalidCredentials`, `ErrForbidden error` as sentinel errors using `errors.New`.

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

> Type-safe HTML using gomponents. No `html/template`. Components are Go functions
> returning `g.Node` — compile-time safe, no template parsing overhead.

- [ ] **T1101** — `internal/view/layout` — Base page layout + navbar
  - Files: `internal/view/layout/base.go`, `internal/view/layout/navbar.go`
  - Cover: **`base.go`** — `Props{Title, CSRFToken string, IsAuthenticated bool, UserName string, Flash []flash.Message, Children []g.Node}`. `Page(p Props) g.Node` — full HTML5 document: `<head>` with charset, viewport, title, `assets.Path("css/app.css")`, CSRF meta tag. `<body>` with `hx-headers` attribute injecting CSRF token for all HTMX requests, disables HTMX runtime style injection (`<meta name="htmx-config" content='{"inlineStyleNonce":""}'>`). Renders `Navbar(p)`, flash alerts loop, `<main>` with children, scripts: `assets.Path("js/htmx.min.js")`, `assets.Path("js/alpine.min.js")` (defer order: app.js before alpine). **`navbar.go`** — `Navbar(p Props) g.Node` — nav bar with site title/logo, auth-conditional links (Login/Register vs Username/Logout).

- [ ] **T1102** — `internal/view/page/home` — Home page
  - Files: `internal/view/page/home.go`
  - Cover: `HomeProps{IsAuthenticated bool, UserName, CSRFToken string}`. `Home(p HomeProps) g.Node` — wraps `layout.Page`. Hero section with headline + subheadline. CTA: if authenticated show link to `/users`; if not show links to `/login` and `/register`. Uses `pkg/ui` Button component.

- [ ] **T1103** — `internal/view/page/auth` — Auth pages
  - Files: `internal/view/page/auth.go`
  - Cover: `LoginProps{Form *form.Submission, CSRFToken string, ErrorMsg string}`. `LoginPage(p) g.Node` — card layout, `<form method="POST" action="/login">`, CSRF hidden input, email+password fields via `ui.Field`, error display, Submit button. `RegisterProps{Form *form.Submission, CSRFToken string}`. `RegisterPage(p) g.Node` — similar layout with email, display name, password fields. Per-field error display from `p.Form.GetFieldErrors("email")` etc.

- [ ] **T1104** — `internal/view/page/users` — User list page
  - Files: `internal/view/page/users.go`
  - Cover: `UserListProps{Users []model.User, Pager pager.Pager, CSRFToken string, IsAuthenticated bool, UserName string}`. `UserListPage(p) g.Node` — wraps `layout.Page`. `ui.Table` with headers [Name, Email, Role, Joined, Actions]. Loops users → `partial.UserRow(rowProps)`. Pagination controls (Prev/Next) reading `pager.HasPrev()` / `pager.HasNext()`. HTMX target container with id for out-of-band swaps.

- [ ] **T1105** — `internal/view/partial/user_form` — Inline edit form
  - Files: `internal/view/partial/user_form.go`
  - Cover: `UserFormProps{User model.User, CSRFToken string, Errors validate.ValidationErrors}`. `UserEditForm(p) g.Node` — `<form hx-put="/users/{id}" hx-target="closest tr" hx-swap="outerHTML">` with CSRF hidden field. Email, display name, role fields using `ui.Field`. Per-field error display. Save button (primary), Cancel button (`hx-get="/users/{id}/row"` to restore original row).

- [ ] **T1106** — `internal/view/partial/user_row` — User table row
  - Files: `internal/view/partial/user_row.go`
  - Cover: `UserRowProps{User model.User, CSRFToken string}`. `UserRow(p) g.Node` — `<tr id="user-{id}">` with cells for display name, email, role badge, formatted CreatedAt. Actions cell: Edit button (`hx-get="/users/{id}/edit"`, swaps row with form), Delete button (`hx-delete="/users/{id}"`, Alpine.js `x-on:click` confirm dialog, `hx-target="closest tr"`, `hx-swap="outerHTML swap:500ms"`).

---

### Phase 12: Server Entry Point

> Wire all layers together into a running application. This is the composition root —
> the only place where concrete types are instantiated and injected.

- [ ] **T1201** — `cmd/server/logging` — Structured logger setup
  - Files: `cmd/server/logging.go`
  - Cover: `SetupLogger(debug bool) *slog.Logger`. Debug mode: `slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})`. Production: `slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})`. Calls `slog.SetDefault(logger)`. Returns logger.

- [ ] **T1202** — `cmd/server/main` — Application bootstrap
  - Files: `cmd/server/main.go`
  - Cover: Ordered initialization sequence:
    1. `config.Init()` — load config (panics on failure)
    2. `SetupLogger(cfg.Log.Debug)` — structured logging
    3. `assets.MustInit("./public", "/public", cfg.IsDevelopment())` — hash static files
    4. `database.Connect(ctx, database.Config{DSN: cfg.DSN()})` — pgxpool
    5. `auth.NewSessionStore(cfg.Auth.Session)` — Gorilla session store
    6. `validate.New()` — validator instance
    7. `repository.NewRepositories(pool)` — data access layer
    8. `service.NewServices(repos, cfg.Auth.JWT, sessions)` — business logic layer
    9. `handler.New(services, sessions, validator, pool, cfg.Server.CSRFCookieName)` — transport layer
    10. `server.New(cfg.Server)` — Echo instance with middleware
    11. `h.RegisterRoutes(e, sessions)` — mount routes
    12. `server.Start(e, cfg.Server)` — blocking listen + graceful shutdown on signal

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
