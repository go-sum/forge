---
title: UI component library
description: Gomponents-based server-rendered UI component library with tiered dependency architecture, shadcn/ui inspired styling, and HTMX support.
weight: 20
---

# UI component library

`github.com/go-sum/componentry` is a [Gomponents]-based server-rendered UI component library built on a strict tiered dependency architecture. Components are inspired by the [shadcn/ui] design system conventions using [Tailwind], render native HTML-first interactivity (no JavaScript required for most interactions), and integrate seamlessly with [HTMX] for progressive enhancement.

## Features

- Tiered dependency DAG enforcing strictly downward imports between component layers
- [shadcn/ui] inspired design system styling via [Tailwind] utility classes
- Native HTML interactivity using `<details>`, `<dialog>`, and radio button groups -- no JavaScript required for tabs, accordions, dropdowns, dialogs, popovers, or tooltips
- Full [HTMX] integration with request/response header helpers, typed attribute builders, reusable interaction patterns, and smart redirects
- Content-hash-based asset cache-busting with SVG sprite registry
- Semantic icon system decoupling icon names from concrete sprite symbols
- Form controls with built-in accessibility wiring (ARIA attributes, label association, error announcements)
- Flash messages via short-lived HTTP-only cookies with Toast rendering helpers
- Web font provider abstraction supporting [Google Fonts], [Bunny Fonts], [Adobe Fonts], and self-hosted faces
- Props-struct API for all components -- readable call sites, backward-compatible extension
- Thread-safe registries for assets and icons after initialization
- Declarative YAML-driven navigation with slot injection for dynamic content
- Table-based HTML and plain-text email construction with inline styles
- YAML-driven asset pipeline configuration

## Dependencies

| Dependency | Version |
|------------|---------|
| [Echo] | v5.0 |
| [Gomponents] | v1.2 |
| [HTMX] | v2.0 |
| [Tailwind] | v4 |
| [validator] | v10.30 |

## Tier Architecture

Imports flow strictly downward. No lateral imports within a tier, no upward imports.

| Tier | Packages | Allowed Imports |
|------|----------|-----------------|
| **0** | `render`, `testutil`, `assetconfig`, `assets`, `icons`, `email` | stdlib + external modules only |
| **1** | `ui/core`, `ui/data`, `ui/feedback`, `ui/layout`, `form`, `interactive` (theme only) | Tier 0 + stdlib/external |
| **2** | `patterns/*`, `interactive/*` (accordion, breadcrumb, dialog, dropdown, pagination, tabs, tooltip) | Tier 0 + Tier 1 + stdlib/external |
| **3** | `examples` | any tier within componentry |
| **Support** | `install`, `icons/render`, `assets/iconset` | stdlib/external only |

## Sub-packages

| Package | Import Path | Purpose |
|---------|-------------|---------|
| `assets` | `github.com/go-sum/componentry/assets` | Content-hash cache-busting and sprite registry |
| `assets/iconset` | `github.com/go-sum/componentry/assets/iconset` | Default sprite and icon catalog |
| `assetconfig` | `github.com/go-sum/componentry/assetconfig` | YAML-based asset pipeline configuration |
| `email` | `github.com/go-sum/componentry/email` | Table-based HTML and plain-text email construction |
| `examples` | `github.com/go-sum/componentry/examples` | Living component showcase page |
| `form` | `github.com/go-sum/componentry/form` | Form input controls with accessibility wiring |
| `icons` | `github.com/go-sum/componentry/icons` | Semantic icon registry |
| `icons/render` | `github.com/go-sum/componentry/icons/render` | Icon resolution to `IconProps` |
| `install` | `github.com/go-sum/componentry/install` | Registry initialization at startup |
| `interactive` | `github.com/go-sum/componentry/interactive` | Theme script and selector (Tier 1) |
| `interactive/accordion` | `github.com/go-sum/componentry/interactive/accordion` | Collapsible sections |
| `interactive/breadcrumb` | `github.com/go-sum/componentry/interactive/breadcrumb` | Breadcrumb navigation |
| `interactive/dialog` | `github.com/go-sum/componentry/interactive/dialog` | Native HTML `<dialog>` modal |
| `interactive/dropdown` | `github.com/go-sum/componentry/interactive/dropdown` | Dropdown menu |
| `interactive/pagination` | `github.com/go-sum/componentry/interactive/pagination` | Pagination controls |
| `interactive/tabs` | `github.com/go-sum/componentry/interactive/tabs` | CSS-only tab switching |
| `interactive/tooltip` | `github.com/go-sum/componentry/interactive/tooltip` | Hover and click tooltips |
| `patterns/flash` | `github.com/go-sum/componentry/patterns/flash` | Flash messages via HTTP-only cookies |
| `patterns/font` | `github.com/go-sum/componentry/patterns/font` | Web font provider abstraction |
| `patterns/form` | `github.com/go-sum/componentry/patterns/form` | Form binding and validation error tracking |
| `patterns/head` | `github.com/go-sum/componentry/patterns/head` | HTML `<head>` builder |
| `patterns/htmx` | `github.com/go-sum/componentry/patterns/htmx` | HTMX request/response helpers and interaction patterns |
| `patterns/pager` | `github.com/go-sum/componentry/patterns/pager` | Pagination state from query parameters |
| `patterns/redirect` | `github.com/go-sum/componentry/patterns/redirect` | HTMX-aware HTTP redirects |
| `render` | `github.com/go-sum/componentry/render` | Gomponents-to-HTTP response writers |
| `render/echo` | `github.com/go-sum/componentry/render/echo` | [Echo] response helpers |
| `testutil` | `github.com/go-sum/componentry/testutil` | Test helper for rendering nodes to strings |
| `ui/core` | `github.com/go-sum/componentry/ui/core` | Button, Icon, Badge, Label, Avatar, Popover, Separator, Skeleton |
| `ui/data` | `github.com/go-sum/componentry/ui/data` | Card, Table |
| `ui/feedback` | `github.com/go-sum/componentry/ui/feedback` | Alert, Toast, Progress |
| `ui/layout` | `github.com/go-sum/componentry/ui/layout` | Navbar, NavMenu, Sidebar |

---

## Tier 0 -- Infrastructure

### assets

Content-hash-based asset cache-busting and SVG sprite registry.

