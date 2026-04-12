---
title: Documentation site
description: HTTP handler for serving Hugo-generated documentation and a CLI tool for scaffolding and building the docs site.
weight: 20
---

# docs

`github.com/go-sum/docs` serves Hugo-generated documentation over HTTP and provides a CLI tool for scaffolding and building the documentation site. The HTTP handler mounts at `/docs` and serves pre-built static files from a `public/doc/` directory with content-type detection, cache control, and a custom 404 page. The CLI tool (`docs`) scaffolds a `.docs/` Hugo source skeleton and compiles it into the output directory. This module follows the `pkg/` leaf-node rule: it imports only the standard library, `github.com/labstack/echo/v5`, and `github.com/spf13/cobra`.

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |
| [Cobra] | v1.10 |

## Features

- Single `Handler` type that serves Hugo-built documentation pages and assets via [Echo]
- Automatic `index.html` resolution for clean URL paths (e.g., `/docs/guide` serves `guide/index.html`)
- Differentiated `Cache-Control` headers: assets cached for one hour, HTML pages served with `no-cache`
- Custom `404.html` fallback page for missing documentation routes
- Path traversal prevention via `..` rejection
- Content-type detection from file extensions with fallback to content sniffing
- CLI scaffolding command (`docs init`) that generates a complete `.docs/` Hugo site skeleton
- CLI build command (`docs build`) that invokes Hugo to compile documentation into `public/doc/`

---

## Installation

### Library (HTTP handler)

```bash
go get github.com/go-sum/docs
```

### CLI tool

```bash
go install github.com/go-sum/docs/cli@latest
```

Or run directly from a project that vendors the module:

```bash
go run ./pkg/docs/cli build
```

---

## HTTP Handler

### Types

**`Handler`** -- serves Hugo-generated documentation files from a `public/doc/` subdirectory of the configured root.

```go
type Handler struct {
    publicDir string
}
```

### Functions

**`NewHandler(publicDir string) *Handler`** -- creates a handler that serves documentation from `filepath.Join(publicDir, "doc")`. The `publicDir` argument is typically the application's top-level public directory (e.g., `"public"`), and the handler appends `"doc"` internally.

```go
docsHandler := docs.NewHandler("public")
// Serves files from public/doc/
```

### Methods

**`(h *Handler) Handle(c *echo.Context) error`** -- serves a documentation page or asset for the current request path. Registered on both `/docs` and `/docs/*` routes.

### Path Resolution Rules

The handler resolves request paths to files under the `public/doc/` root using these rules:

| Request Path | Resolved File | Classification |
|-------------|---------------|----------------|
| `/docs` | `public/doc/index.html` | HTML page |
| `/docs/` | `public/doc/index.html` | HTML page |
| `/docs/guide` | `public/doc/guide/index.html` | HTML page |
| `/docs/guide/setup` | `public/doc/guide/setup/index.html` | HTML page |
| `/docs/css/main.css` | `public/doc/css/main.css` | Asset |
| `/docs/js/theme.js` | `public/doc/js/theme.js` | Asset |

Paths with a file extension are treated as assets. Paths without an extension are treated as HTML pages and resolve to the `index.html` file within the corresponding directory.

### Cache-Control Behaviour

| Content Type | Cache-Control Header |
|-------------|---------------------|
| Assets (paths with a file extension) | `public, max-age=3600` (1 hour) |
| HTML pages (paths without a file extension) | `no-cache` |

### Custom 404 Fallback

When a requested HTML page does not exist, the handler looks for a `404.html` file at the documentation root (`public/doc/404.html`). If found, it is served with HTTP status `404` and the correct `text/html` content type. If no custom 404 page exists, the handler returns Echo's default JSON 404 response.

Missing assets (paths with a file extension) always return Echo's default JSON 404 response -- the custom 404 page is not used for asset requests.

### Path Traversal Security

Requests containing `..` anywhere in the path are rejected immediately with a 404 response. This prevents directory traversal attacks that attempt to read files outside the documentation root.

---

## CLI Tool

The `docs` CLI provides two subcommands for managing a Hugo documentation site.

```
docs
  init    Scaffold a .docs/ Hugo source directory
  build   Build Hugo documentation
```

### `docs init`

Scaffolds a barebones `.docs/` directory in the current working directory. The scaffolded directory contains Hugo layouts, CSS, JavaScript, and a starter content page ready to build.

```bash
docs init
```

**Behaviour:**

- Fails if `.docs/` already exists (prevents accidental overwrites)
- Copies the embedded template directory to `.docs/`
- Prints next-steps guidance on completion:

