---
title: Site generation
description: Reusable package for generating `robots.txt` and `sitemap.xml` content.
weight: 20
---

# Site generation

`github.com/go-sum/site` is a standalone, reusable Go module for generating `robots.txt` and `sitemap.xml` content. It provides pure generation functions with no framework dependencies in the core package, plus an optional `handlers` sub-package that serves the generated content over HTTP using [Echo].

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |

## Sub-packages

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `site` (root) | `github.com/go-sum/site` | Pure generation functions for robots.txt and sitemap.xml |
| `handlers` | `github.com/go-sum/site/handlers` | [Echo] HTTP handlers that serve `/robots.txt` and `/sitemap.xml` |

## Core Package

### `BuildRobots`

Generates a valid `robots.txt` document as a string. An empty config produces a permissive file that allows all crawlers.

```go
import "github.com/go-sum/site"

content, err := site.BuildRobots(site.RobotsConfig{
    DefaultAllow:  true,
    DisallowPaths: []string{"/admin", "/private"},
    SitemapURL:    "https://example.com/sitemap.xml",
})
// content:
// User-agent: *
// Disallow: /admin
// Disallow: /private
//
// Sitemap: https://example.com/sitemap.xml
```

#### `RobotsConfig`

| Field | Type | YAML Key | Description |
|-------|------|----------|-------------|
| `DefaultAllow` | `bool` | `default_allow` | `true` allows all crawlers with specific path exclusions. `false` emits `Disallow: /` to block all crawlers. |
| `DisallowPaths` | `[]string` | `disallow_paths` | Paths to disallow when `DefaultAllow` is `true`. Falls back to `DefaultDisallowPaths` when empty. |
| `SitemapURL` | `string` | `-` (not from YAML) | Absolute sitemap URL. When non-empty, a `Sitemap:` directive is appended. Set at handler time, not from configuration. |

#### `DefaultDisallowPaths`

When `DefaultAllow` is `true` and `DisallowPaths` is empty, the following paths are disallowed automatically:

```go
var DefaultDisallowPaths = []string{
    "/_components",
    "/users",
    "/signin",
    "/signup",
    "/signout",
    "/health",
}
```

These represent internal surfaces and authentication endpoints with no SEO value.

### `BuildSitemap`

Generates a sitemap.xml document as `[]byte`, including the `<?xml ...?>` declaration header. Returns a valid but empty `<urlset>` for nil or empty input.

```go
import "github.com/go-sum/site"

now := time.Now()
priority := 0.8

entries := []site.Entry{
    {
        Loc:        "https://example.com/",
        LastMod:    &now,
        ChangeFreq: "daily",
        Priority:   &priority,
    },
    {
        Loc:        "https://example.com/about",
        ChangeFreq: "monthly",
    },
}

data, err := site.BuildSitemap(entries)
```

#### `Entry`

| Field | Type | Description |
|-------|------|-------------|
| `Loc` | `string` | Absolute URL of the page (required). |
| `LastMod` | `*time.Time` | Optional last modification date. Formatted as `YYYY-MM-DD` in output. `nil` omits the element. |
| `ChangeFreq` | `string` | Optional change frequency hint. Valid values: `always`, `hourly`, `daily`, `weekly`, `monthly`, `yearly`, `never`. |
| `Priority` | `*float64` | Optional relative priority (`0.0`--`1.0`). `nil` omits the element (crawlers default to `0.5`). A non-nil pointer to `0.0` is emitted explicitly. |

### YAML-Driven Configuration Types

The following types support declarative sitemap configuration via YAML (deserialized with koanf tags). They are consumed by the `handlers` sub-package to resolve entries at request time. All structs carry [validator] tags for struct-level validation at config load time.

**`SitemapConfig`** -- Top-level sitemap configuration holding route entries, static page entries, and defaults for change frequency and priority.

| Field | Type | YAML Key | Validation | Description |
|-------|------|----------|------------|-------------|
| `Routes` | `[]RouteEntry` | `routes` | `omitempty,dive` | Named application routes to include in the sitemap. Each entry is validated recursively. |
| `StaticPages` | `[]StaticEntry` | `static_pages` | `omitempty,dive` | Explicit static path entries to include. Each entry is validated recursively. |
| `DefaultChangeFreq` | `string` | `default_changefreq` | `omitempty,oneof=always hourly daily weekly monthly yearly never` | Fallback change frequency for entries that do not specify one. |
| `DefaultPriority` | `float64` | `default_priority` | `omitempty,min=0,max=1` | Fallback priority (`0.0`--`1.0`) for entries that do not specify one. |

**`RouteEntry`** -- References a named [Echo] route for sitemap inclusion.