#### `Assets` Type

```go
func New(publicDir, prefix string) (*Assets, error)
func Must(a *Assets, err error) *Assets
func (a *Assets) Path(name string) string
```

`New` scans `publicDir` and computes SHA-256 hashes for all files. `Path` returns a versioned URL such as `/public/css/app.css?v=a1b2c3d4` (first 8 hex chars of the hash). Returns the bare fallback URL if the file is not found.

#### Package-Level API

Delegates to a package-level `Default` instance:

```go
var Default *Assets

func Init(publicDir, prefix string) error
func MustInit(publicDir, prefix string)
func Path(name string) string
```

#### Sprite Registry

```go
type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) RegisterSprite(key, rel string)
func (r *Registry) RegisterSprites(files map[string]string)
func (r *Registry) SetPathFunc(f func(string) string)
func (r *Registry) SpritePath(key string) string
```

Package-level delegates to `DefaultRegistry`:

```go
var DefaultRegistry *Registry

func PublicPath(rel string) string
func RegisterSprite(key, rel string)
func RegisterSprites(files map[string]string)
func SetPathFunc(f func(string) string)
func SpritePath(key string) string
```

```go
assets.MustInit("public", "/public")
assets.RegisterSprites(map[string]string{
    "lucide-icons": "img/svg/lucide-icons.svg",
})
assets.SetPathFunc(assets.Path)

url := assets.SpritePath("lucide-icons")
// => "/public/img/svg/lucide-icons.svg?v=abcd1234"
```

#### Sub-package `assets/iconset`

Ships a default catalog of sprite file associations and semantic icon bindings.

```go
type Catalog struct {
    Sprites map[string]string
    Icons   map[icons.Key]icons.Ref
}

var Default Catalog
```

`Default` includes `"lucide-icons"` and `"theme-icons"` sprite sets.

---

### assetconfig

YAML-based asset pipeline configuration loaded from `.assets.yaml`.

```go
const DefaultConfigPath = ".assets.yaml"

func Load(path string) (*Config, error)

type Config struct {
    Paths   Paths
    JS      JSConfig
    CSS     []CSSConfig
    Sprites map[string]SpriteConfig
    Fonts   FontConfig
}
```

#### `Paths`

| Method | Return | Description |
|--------|--------|-------------|
| `SourceRoot()` | `string` | Resolved source directory |
| `PublicRoot()` | `string` | Resolved public output directory |
| `URLPrefix()` | `string` | URL prefix for public assets |
| `PublicURL(rel string)` | `string` | Full public URL for a relative path |

#### `JSConfig`

| Field | Type | Description |
|-------|------|-------------|
| `Downloads` | `[]JSDownload` | Third-party JS files to fetch |
| `Bundles` | `[]JSBundle` | Browser bundles emitted from source entrypoints |

#### `CSSConfig`

| Field | Type | Description |
|-------|------|-------------|
| `Tool` | `string` | CSS compilation tool |
| `Input` | `string` | Source input path |
| `Output` | `string` | Compiled output path |

#### `SpriteConfig`

| Field | Type | Description |
|-------|------|-------------|
| `Enabled` | `bool` | Whether sprite generation is active |
| `Target` | `string` | Output path for the combined sprite file |
| `Sources` | `[]SourcesConfig` | Source SVG directories and file lists |

#### `FontConfig`

| Field | Type | Description |
|-------|------|-------------|
| `Downloads` | `[]FontDownload` | Font files to fetch from remote URLs |

---

### email

Table-based HTML and plain-text email construction with inline styles for email client compatibility.

```go
func Layout(title string, body g.Node) g.Node
func H1(text string) g.Node
func P(text string) g.Node
func PlainText(lines ...string) string
```

All styles are inline -- no CSS classes. Use `PlainText` for the `Text` field of outbound messages. `PlainText` joins lines with RFC 5322 CRLF separators.

```go
html := email.Layout("Welcome",
    g.Group([]g.Node{
        email.H1("Welcome to Acme"),
        email.P("Your account is ready."),
    }),
)
plain := email.PlainText(
    "Welcome to Acme",
    "",
    "Your account is ready.",
)
```

---

### icons

Semantic icon registry that decouples named icon keys from concrete sprite/symbol pairs.

#### Types

**`Key`** (`string`) -- semantic icon identifier.

| Constant | Description |
|----------|-------------|
| `ChevronDown`, `ChevronLeft`, `ChevronRight`, `ChevronsUp` | Directional arrows |
| `Close` | Close / dismiss |
| `ThemeLight`, `ThemeDark`, `ThemeSystem` | Theme switcher states |

**`Ref`** -- concrete sprite reference.

| Field | Type | Description |
|-------|------|-------------|
| `Sprite` | `string` | File key in assets registry |
| `ID` | `string` | Symbol ID within the SVG sprite |

#### Registry

```go
type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) Register(key Key, ref Ref)
func (r *Registry) RegisterSet(symbols map[Key]Ref)
func (r *Registry) Resolve(key Key) (Ref, bool)
```

Package-level delegates to `Default`:

```go
var Default *Registry

func Register(key Key, ref Ref)
func RegisterSet(symbols map[Key]Ref)
func Resolve(key Key) (Ref, bool)
```

#### Sub-package `icons/render`

Resolves semantic icon keys into `core.IconProps` for use with `ui/core.Icon`.

```go
func PropsFor(key icons.Key, p core.IconProps) core.IconProps
func PropsForRegistry(r *icons.Registry, key icons.Key, p core.IconProps) core.IconProps
func Props(spriteKey, symbolID string, p core.IconProps) core.IconProps
func PropsForAssets(r *assets.Registry, spriteKey, symbolID string, p core.IconProps) core.IconProps
func PropsForRegistries(assetReg *assets.Registry, iconReg *icons.Registry, key icons.Key, p core.IconProps) core.IconProps
```

```go
props := render.PropsFor(icons.Close, core.IconProps{Size: "size-5"})
node := core.Icon(props)
```

---

### render

Writes Gomponents nodes as HTTP responses. The root package is framework-agnostic (stdlib `net/http`).

```go
func Component(w http.ResponseWriter, node g.Node) error
func ComponentWithStatus(w http.ResponseWriter, status int, node g.Node) error
func Fragment(w http.ResponseWriter, node g.Node) error
func FragmentWithStatus(w http.ResponseWriter, status int, node g.Node) error
```