```
created .docs/
next steps:
  edit .docs/hugo.toml to set the title
  add markdown files under .docs/content/
  go run ./pkg/docs/cli build
```

### `docs build`

Invokes Hugo to compile the documentation source into the output directory. Removes any stale output before building.

```bash
docs build
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--source` | `string` | `.docs` | Hugo source directory |
| `--destination` | `string` | `public/doc` | Output directory for built documentation |

**Behaviour:**

- Resolves `--destination` to an absolute path
- Removes the existing destination directory (cleans stale output)
- Creates the parent directory of the destination if it does not exist
- Invokes `hugo --source <source> --destination <abs-destination> --quiet`
- Streams Hugo's stdout and stderr to the terminal

```bash
# Custom source and destination
docs build --source ./my-docs --destination ./dist/doc
```

---

## Directory Layout

### Scaffolded `.docs/` Skeleton

After running `docs init`, the following structure is created:

```
.docs/
  hugo.toml                          # Hugo configuration
  content/
    _index.md                        # Documentation home page
  layouts/
    _default/
      baseof.html                    # Base template with header, sidebar, main
      list.html                      # List page template (sections, home)
      single.html                    # Single page template (leaf pages)
    partials/
      sidebar.html                   # Navigation sidebar partial
    404.html                         # Custom not-found page
  assets/
    css/
      docs.css                       # Documentation layout styles
      theme-base.css                 # Base theme variables
      theme-slate.css                # Slate colour scheme
      chroma.css                     # Syntax highlighting styles
      chromastyles.css               # Additional Chroma token styles
    js/
      theme.js                       # Theme toggle (light/dark/system)
```

### Built Output

After running `docs build`, the compiled site appears under `public/doc/`:

```
public/
  doc/
    index.html                       # Documentation home page
    404.html                         # Custom not-found page
    css/
      main.css                       # Compiled stylesheet
    js/
      ...                            # Compiled JavaScript
    <section>/
      index.html                     # Section listing page
      <page>/
        index.html                   # Individual documentation page
```

---

## Hugo Template Features

### Sidebar Navigation

The sidebar partial (`layouts/partials/sidebar.html`) renders a two-level navigation tree automatically derived from Hugo's content structure. Top-level sections are listed by `weight`, and each section expands to show its child pages and sub-sections. The current page receives the `is-active` CSS class and `aria-current="page"` attribute.

Ordering is controlled by the `weight` front-matter parameter in each content file:

```toml
+++
title = "Getting Started"
weight = 10
+++
```

### Theme Switching

The scaffolded site includes a three-state theme toggle that cycles through light, dark, and system modes. The preference is persisted in `localStorage` under the `themePreference` key. When set to `system`, the site reacts to the operating system's `prefers-color-scheme` media query in real time.

The theme toggle button is in the page header and uses distinct SVG icons for each state (sun for light, moon for dark, monitor for system).

### Syntax Highlighting

Hugo's built-in code fence highlighting is enabled via the `hugo.toml` configuration:

```toml
[markup.highlight]
  codeFences = true
  noClasses = false
```

The `noClasses = false` setting causes Hugo to emit CSS class names rather than inline styles, which are then styled by the included `chroma.css` and `chromastyles.css` stylesheets. This ensures syntax highlighting respects the active theme.

---

## Integration Example

Register the docs handler on an [Echo] instance alongside other application routes. The handler requires two route entries -- one for the exact `/docs` path and one for the wildcard `/docs/*` path:

```go
import (
    "github.com/go-sum/docs"
    "github.com/go-sum/server/route"
)

docsHandler := docs.NewHandler(publicDir)

route.Register(e,
    route.GET("/docs", "docs.index", docsHandler.Handle),
    route.GET("/docs/*", "docs.show", docsHandler.Handle),
)
```

Both routes bind to the same `Handle` method. The `/docs` route handles the documentation root, while `/docs/*` handles all nested pages and assets.

---

## Testing

The test suite in `handler_test.go` covers:

- **Path resolution** -- verifies that `resolvePath` maps request paths to the correct file system paths, classifies assets vs. HTML pages, and rejects path traversal attempts
- **Page and asset serving** -- confirms correct HTTP status codes, response bodies, and content types for HTML pages, CSS files, JavaScript files, missing pages (custom 404), missing assets (JSON 404), and traversal attempts
- **Cache-Control headers** -- asserts that HTML pages receive `no-cache` and assets receive `public, max-age=3600`

Tests use a temporary directory populated with known files and an in-process [Echo] instance, requiring no external Hugo build or running server.

```bash
go test github.com/go-sum/docs
```

[Echo]: https://echo.labstack.com/
[Cobra]: https://cobra.dev/
