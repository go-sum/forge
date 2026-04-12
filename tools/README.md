# tools/ — Monorepo tooling (excluded from app clones)

Everything in this directory is **forge-the-monorepo** tooling. It is
stripped when forge is cloned into a new app; the rest of the repository
is **forge-the-starter**.

---

## Directory Layout

```
tools/
├── cli/
│   ├── package/    — subtree-split, push, release, status, sync for pkg/*
│   ├── workspace/  — fan-out a command across all go.work modules
│   └── starter/    — clone forge into a new app
├── Taskfile.yml        — overlay: ws:* targets + workspace fan-out overrides
├── starter/
│   └── manifest.yaml   — include/exclude/rename rules for the clone operation
└── README.md           — this file
```

---

## The Dual-Role Design

`forge/` serves two roles simultaneously:

| Role | Zone | Key files |
|---|---|---|
| **Monorepo** — develop & release `pkg/*` modules | `tools/`, `pkg/`, `go.work*`, `go.prod.mod*` | `tools/Taskfile.yml` included as `ws:` namespace |
| **Starter template** — bootstrap new Go web apps | everything else | `tools/starter/manifest.yaml` defines the boundary |

The root `Taskfile.yml` defines plain-app targets. `tools/Taskfile.yml` is
included as the `ws:` namespace and provides workspace-aware versions of
`test`, `test-race`, and `db:gen` that fan out across all `pkg/*` modules.
When a new app is cloned from forge, the `tools/` directory is absent and
the plain-app versions run unchanged.

---

## Cloning a New App

```bash
go run ./tools/cli/starter clone \
  --target ../myapp \
  --module github.com/myorg/myapp
```

What happens:
1. All paths not in `tools/starter/manifest.yaml`'s `exclude` list are copied.
2. `go.prod.mod` → `go.mod`, `go.prod.sum` → `go.sum`.
3. `github.com/go-sum/forge` is rewritten to the `--module` value throughout Go source.
4. A summary of copied files and rewritten imports is printed.

Post-clone:
```bash
cd ../myapp
go mod tidy
task db:migrate
task dev
```

### Verifying the starter is self-contained

```bash
go run ./tools/cli/starter verify
```

Clones into a temp directory, runs `go build ./cmd/server` and `go vet ./...`.
Exits non-zero on any failure. Run before every release and wired into CI.

---

## Package Management (pkg/*)

All `pkg/*` release operations are handled by `tools/cli/package`:

| Task | What it does |
|---|---|
| `task ws:list` | List all discovered `pkg/*` modules |
| `task ws:push -- auth` | Push subtree to mirror repo |
| `task ws:release -- auth` | Tag and release a package |
| `task ws:status -- auth` | Show sync status |
| `task ws:sync` | Regenerate `go.prod.mod` + `go.prod.sum` |
| `task ws:deploy` | Full release pipeline (release + push + tag) |

---

## Adding a New `pkg/*` Module

1. Create `pkg/<name>/` with its own `go.mod` (`module github.com/go-sum/<name>`).
2. Add `./pkg/<name>` to `go.work` and a `replace` + `require` pair to root `go.mod`.
3. Add the sqlc config path to the `ws:db:gen` task in `tools/Taskfile.yml`.
4. Add a `task ws:release -- <name>` call to the release sequence in
   `tools/cli/package/release.go`.
5. Add the package to `tools/starter/manifest.yaml`'s exclude list (already
   covered by the `pkg/` directory exclusion, but add explicitly if needed).
6. Document the module's public surface in `pkg/<name>/README.md`.