`Component` renders a full HTML page (status 200). `Fragment` renders an HTMX partial (status 200). Both `WithStatus` variants accept a custom status code.

#### Sub-package `render/echo`

Delegates to the stdlib render functions via [Echo]'s `ResponseWriter`:

```go
func Component(c *echo.Context, node g.Node) error
func ComponentWithStatus(c *echo.Context, status int, node g.Node) error
func Fragment(c *echo.Context, node g.Node) error
func FragmentWithStatus(c *echo.Context, status int, node g.Node) error
```

```go
func (h *Handler) UserList(c *echo.Context) error {
    page := userListPage(users)
    return render.Component(c, page)
}
```

---

### testutil

Test helper for rendering Gomponents nodes to strings.

```go
func RenderNode(t *testing.T, node g.Node) string
```

Renders the node to a string and calls `t.Fatal` on error.

```go
html := testutil.RenderNode(t, core.Button(core.ButtonProps{Label: "Save"}))
assert.Equal(t, `<button class="..." type="button">Save</button>`, html)
```

---

## Tier 1 -- Base UI Components

### form

Form input controls inspired by [shadcn/ui] styling with built-in accessibility wiring (ARIA attributes, label association, error announcements).

#### Input

```go
type InputType string
const TypeText, TypePassword, TypeEmail, TypeNumber, TypeTel, TypeURL, TypeSearch, TypeDate, TypeFile, TypeColor InputType = ...

type InputProps struct {
    ID, Name, Placeholder, Value string
    Type                         InputType
    Disabled, Readonly, Required, HasError bool
    Extra                        []g.Node
}

func Input(p InputProps) g.Node
```

#### Select

```go
type Option struct{ Value, Label string }
type OptGroup struct{ Label string; Options []Option }

type SelectProps struct {
    ID, Name, Selected string
    Disabled, HasError  bool
    Options             []Option
    Groups              []OptGroup
    Extra               []g.Node
}

func Select(p SelectProps) g.Node
```

#### Textarea

```go
type TextareaProps struct {
    ID, Name, Placeholder, Value string
    Rows                          int
    Disabled, Readonly, Required, HasError bool
    Extra                         []g.Node
}

func Textarea(p TextareaProps) g.Node
```

#### Checkbox, Radio, Switch, Toggle

```go
type CheckboxProps struct { ID, Name, Value string; Checked, Disabled, Required, HasError bool; Extra []g.Node }
func Checkbox(p CheckboxProps) g.Node

type RadioProps struct { ID, Name, Value string; Checked, Disabled, Required, HasError bool; Extra []g.Node }
func Radio(p RadioProps) g.Node

type SwitchProps struct { ID, Name string; Checked, Disabled, Required bool; Extra []g.Node }
func Switch(p SwitchProps) g.Node

type ToggleProps struct { ID string; Pressed, Disabled bool; Children, Extra []g.Node }
func Toggle(p ToggleProps) g.Node
```

#### FileUpload

```go
type FileUploadProps struct {
    ID, Name, Accept, Prompt string
    Multiple, Disabled       bool
    Extra                    []g.Node
}

func FileUpload(p FileUploadProps) g.Node
```

Renders a drag-and-drop upload area with a file input fallback.

#### Field

Wraps a control with label, description, hint, and error messages.

```go
type FieldProps struct {
    ID, Label, Description, Hint string
    Errors                        []string
    Control                       g.Node
    Extra                         []g.Node
}

func Field(p FieldProps) g.Node
func FieldControlAttrs(controlID, description, hint string, errors []string) []g.Node
func Description(controlID, text string) g.Node
func Hint(controlID, text string) g.Node
func ErrorMessage(controlID string, errors ...string) g.Node
```

#### FieldSet

```go
type FieldSetProps struct { ID, Legend string; Disabled bool; Extra []g.Node }

func FieldSet(p FieldSetProps, children ...g.Node) g.Node
```

```go
form.Field(form.FieldProps{
    ID:    "email",
    Label: "Email address",
    Errors: submission.GetFieldErrors("email"),
    Control: form.Input(form.InputProps{
        ID:       "email",
        Name:     "email",
        Type:     form.TypeEmail,
        Value:    input.Email,
        HasError: submission.FieldHasErrors("email"),
    }),
})
```

---

### interactive (theme -- Tier 1)

Theme switching without FOUC (flash of unstyled content).

```go
func ThemeScript() g.Node
func ThemeSelector() g.Node

var ScriptCSPHash string
```

`ThemeScript` emits a synchronous `<script>` block that reads `localStorage['themePreference']` and sets the `data-theme-preference` attribute and `.dark` class on `<html>` before any content renders. Place it at the top of `<head>`. Add `ScriptCSPHash` to your `Content-Security-Policy` `script-src` directive.

`ThemeSelector` renders a button that cycles through light, dark, and system modes using three icon spans controlled by CSS visibility rules keyed on `data-theme-preference`.

---

### ui/core

Foundational elements used throughout the library.

#### Button

```go
type Variant string
const VariantDefault, VariantDestructive, VariantDestructiveGhost, VariantOutline, VariantSecondary, VariantGhost, VariantLink Variant = ...

type Size string
const SizeDefault, SizeSm, SizeLg Size = ...

type ButtonProps struct {
    ID, Label, Type, Href, Target string
    Variant                       Variant
    Size                          Size
    Disabled, FullWidth           bool
    Children, Extra               []g.Node
}

func Button(p ButtonProps) g.Node
```

Renders a `<button>` element. When `Href` is set, renders an `<a>` styled as a button instead. `Type` defaults to `"button"` to avoid accidental form submission.

#### Icon

```go
type IconProps struct {
    Src, ID, Size, Label string
    Extra                []g.Node
}

func Icon(p IconProps) g.Node
```

Renders `<svg><use href="Src#ID"/></svg>`. `Size` defaults to `"size-4"`. When `Label` is empty, the icon is marked `aria-hidden="true"`. When `Label` is set, the icon gets `role="img"` and `aria-label`.

#### Badge