| Field | Type | YAML Key | Validation | Description |
|-------|------|----------|------------|-------------|
| `Name` | `string` | `name` | `required` | Registered [Echo] route name, e.g. `"home.show"`. |
| `ChangeFreq` | `string` | `changefreq` | `omitempty,oneof=always hourly daily weekly monthly yearly never` | Overrides `DefaultChangeFreq` for this entry. |
| `Priority` | `*float64` | `priority` | `omitempty,min=0,max=1` | Overrides `DefaultPriority` for this entry. `nil` means use default. |

**`StaticEntry`** -- References an explicit absolute path for sitemap inclusion.

| Field | Type | YAML Key | Validation | Description |
|-------|------|----------|------------|-------------|
| `Path` | `string` | `path` | `required` | Absolute path, e.g. `"/about"`. The handler prepends the origin. |
| `ChangeFreq` | `string` | `changefreq` | `omitempty,oneof=always hourly daily weekly monthly yearly never` | Overrides `DefaultChangeFreq` for this entry. |
| `Priority` | `*float64` | `priority` | `omitempty,min=0,max=1` | Overrides `DefaultPriority` for this entry. `nil` means use default. |

## HTTP Handlers

The `handlers` sub-package provides ready-to-use [Echo] handlers that serve `/robots.txt` and `/sitemap.xml`.

### `New`

```go
func New(cfg Config, routes func() echo.Routes) *Handler
```

Constructs a `Handler` from a `Config` and a lazy route accessor function. The `routes` function is evaluated at request time (not at construction), so it always reflects the fully registered route table.

### `Config`

| Field | Type | Description |
|-------|------|-------------|
| `Origin` | `string` | External canonical origin, e.g. `"https://example.com"`. Used to build absolute URLs. |
| `Robots` | `site.RobotsConfig` | Robots.txt generation settings. |
| `Sitemap` | `site.SitemapConfig` | Sitemap.xml generation settings. |

### `Handler.RobotsTxt`

Serves `robots.txt` as `text/plain`. Automatically derives the `SitemapURL` from `Config.Origin + "/sitemap.xml"`.

### `Handler.SitemapXML`

Serves `sitemap.xml` as `application/xml; charset=utf-8`. Resolves named routes from the config against the live route table, prepends `Config.Origin` to build absolute `<loc>` URLs, and applies default change frequency and priority from `SitemapConfig`.

### Cache Headers and Content Types

| Endpoint | Content-Type | Cache-Control | TTL |
|----------|-------------|---------------|-----|
| `/robots.txt` | `text/plain` | `public, max-age=86400` | 24 hours |
| `/sitemap.xml` | `application/xml; charset=utf-8` | `public, max-age=3600` | 1 hour |

## Configuration

Below is an annotated `site.yaml` snippet showing the robots and sitemap sections:

```yaml
site:
  robots:
    default_allow: true
    disallow_paths: []       # empty = uses DefaultDisallowPaths
  sitemap:
    default_changefreq: weekly
    default_priority: 0.5
    routes:
      - name: home.show
        changefreq: daily
        priority: 1.0
      - name: about.show
    static_pages:
      - path: /terms
        changefreq: yearly
        priority: 0.3
```

## Wiring Example

Register the handlers alongside your application routes:

```go
import sitehandlers "github.com/go-sum/site/handlers"

siteH := sitehandlers.New(sitehandlers.Config{
    Origin:  c.Config.App.Security.ExternalOrigin,
    Robots:  c.Config.Site.Robots,
    Sitemap: c.Config.Site.Sitemap,
}, func() echo.Routes { return c.Web.Router().Routes() })

// Route registration
route.Add(g, echo.Route{
    Method:  http.MethodGet,
    Path:    "/robots.txt",
    Name:    "robots.show",
    Handler: siteH.RobotsTxt,
})
route.Add(g, echo.Route{
    Method:  http.MethodGet,
    Path:    "/sitemap.xml",
    Name:    "sitemap.show",
    Handler: siteH.SitemapXML,
})
```

The `func() echo.Routes` parameter is a closure evaluated lazily at request time. This ensures the handler sees the complete route table even when routes are registered after the handler is constructed.

## Parameterized Route Exclusion

Routes whose resolved paths contain `:` (such as `/users/:id` or `/posts/:slug/edit`) are automatically excluded from the generated sitemap. These routes require runtime parameters to produce valid URLs and cannot be enumerated statically. Unknown route names are also silently skipped. The handler safely handles both cases during sitemap generation, so it is safe to reference parameterized or potentially missing route names in configuration -- they simply produce no output.

[Echo]: https://echo.labstack.com/
[validator]: https://github.com/go-playground/validator