```go
type BadgeVariant string
const BadgeDefault, BadgeSecondary, BadgeDestructive, BadgeOutline BadgeVariant = ...

type BadgeProps struct { ID string; Variant BadgeVariant; Children, Extra []g.Node }

func Badge(p BadgeProps) g.Node
```

#### Label

```go
type LabelProps struct {
    For   string
    Error string
    Extra []g.Node
}

func Label(p LabelProps, children ...g.Node) g.Node
```

Renders a `<label>` element. When `Error` is non-empty, adds destructive text colour.

#### Avatar

Namespace pattern with `Image` and `Fallback` sub-components:

```go
var Avatar avatarNS

Avatar.Image(src, alt string, extra ...g.Node) g.Node
Avatar.Fallback(children ...g.Node) g.Node
```

#### Popover

CSS-first popover using native `<details>/<summary>`. No JavaScript required.

```go
var Popover popoverNS

type PopoverRootProps struct {
    ID    string
    Class string   // override default "relative inline-block"
    Extra []g.Node
}
type PopoverTriggerProps struct {
    Class string   // appended to base "list-none cursor-pointer" classes
    Extra []g.Node
}
type PopoverContentProps struct {
    Width string   // Tailwind width class; default "w-72"
    Align string   // "left" | "right" | "center"; default "left"
    Extra []g.Node
}

Popover.Root(p PopoverRootProps, children ...g.Node) g.Node
Popover.Trigger(p PopoverTriggerProps, children ...g.Node) g.Node
Popover.Content(p PopoverContentProps, children ...g.Node) g.Node
```

#### Separator

```go
type Orientation string
const OrientationHorizontal, OrientationVertical Orientation = ...

type Decoration string
const DecorationDefault, DecorationDashed, DecorationDotted Decoration = ...

type SeparatorProps struct {
    Orientation Orientation
    Decoration  Decoration
    Label       string
    Extra       []g.Node
}

func Separator(p SeparatorProps) g.Node
```

Renders a horizontal or vertical divider with an optional centred label.

#### Skeleton

```go
func Skeleton(extra ...g.Node) g.Node
```

Animated loading placeholder.

---

### ui/data

Data display components.

#### Card

```go
var Card cardNS

Card.Root(children ...g.Node) g.Node
Card.Header(children ...g.Node) g.Node
Card.Title(children ...g.Node) g.Node
Card.Description(children ...g.Node) g.Node
Card.Content(children ...g.Node) g.Node
Card.Footer(children ...g.Node) g.Node
```

#### Table

```go
var Table tableNS

type BodyProps struct {
    ID    string
    Extra []g.Node
}

type RowProps struct {
    Selected bool
    Extra    []g.Node
}

Table.Root(children ...g.Node) g.Node
Table.Header(children ...g.Node) g.Node
Table.Body(props BodyProps, children ...g.Node) g.Node
Table.Footer(children ...g.Node) g.Node
Table.Row(props RowProps, children ...g.Node) g.Node
Table.Head(children ...g.Node) g.Node
Table.Cell(children ...g.Node) g.Node
Table.Caption(children ...g.Node) g.Node
```

```go
data.Card.Root(
    data.Card.Header(
        data.Card.Title(g.Text("Users")),
    ),
    data.Card.Content(
        data.Table.Root(
            data.Table.Header(
                data.Table.Row(data.RowProps{},
                    data.Table.Head(g.Text("Name")),
                    data.Table.Head(g.Text("Email")),
                ),
            ),
            data.Table.Body(data.BodyProps{},
                data.Table.Row(data.RowProps{},
                    data.Table.Cell(g.Text("Alice")),
                    data.Table.Cell(g.Text("alice@example.com")),
                ),
            ),
        ),
    ),
)
```

---

### ui/feedback

Notifications and status indicators.

#### Alert

```go
type AlertVariant string
const AlertDefault, AlertDestructive AlertVariant = ...

type AlertProps struct {
    ID          string
    Variant     AlertVariant
    Dismissible bool
    Icon        g.Node
    Extra       []g.Node
}

var Alert alertNS

Alert.Root(p AlertProps, children ...g.Node) g.Node
Alert.Title(children ...g.Node) g.Node
Alert.Description(children ...g.Node) g.Node
Alert.List(types []string, texts []string) g.Node
```

When `Icon` is set, the layout switches to a two-column grid so the icon sits in its own column alongside the title and description. `Alert.List` maps parallel type/text slices into dismissible alerts, mapping `"destructive"` and `"error"` to `AlertDestructive` and all other types to `AlertDefault`.

#### Toast

```go
type ToastVariant string
const ToastDefault, ToastSuccess, ToastError, ToastWarning, ToastInfo ToastVariant = ...

type ToastPosition string
const PositionTopRight, PositionTopLeft, PositionTopCenter, PositionBottomRight, PositionBottomLeft, PositionBottomCenter ToastPosition = ...

type ToastProps struct {
    ID, Title, Description string
    Variant                ToastVariant
    Position               ToastPosition
    Dismissible            bool
    Extra                  []g.Node
}

func Toast(p ToastProps) g.Node
```

When `Position` is set, the toast is fixed-positioned and self-contained. When `Position` is empty (zero value), the toast renders as a plain card suitable for injection into a container div or out-of-band HTMX swap.

#### Progress

```go
type ProgressVariant string
const ProgressDefault, ProgressSuccess, ProgressDanger, ProgressWarning ProgressVariant = ...

type ProgressSize string
const ProgressSm, ProgressLg ProgressSize = ...

type ProgressProps struct {
    ID        string
    Max       int
    Value     int
    Label     string
    ShowValue bool
    Size      ProgressSize
    Variant   ProgressVariant
    Extra     []g.Node
}

func Progress(p ProgressProps) g.Node
```

`Value` and `Max` range from 0 to `Max` (defaults to 100 when zero or negative). When `Size` is omitted, the default height is used.

---

### ui/layout

Page structure components.

#### Navbar

Responsive navigation bar with dropdowns on desktop and a sidebar drawer on mobile.

**Item types** -- `NavbarItem` is a sealed interface. Only these types implement it:

| Type | Description |
|------|-------------|
| `NavLink` | Anchor link with optional icon and prefix matching |
| `NavGroup` | Dropdown group containing nested items |
| `NavSeparator` | Visual divider |
| `NavText` | Static text with optional icon |
| `NavNode` | Custom [Gomponents] nodes (separate desktop/mobile variants) |
| `NavForm` | Inline form with hidden fields (e.g., sign-out button) |

All item types carry a `Visibility` field (`VisibilityAll`, `VisibilityGuest`, `VisibilityUser`) that controls rendering based on authentication state.

```go
type NavbarBrand struct {
    Label    string
    Href     string
    LogoPath string
}

type NavbarSection struct {
    Label string
    Align NavbarSectionAlign
    Items []NavbarItem
}

type NavbarProps struct {
    ID              string
    Brand           NavbarBrand
    Sections        []NavbarSection
    IsAuthenticated bool
    CurrentPath     string
}

func Navbar(p NavbarProps) g.Node
```

#### NavMenu

High-level declarative navigation driven by YAML configuration with support for named slots.

```go
type NavConfig struct {
    Brand    NavbarBrand  `koanf:"brand"`
    Sections []NavSection `koanf:"sections"`
}

type NavSection struct {
    Label string             `koanf:"label"`
    Align NavbarSectionAlign `koanf:"align"`
    Items []NavItem          `koanf:"items"`
}

type NavItem struct {
    Type         string           `koanf:"type"`
    Slot         string           `koanf:"slot"`
    Visibility   NavbarVisibility `koanf:"visibility"`
    Label        string           `koanf:"label"`
    Href         string           `koanf:"href"`
    Action       string           `koanf:"action"`
    Method       string           `koanf:"method"`
    Icon         string           `koanf:"icon"`
    MatchPrefix  bool             `koanf:"match_prefix"`
    HiddenFields []NavHiddenField `koanf:"hidden_fields"`
    Items        []NavItem        `koanf:"items"`
}

type NavMenuProps struct {
    ID              string
    Config          NavConfig
    Slots           NavSlots
    CurrentPath     string
    IsAuthenticated bool
}

func NavMenu(p NavMenuProps) g.Node
```

`NavMenu` converts declarative `NavConfig` items into the lower-level `NavbarItem` types and renders through `Navbar`. Named `Slot` fields in config items are resolved against the `Slots` map, enabling dynamic content injection.

##### Slot Helpers

```go
type NavSlots map[string]NavSlot

func TextSlot(text string) NavSlot
func ControlSlot(label string, control g.Node) NavSlot
func FormSlot(p FormSlotProps) NavSlot
func RegisterNavValidations(v *validator.Validate)
```

`TextSlot` renders non-interactive text. `ControlSlot` renders a control directly on desktop and wraps it in a labelled mobile row. `FormSlot` renders a form action using shared `NavForm` styling. `RegisterNavValidations` registers the declarative nav schema validation rules on a [validator] instance.

#### Sidebar

CSS-only off-canvas drawer controlled by a hidden checkbox and labels.

```go
type SidebarProps struct {
    ID      string
    Nav     g.Node
    Content []g.Node
}

func Sidebar(p SidebarProps) g.Node
func ToggleAttrs(id string) []g.Node
func CloseAttrs(id string) []g.Node
```

`ToggleAttrs` returns the attributes a `<label>` should carry to toggle the sidebar. `CloseAttrs` returns the attributes to close the sidebar.

---

## Tier 2 -- Patterns and Interactive

### interactive/accordion

CSS-first collapsible sections using `<details>/<summary>`. No JavaScript required.

```go
func Root(children ...g.Node) g.Node
func Item(children ...g.Node) g.Node
func Trigger(children ...g.Node) g.Node
func Content(children ...g.Node) g.Node
```

```go
accordion.Root(
    accordion.Item(
        accordion.Trigger(g.Text("Section 1")),
        accordion.Content(g.Text("Content for section 1")),
    ),
    accordion.Item(
        accordion.Trigger(g.Text("Section 2")),
        accordion.Content(g.Text("Content for section 2")),
    ),
)
```

---

### interactive/breadcrumb

```go
func Root(children ...g.Node) g.Node
func List(children ...g.Node) g.Node
func Item(children ...g.Node) g.Node
func Link(href string, children ...g.Node) g.Node
func Page(children ...g.Node) g.Node
func Separator() g.Node
```

`Page` marks the current page with `aria-current="page"`.

```go
breadcrumb.Root(
    breadcrumb.List(
        breadcrumb.Item(breadcrumb.Link("/", g.Text("Home"))),
        breadcrumb.Separator(),
        breadcrumb.Item(breadcrumb.Link("/users", g.Text("Users"))),
        breadcrumb.Separator(),
        breadcrumb.Item(breadcrumb.Page(g.Text("Alice"))),
    ),
)
```

---

### interactive/dialog

Native HTML `<dialog>` with `showModal`. Closes on click outside the content area.

```go
func Root(children ...g.Node) g.Node
func Trigger(id string, children ...g.Node) g.Node
func Content(id string, children ...g.Node) g.Node
func Header(children ...g.Node) g.Node
func Title(id string, children ...g.Node) g.Node
func Description(id string, children ...g.Node) g.Node
func Footer(children ...g.Node) g.Node
func Close(children ...g.Node) g.Node
```

```go
dialog.Root(
    dialog.Trigger("confirm-delete", core.Button(core.ButtonProps{
        Label:   "Delete",
        Variant: core.VariantDestructive,
    })),
    dialog.Content("confirm-delete",
        dialog.Header(
            dialog.Title("confirm-delete", g.Text("Confirm deletion")),
            dialog.Description("confirm-delete", g.Text("This action cannot be undone.")),
        ),
        dialog.Footer(
            dialog.Close(core.Button(core.ButtonProps{Label: "Cancel", Variant: core.VariantOutline})),
            core.Button(core.ButtonProps{Label: "Delete", Variant: core.VariantDestructive, Type: "submit"}),
        ),
    ),
)
```

---

### interactive/dropdown

Dropdown menu using `<details>/<summary>`.

```go
type Props struct { ID, Align string; Extra []g.Node }
type TriggerProps struct { Extra []g.Node }
type ItemProps struct { Href, Label string; Icon *core.IconProps; Extra []g.Node }

func Root(p Props, children ...g.Node) g.Node
func Trigger(p TriggerProps, children ...g.Node) g.Node
func Content(children ...g.Node) g.Node
func Label(text string) g.Node
func Item(p ItemProps) g.Node
func Separator() g.Node
```

`Align` accepts `"left"` or `"right"`.

---

### interactive/pagination

```go
func Root(children ...g.Node) g.Node
func Content(children ...g.Node) g.Node
func Item(children ...g.Node) g.Node
func Link(href string, active bool, children ...g.Node) g.Node
func Previous(href string, disabled bool) g.Node
func Next(href string, disabled bool) g.Node
```

---

### interactive/tabs

CSS-only tab switching using radio button groups. No JavaScript required.

```go
func Root(id, value string, children ...g.Node) g.Node
func List(children ...g.Node) g.Node
func Trigger(id, value string, active bool, children ...g.Node) g.Node
func Content(id, value string, active bool, children ...g.Node) g.Node
```

```go
tabs.Root("user-tabs", "profile",
    tabs.List(
        tabs.Trigger("user-tabs", "profile", true, g.Text("Profile")),
        tabs.Trigger("user-tabs", "settings", false, g.Text("Settings")),
    ),
    tabs.Content("user-tabs", "profile", true, g.Text("Profile content")),
    tabs.Content("user-tabs", "settings", false, g.Text("Settings content")),
)
```

---

### interactive/tooltip

Hover and click tooltip variants.

```go
// Hover variant (CSS :hover)
func Root(children ...g.Node) g.Node
func Trigger(children ...g.Node) g.Node
func Content(id string, children ...g.Node) g.Node
func TriggerAttrs(id string) []g.Node

// Click variant (popover API)
func ClickRoot(children ...g.Node) g.Node
func ClickTrigger(extra ...g.Node) g.Node
func ClickContent(id string, children ...g.Node) g.Node
```

---

### patterns/flash

Flash messages via short-lived HTTP-only cookies. Cookie payload is base64-encoded JSON with `MaxAge=60s`, `HttpOnly`, and `SameSite=Lax`. Automatically cleared after `GetAll`.

```go
type Type string
const TypeSuccess, TypeInfo, TypeWarning, TypeError Type = ...

type Message struct{ Type Type; Text string }

func Set(w http.ResponseWriter, msgs []Message) error
func GetAll(r *http.Request, w http.ResponseWriter) ([]Message, error)

func Success(w http.ResponseWriter, text string) error
func Info(w http.ResponseWriter, text string) error
func Warning(w http.ResponseWriter, text string) error
func Error(w http.ResponseWriter, text string) error
```

#### Rendering Helpers

```go
func Render(msgs []Message) g.Node
func RenderOOB(msgs []Message) g.Node
```

`Render` converts flash messages into dismissible `Toast` components for direct container injection. `RenderOOB` converts messages into toasts with `hx-swap-oob="beforeend:#toast-container"` for HTMX out-of-band insertion.

```go
// In a handler after successful update:
flash.Success(w, "User updated successfully.")
http.Redirect(w, r, "/users", http.StatusSeeOther)

// In a layout renderer:
msgs, _ := flash.GetAll(r, w)
toasts := flash.Render(msgs)

// In an HTMX response:
msgs, _ := flash.GetAll(r, w)
oobToasts := flash.RenderOOB(msgs)
```

---

### patterns/font

Web font provider abstraction with CSP source collection.

#### Providers

| Factory | Description |
|---------|-------------|
| `Google(families ...string)` | [Google Fonts] CDN |
| `Bunny(families ...string)` | [Bunny Fonts] (GDPR-friendly) |
| `Adobe(projectID string)` | [Adobe Fonts] (Typekit) |
| `Self(faces ...Face)` | Self-hosted `@font-face` with preload |

#### Provider Interface

```go
type Provider interface {
    Nodes() []g.Node
    CSPSources() CSPSources
}

type CSPSources struct {
    StyleSources      []string
    FontSources       []string
    StyleInlineHashes []string
}
```

#### Face

| Field | Type | Description |
|-------|------|-------------|
| `Family` | `string` | CSS font-family name |
| `URL` | `string` | Fully-resolved public URL to the font file |
| `Format` | `string` | Font format hint: `"woff2"`, `"woff"`, or `"truetype"` (default `"woff2"`) |
| `Weight` | `string` | CSS font-weight value (default `"400"`) |
| `Style` | `string` | CSS font-style value: `"normal"` or `"italic"` (default `"normal"`) |

#### Configuration Types

```go
type Config struct {
    Google     GoogleConfig      `koanf:"google"`
    Bunny      BunnyConfig       `koanf:"bunny"`
    Adobe      AdobeConfig       `koanf:"adobe"`
    SelfHosted []SelfHostedGroup `koanf:"self_hosted"`
}
```

#### Helpers

```go
func BuildProviders(cfg Config, pathFunc func(string) string) []Provider
func Nodes(providers ...Provider) []g.Node
func CollectCSPSources(providers []Provider) CSPSources
```

```go
providers := font.BuildProviders(cfg.Fonts, assets.Path)
fontNodes := font.Nodes(providers...)
csp := font.CollectCSPSources(providers)
```

---

### patterns/form

Form submission binding and per-field validation error tracking.

#### Interfaces

```go
type Form interface {
    IsSubmitted() bool
    IsValid() bool
    FieldHasErrors(field string) bool
    GetFieldErrors(field string) []string
    SetFieldError(field, msg string)
    GetErrors() map[string][]string
}

type Binder interface {
    Bind(dest any) error
}

type StructValidator interface {
    Validate(i any) error
}
```

`Binder` is satisfied by `echo.Context` via its `Bind` method. `StructValidator` is satisfied by `*validate.Validator` from `pkg/server/validate`.

#### Submission

```go
func New(v StructValidator) *Submission
func (s *Submission) Submit(b Binder, dest any)
func (s *Submission) IsSubmitted() bool
func (s *Submission) IsValid() bool
func (s *Submission) FieldHasErrors(field string) bool
func (s *Submission) GetFieldErrors(field string) []string
func (s *Submission) GetErrors() map[string][]string
func (s *Submission) SetFieldError(field, msg string)
func (s *Submission) HasFormErrors() bool
func (s *Submission) GetFormErrors() []string
func (s *Submission) SetFormError(msg string)
```

`Submit` binds request data into `dest` and validates it. It never returns an error -- validation failures are stored internally and accessed via `IsValid`, `FieldHasErrors`, and `GetFieldErrors`. Struct-level failures and binding errors are stored under the `"_"` key, accessible via `HasFormErrors` and `GetFormErrors`.

```go
sub := form.New(validator)
var input model.CreateUserInput
sub.Submit(c, &input)

if !sub.IsValid() {
    // re-render form with errors
    return view.Render(c, req, fullPage, req.FormError(sub))
}
```

---

### patterns/head

HTML `<head>` element builder.

```go
type MetaProps struct {
    Title, Description string
    Keywords           []string
    FaviconHref        string
    OGImage            string
}
type Stylesheet struct{ Href string }
type Script struct{ Src string; Defer, Async bool }

type Props struct {
    Meta        MetaProps
    Stylesheets []Stylesheet
    Scripts     []Script
    Extra       []g.Node
}

func Head(p Props) g.Node
func Metatags(p MetaProps) g.Node
func CSS(stylesheets ...Stylesheet) g.Node
func JS(scripts ...Script) g.Node
```

`Head` renders the complete `<head>` element. `Metatags`, `CSS`, and `JS` render individual sections for use in custom head compositions. When `OGImage` or `Description` is set, Open Graph meta tags are emitted.

---

### patterns/htmx

HTMX request header extraction, response header helpers, typed attribute builders, and reusable interaction patterns.

#### Request

```go
type Request struct {
    Enabled, Boosted                                      bool
    Trigger, Target, TriggerName, CurrentURL              string
}

func NewRequest(r *http.Request) Request
func (r Request) IsPartial() bool
```

`IsPartial` returns `true` when the request is an HTMX request that is not boosted -- indicating a partial HTML fragment is expected.

Standalone header extraction functions:

```go
func IsRequest(r *http.Request) bool
func IsBoosted(r *http.Request) bool
func GetTrigger(r *http.Request) string
func GetTarget(r *http.Request) string
func GetTriggerName(r *http.Request) string
func GetCurrentURL(r *http.Request) string
```

#### Response

```go
type Response struct {
    Redirect, PushURL, ReplaceURL                         string
    Refresh                                               bool
    Trigger, TriggerAfterSettle, TriggerAfterSwap         string
    Retarget, Reswap                                      string
}

func (r Response) Apply(w http.ResponseWriter)
```

Standalone header setter functions:

```go
func SetRedirect(w http.ResponseWriter, url string)
func SetRefresh(w http.ResponseWriter)
func SetPushURL(w http.ResponseWriter, url string)
func SetReplaceURL(w http.ResponseWriter, url string)
func SetTrigger(w http.ResponseWriter, event string)
func SetTriggerAfterSettle(w http.ResponseWriter, event string)
func SetTriggerAfterSwap(w http.ResponseWriter, event string)
func SetRetarget(w http.ResponseWriter, selector string)
func SetReswap(w http.ResponseWriter, strategy string)
```

#### Typed Attribute Builder

```go
type AttrsProps struct {
    Get, Post, Put, Patch, Delete                         string
    Target, Swap, Select, SelectOOB                       string
    Trigger, Include, Indicator, DisabledElt               string
    Sync, Confirm, Encoding, Params                        string
    PushURL, ReplaceURL                                    string
    Values, Headers                                        map[string]string
    Boost                                                  *bool
    Extra                                                  []g.Node
}

func Attrs(p AttrsProps) []g.Node
```

Returns a slice of `g.Node` attributes for use inside [Gomponents] elements:

```go
h.Div(
    htmx.Attrs(htmx.AttrsProps{
        Get:    "/users?page=2",
        Target: "#user-list",
        Swap:   "innerHTML",
    })...,
)
```

#### Swap Constants

```go
const (
    SwapInnerHTML = "innerHTML"
    SwapOuterHTML = "outerHTML"
    SwapBeforeEnd = "beforeend"
)
```

#### Interaction Patterns

Reusable HTMX interaction patterns that compose `Attrs` with sensible defaults.

**`LiveSearch`** -- configures an input that fetches server-rendered results as the user types. Default trigger: `input changed delay:300ms, search`.

```go
type LiveSearchProps struct {
    Path, Target, Swap, Trigger, Delay, Include, Indicator, DisabledElt string
    PushURL bool
}

func LiveSearch(p LiveSearchProps) []g.Node
```

**`InlineValidation`** -- configures a field that validates server-side on change/blur. Default trigger: `change delay:200ms, blur`. Default sync: `closest form:abort`.

```go
type InlineValidationProps struct {
    Path, Target, Swap, Trigger, Include, Indicator, DisabledElt, Sync string
}

func InlineValidation(p InlineValidationProps) []g.Node
```

**`PaginatedTableLink`** -- configures a link or button that swaps a server-rendered table region.

```go
type PaginatedTableProps struct {
    Path, PageParam, Target, Swap, Include, Indicator, DisabledElt string
    Page    int
    Query   map[string]string
    PushURL bool
}

func PaginatedTableLink(p PaginatedTableProps) []g.Node
```

**`AsyncDialogTrigger`** -- configures a trigger that opens a native dialog and fetches its body asynchronously.

```go
type AsyncDialogProps struct {
    Path, DialogID, Target, Swap, Select, Indicator, DisabledElt string
}

func AsyncDialogTrigger(p AsyncDialogProps) []g.Node
```

**`DependentSelect`** -- configures a select that swaps a downstream field when its value changes.

```go
type DependentSelectProps struct {
    Path, Target, Swap, Trigger, Include, Indicator, DisabledElt string
}

func DependentSelect(p DependentSelectProps) []g.Node
```

**`OOBSwap` / `OOBAppend`** -- out-of-band swap attributes.

```go
type OOBSwapProps struct { Strategy, Selector string }

func OOBSwap(p OOBSwapProps) []g.Node
func OOBAppend(selector string) []g.Node
```

**`ToastOOB`** -- wraps a `feedback.Toast` for out-of-band insertion into a toast container.

```go
type ToastOOBProps struct {
    Toast    feedback.ToastProps
    Selector string
    Strategy string
}

func ToastOOB(p ToastOOBProps) g.Node
```

---

### patterns/pager

Pagination state extracted from query parameters.

```go
type Pager struct {
    Page, PerPage, TotalItems, TotalPages int
}

func New(r *http.Request, defaultPerPage int) Pager
func (p *Pager) SetTotal(total int)
func (p *Pager) Offset() int
func (p *Pager) IsFirst() bool
func (p *Pager) IsLast() bool
func (p *Pager) PrevPage() int
func (p *Pager) NextPage() int
```

`New` reads `?page=` and `?per_page=` from the request URL, clamping page to a minimum of 1. `SetTotal` computes `TotalPages`. `Offset` returns the SQL `OFFSET` for the current page.

```go
pgr := pager.New(r, 25)
users, total := repo.ListUsers(ctx, pgr.PerPage, pgr.Offset())
pgr.SetTotal(total)
```

---

### patterns/redirect

HTMX-aware HTTP redirect builder.

```go
func New(w http.ResponseWriter, r *http.Request) *Builder
func (b *Builder) To(url string) *Builder
func (b *Builder) StatusCode(code int) *Builder
func (b *Builder) Go() error
```

For HTMX non-boosted requests, sets the `HX-Redirect` header and responds with `204 No Content`. For standard or boosted requests, performs a standard HTTP redirect (default status 303).

```go
return redirect.New(w, r).To("/users").Go()
```

---

## Support Packages

### install

Initializes package-level asset and icon registries at application startup.

```go
type Config struct {
    PathFunc      func(string) string
    Catalog       iconset.Catalog
    IconOverrides map[icons.Key]icons.Ref
}

type Registries struct {
    Assets *assets.Registry
    Icons  *icons.Registry
}

func New(c Config) Registries
func ApplyDefault(c Config) Registries
```

`New` creates isolated registries (useful for tests). `ApplyDefault` populates the package-level `Default` registries in `assets` and `icons`. Call `ApplyDefault` once at startup before handling any requests. When `Catalog` is zero-valued (both `Sprites` and `Icons` are nil), `iconset.Default` is used automatically.

---

## Tier 3 -- Examples

### examples

Living component reference page that renders every component family with variants, sizes, and states.

```go
func Page() g.Node
```

Mount at a development route (e.g., `/_components`) to provide an in-app component reference:

```go
route.Add(g, echo.Route{
    Method:  http.MethodGet,
    Path:    "/_components",
    Name:    "components.show",
    Handler: func(c *echo.Context) error {
        return render.Component(c, examples.Page())
    },
})
```

---

## Design Patterns

### Props Structs

All components accept a single `Props` struct rather than positional arguments. This keeps call sites readable and allows adding new optional fields without breaking existing callers:

```go
core.Button(core.ButtonProps{
    Label:   "Save",
    Variant: core.VariantDefault,
    Type:    "submit",
})
```

### Namespace Pattern

Related sub-components are grouped under a namespace variable to avoid flat-package pollution:

```go
data.Card.Root(
    data.Card.Header(data.Card.Title(g.Text("Users"))),
    data.Card.Content(data.Table.Root(...)),
)
```

### Sealed Interface

`NavbarItem` is a private interface -- only types defined within `ui/layout` implement it. This guarantees the Navbar renderer handles every item type at compile time without a type-assertion fallback.

### Registry Pattern

Both `assets` and `icons` expose a global `Default` registry plus a constructor for isolated instances. Application code calls `assets.Path(...)` and `icons.Resolve(...)` directly, while tests construct fresh registries to avoid cross-test state:

```go
// Application startup
assets.MustInit(cfg.PublicDir, "/public")
install.ApplyDefault(install.Config{
    PathFunc: assets.Path,
    Catalog:  iconset.Default,
})

// Test isolation
reg := assets.NewRegistry()
iconReg := icons.NewRegistry()
r := install.New(install.Config{
    PathFunc: reg.SpritePath,
    Catalog:  iconset.Default,
})
```

### HTMX and Progressive Enhancement

The library is designed around server-rendered HTML with HTMX enhancement:

- `render.Component` / `render.Fragment` write full-page or partial responses
- `patterns/htmx` extracts HTMX request headers, sets response headers, and provides reusable interaction patterns (live search, inline validation, paginated tables, async dialogs, dependent selects, OOB swaps)
- `patterns/redirect.Builder` automatically chooses `HX-Redirect` vs HTTP redirect
- `patterns/flash.RenderOOB` converts flash messages to HTMX out-of-band toasts
- Interactive components use native HTML (`<details>`, `<dialog>`, radio groups) -- they work without JavaScript and are enhanced when JS is available

---

## Startup Wiring

```go
// 1. Initialize asset manifest (content-hash cache-busting)
assets.MustInit(cfg.PublicDir, "/public")

// 2. Register sprites and icon bindings
install.ApplyDefault(install.Config{
    PathFunc: assets.Path,
    Catalog:  iconset.Default,
    IconOverrides: map[icons.Key]icons.Ref{
        icons.Close: {Sprite: "my-icons", ID: "x-mark"},
    },
})

// 3. Add ThemeScript to <head> (prevents FOUC)
// In your layout component:
head.Head(head.Props{
    Extra: []g.Node{
        interactive.ThemeScript(),
        // CSP: add interactive.ScriptCSPHash to script-src
    },
})
```

---

## Thread Safety

- `assets.Registry` -- thread-safe reads after initialization (`sync.RWMutex`)
- `icons.Registry` -- thread-safe reads after initialization (`sync.RWMutex`)
- All component functions -- stateless; safe to call concurrently
- `patterns/form.Submission` -- per-request; not shared across goroutines

---

## Leaf-Node Rule

Every package in this module imports only the Go standard library and external modules within its permitted tier. There are no imports from application-specific `internal/` packages and no cross-imports between sibling `pkg/` packages outside the defined tier hierarchy. This makes the entire `github.com/go-sum/componentry` module portable to any Go web application using [Gomponents].

[Gomponents]: https://www.gomponents.com/
[Echo]: https://echo.labstack.com/
[Tailwind]: https://tailwindcss.com/
[HTMX]: https://htmx.org/
[shadcn/ui]: https://ui.shadcn.com/
[validator]: https://github.com/go-playground/validator
[Google Fonts]: https://fonts.google.com/
[Bunny Fonts]: https://fonts.bunny.net/
[Adobe Fonts]: https://fonts.adobe.com/
