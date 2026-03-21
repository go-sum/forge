# UI_GUIDE.md — UI Design Guide

> This guide defines how UI should be designed and composed in the starter.
> It is the visual companion to [`CLAUDE.md`](../CLAUDE.md) and the
> implementation companion to [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md).
>
> It incorporates the relevant design principles that were previously carried
> in `RefactoringUI-small.pdf`, adapted to this repository's actual UI surface.
> This file is intended to stand on its own after that PDF is removed.
>
> The guidance is tailored to:
>
> - reusable components in `pkg/components/ui/`
> - supporting form and HTMX helpers in `pkg/components/`
> - app-specific views in `internal/view/`

---

## Purpose

This starter targets high-performance modern web applications that are rendered
primarily on the server. The UI should therefore feel:

- clear before it feels decorative
- fast before it feels clever
- reusable before it becomes page-specific
- consistent across full-page views and HTMX partials

The component library exists to make those goals cheap to achieve. Prefer using
the shared components and tokens over rebuilding visual patterns ad hoc in
views.

---

## Scope

This guide covers:

- visual principles for components in `pkg/components/ui/`
- how page and partial views in `internal/view/` should use those components
- the default spacing, typography, color, and elevation language of the app

This guide does not try to document every exported API. For exact props and
rendering behavior, read the component source and tests.

Primary code references:

- `pkg/components/ui/core/`
- `pkg/components/ui/data/`
- `pkg/components/ui/feedback/`
- `pkg/components/ui/layout/`
- `internal/view/layout/`
- `internal/view/page/`
- `internal/view/partial/`
- `internal/view/errorpage/`

---

## Design Principles

### 1. Start with a feature, not the shell

From Refactoring UI, the most important rule is to design around the job the
user is doing, not around abstract page chrome.

In this repo that means:

- design the login form before rethinking the navbar
- design the users table before inventing a dashboard layout
- design the inline edit flow before adding decorative wrappers

`internal/view/page/auth.go` and `internal/view/page/users.go` are the model:
the core task is visible immediately, and surrounding layout is minimal.

### 2. Detail comes later

Do not start by tuning shadows, icon sizes, or decorative accents.

Start with:

- the job the user is trying to do
- the information they need to see
- the action they need to take
- the states the screen has to support

Then refine:

- hierarchy
- spacing
- typography
- color
- depth

This is especially important in server-rendered UI work, where it is easy to
burn time polishing the shell before the actual feature exists.

### 3. Ship the smallest useful version first

If part of a feature is optional, nice-to-have, or likely to be expensive, do
not let it block the first useful version.

For new UI, build:

- the happy path first
- the minimum credible empty state
- the minimum credible error state
- the minimum credible loading state if the surface needs one

Then iterate.

### 4. Choose a personality deliberately

Every interface communicates a personality whether intended or not. In this
starter, the default personality should be:

- clear
- competent
- modern
- understated

That personality is expressed through:

- restrained color usage
- consistent corner radius
- a small type scale
- straightforward copy
- quiet but polished interaction states

Do not mix competing personalities on the same screen. If a future project
using this starter wants a different tone, that should be changed centrally in
theme tokens, typography, radius, and copy style rather than page by page.

### 5. Limit choices on purpose

Design quality improves when the system narrows the decision surface.

This starter should rely on predefined systems for:

- typography
- spacing
- semantic color
- elevation
- component variants
- widths and layout constraints

When making a UI decision, choose from the system first. If the system feels
too small, expand it deliberately; do not bypass it with arbitrary one-off
values.

### 6. Use hierarchy before decoration

Most emphasis should come from:

- spacing
- font weight
- limited text-size changes
- muted vs. foreground text

Do not reach for extra colors, borders, or badges first.

Existing hierarchy patterns to follow:

- page title: `text-2xl font-bold`
- card title: `text-lg font-semibold`
- body and controls: `text-sm`
- secondary/supporting text: `text-muted-foreground`
- badges and helper text: `text-xs`

Examples:

- [`pkg/components/ui/data/card.go`](../pkg/components/ui/data/card.go)
- [`pkg/components/ui/core/button.go`](../pkg/components/ui/core/button.go)
- [`internal/view/page/home.go`](../internal/view/page/home.go)

### 7. Design in grayscale first, then apply semantic color

Refactoring UI's advice to delay color maps well to this starter.

Use semantic tokens only when they carry meaning:

- `primary` for the main action
- `destructive` for dangerous actions and error emphasis
- `secondary` and `muted` for lower emphasis
- `accent` for hover/focus surfaces

Do not introduce arbitrary color families in views when a semantic token
already exists.

Use the component defaults where possible:

- primary button: `core.VariantDefault`
- destructive button: `core.VariantDestructive`
- destructive alert: `feedback.AlertDestructive`
- status badge: `core.BadgeDefault`, `core.BadgeSecondary`, `core.BadgeDestructive`

### 8. Keep the scale tight

The visual system in this repo already leans on a small number of recurring
sizes. Keep using them.

Recommended spacing rhythm:

- `gap-2` / `p-2` for dense controls and table cells
- `gap-3` for compact form flows
- `gap-4` for related blocks
- `p-4` for compact panels and alerts
- `p-6` for cards and major grouped content
- `py-6`, `py-12`, `py-16`, `py-24` for page-level breathing room

Recommended text rhythm:

- `text-xs` for badges and tiny metadata
- `text-sm` for most UI copy and controls
- `text-lg` for card titles
- `text-2xl` for page headings

If a new design needs a larger type ramp or more spacing values than these,
the design probably needs simplification before it needs new utilities.

### 9. Let empty space do work

Refactoring UI emphasizes starting with more white space than you think you
need. That is correct for this codebase.

Preferred patterns:

- constrain forms with `max-w-sm` or `max-w-md`
- constrain error pages with `max-w-2xl`
- keep the main shell readable with `container mx-auto px-4 py-6`
- use `space-y-*` and `gap-*` instead of stacking many unrelated margins

Avoid stretching narrow content across the full screen unless the content is
truly tabular or dashboard-like.

### 10. Use depth sparingly and intentionally

Depth is already encoded in the shared components:

- cards and buttons use `shadow-xs`
- toasts use `shadow-md`
- mobile sidebar uses `shadow-xl`

Use borders for separation and shadows for elevation. Do not stack both
aggressively everywhere.

Good defaults:

- cards: elevated surface
- tables: border-driven structure
- overlays and drawers: stronger shadow
- forms inside tables: minimal container treatment, rely on structure and spacing

### 11. Accessibility is part of the design language

Accessible UI is the default, not an optional pass afterward.

Current patterns to preserve:

- focus-visible rings on buttons and inputs
- destructive color on invalid labels and fields
- `aria-describedby` / `aria-errormessage` wiring via form helpers
- correct announcement roles for alerts and toasts
- semantic HTML tables, headings, forms, and navigation

Examples:

- [`pkg/components/ui/core/button.go`](../pkg/components/ui/core/button.go)
- [`pkg/components/ui/core/label.go`](../pkg/components/ui/core/label.go)
- [`pkg/components/ui/feedback/alert.go`](../pkg/components/ui/feedback/alert.go)
- [`pkg/components/ui/feedback/toast.go`](../pkg/components/ui/feedback/toast.go)

---

## Visual Language

### Typography

Use a narrow, purposeful type scale:

- page headings: `text-2xl font-bold`
- section or card headings: `text-lg font-semibold`
- controls, paragraphs, table content: `text-sm`
- metadata and badges: `text-xs`

Supporting rules:

- prefer weight and contrast over jumping multiple font sizes
- use `leading-none` for compact headings when the component already provides it
- use muted text for descriptions, hints, empty states, and secondary metadata

Additional typography rules carried into this guide:

- separate visual hierarchy from document hierarchy
- do not let semantic element choice force oversized styling
- keep line lengths readable for prose and descriptive content
- adjust large text disproportionately faster than small text on smaller screens

#### Separate visual hierarchy from document hierarchy

Choose heading tags for semantics and accessibility, then style them based on
their visual role.

Examples:

- a page title can be an `h1` and still be visually restrained
- a card title can be semantically important without looking like a hero
- some section titles are effectively labels and should be styled that way

#### Keep line length readable

For longer descriptive copy, constrain width rather than letting text fill the
layout indefinitely.

Use:

- narrow containers for auth and settings forms
- constrained prose widths for explanations and help text
- wider containers only for tables, dashboards, and data-heavy surfaces

#### Relative sizing does not scale automatically

Do not assume that if body text shrinks by some ratio, headings, padding, and
adjacent elements should shrink by the same ratio.

In practice:

- large headings often need to shrink faster than body text on smaller screens
- button padding and font size should be tuned independently by size variant
- card and form spacing can stay comfortable even when type scales down slightly

#### Use letter spacing intentionally

Letter spacing is not neutral. It affects both legibility and personality:

- headings benefit from slightly tighter tracking (`tracking-tight`, ~-0.025em)
  to look intentional and confident rather than spaced-apart
- all-caps labels and overlines need looser tracking (`tracking-wide` or
  `tracking-wider`, ~+0.05–0.1em) because uppercase letters crowd each other
  at normal spacing
- body text and controls should remain at default tracking

Do not add letter spacing to mixed-case body copy — it degrades readability.

#### Line height scales inversely with font size

Larger text needs less line height than small text:

- headings (`text-lg` and above): `leading-tight` or `leading-snug` (1.1–1.3)
- body text (`text-sm` / `text-base`): `leading-normal` or `leading-relaxed`
  (1.5–1.6)
- dense metadata and badges: can use `leading-none` or `leading-tight`

A large heading with `leading-relaxed` looks unintentionally airy. Body text
with `leading-tight` becomes hard to read.

#### Right-align numbers in tables

Numeric columns in tables should be right-aligned so that values of different
magnitudes stay comparable at a glance. Label columns and free-text columns
should remain left-aligned.

Use `text-right` on both the `<th>` and each `<td>` in numeric columns. Keep
mixed columns (e.g. name + secondary info stacked) left-aligned.

### Color

Use semantic color tokens already present in the design system:

- `bg-background`, `text-foreground`
- `bg-card`, `text-card-foreground`
- `text-muted-foreground`
- `bg-primary`, `text-primary-foreground`
- `bg-secondary`, `text-secondary-foreground`
- `bg-destructive`, `text-destructive`
- `hover:bg-accent`, `hover:text-accent-foreground`
- `border-border`, `border-input`, `border-ring`

Rules:

- never use color as the only signal for important state
- on colored surfaces, use the matching foreground token or opacity variant
- prefer semantic variants over raw palette classes in shared UI

Additional color rules carried into this guide:

- do not use generic grey text on colored backgrounds
- define shades and state meaning up front
- avoid washed-out tints when a saturated low-opacity version communicates better
- neutral colors do not need to be mathematically grey if a warmer or cooler
  neutral fits the palette better

#### Build palettes with enough shades

Most color decisions require more steps in the scale than designers initially
expect. A useful scale has:

- 8–10 steps for greys (near-white → near-black)
- 5–10 steps for each accent or brand color (very light tint → very dark shade)

#### Meet WCAG contrast ratios

Text must be readable. The WCAG minimum contrast requirements are:

- **4.5:1** for normal text (roughly `text-sm` and smaller, or under ~18px)
- **3:1** for large text (roughly `text-lg` and above, or over ~18px bold)

#### Do not use grey text on colored backgrounds

On colored surfaces, use:

- the matching foreground token
- reduced opacity of that foreground when needed
- a hand-picked semantic token if the component requires one

Do not drop in `text-gray-*` style thinking on top of colored or image-based
surfaces.

#### Do not rely on color alone

Use shape, copy, icons, placement, and text labels alongside color when
communicating:

- destructive state
- success
- warning
- selection
- disabled state

In this codebase, that often means pairing:

- destructive color with alert copy or an explicit destructive variant
- badge color with readable badge text
- invalid input styling with an error message
- current nav state with both styling and structural position

### Spacing

Use consistent spacing instead of one-off values. Page composition should read
as a rhythm, not as a pile of local tweaks.

Good existing examples:

- auth forms: compact vertical flow inside a padded card
- user list region: `space-y-4` between table and pagination
- error page: `flex max-w-2xl flex-col gap-6 py-16`

Additional layout and spacing rules carried into this guide:

- start with more white space than you think you need, then remove it
- keep more space around a group than inside it
- dense UIs should be a deliberate exception
- grids are tools, not laws
- use fixed widths or max widths when the content demands them

#### Start with more white space, then remove it

When a UI feels cramped, the problem is usually not subtle styling, it is
insufficient breathing room.

Default bias:

- give forms, cards, and major sections more room first
- tighten only when the screen density has a clear product reason

#### Keep more space around a group than inside it

This is one of the most important spacing rules in the system.

Examples:

- label-to-input spacing should be smaller than field-to-field spacing
- row action gaps should be smaller than the distance between the actions and the next table row
- card title/description spacing should be tighter than card-to-card spacing

If intra-group and inter-group spacing are too similar, the UI becomes hard to
scan.

#### Grids are tools, not laws

Do not force every layout into equal fluid columns.

Prefer:

- fixed or max-width sidebars when navigation needs a stable width
- constrained cards and forms that only shrink when necessary
- content-driven widths inside a flexible container

Use fluid distribution where it helps, not because a grid exists.

### Elevation

Use these defaults:

- none or border-only for structural rows and sections
- `shadow-xs` for cards and small controls
- `shadow-md` for transient feedback
- `shadow-xl` for off-canvas drawers and overlays

Additional depth rules carried into this guide:

- treat the light source as consistent
- use shadows to show elevation, not decoration for its own sake
- overlap and layering should be rare and meaningful
- borders and shadows should complement one another, not compete

#### Keep the implied light source consistent

This UI system assumes a conventional top-down light source. Avoid mixing
shadow directions or adding effects that imply competing lighting logic.

#### Shadow scale

Define a small, fixed shadow scale rather than picking values ad hoc. Five
levels is enough for most interfaces:

| Level | Typical use | Example value |
|-------|-------------|---------------|
| 1 | Buttons, small controls | `0 1px 3px hsla(0,0%,0%,.2)` |
| 2 | Cards | `0 4px 6px hsla(0,0%,0%,.1)` |
| 3 | Dropdowns, popovers | `0 5px 15px hsla(0,0%,0%,.1)` |
| 4 | Toasts, non-modal overlays | `0 10px 24px hsla(0,0%,0%,.15)` |
| 5 | Modals, dialogs | `0 15px 35px hsla(0,0%,0%,.2)` |

This maps to `shadow-xs` → `shadow-md` → `shadow-xl` in the token scale.
Pick the level based on where the element lives on the z-axis — do not think
about the shadow itself; think about how close the element feels to the user.

#### Overlap only when it clarifies layers

Layering is appropriate for:

- the mobile sidebar
- toasts
- dropdowns and popovers
- dialogs

Do not create overlap in normal page flow merely for visual novelty.

---

## Component Guide

### `pkg/components/ui/core`

Use `core` for the smallest shared primitives.

Primary components:

- `Button`: shared action styling, link/button dual rendering, size and variant system
- `Badge`: compact status and role indicators
- `Label`: form labels with built-in invalid styling
- `Avatar`, `Icon`, `Separator`, `Skeleton`, `Popover`: primitive supporting UI

Rules:

- use `core.Button` for actions instead of hand-rolled `<button>` class strings
- use `core.Badge` for terse categorical status, not for paragraphs or feedback messages
- use `core.Label` through the form field helpers instead of styling labels manually
- de-emphasize heavy icons when they compete with adjacent text

### Button usage

Use variants consistently:

- `VariantDefault`: primary action
- `VariantSecondary`: lower-emphasis filled action
- `VariantOutline`: secondary action needing boundary
- `VariantGhost`: quiet inline actions
- `VariantDestructive`: dangerous action
- `VariantLink`: text-only navigation/action

Use sizes consistently:

- default for primary form and page actions
- `SizeSm` for row actions, pagination controls, nav form actions
- `SizeLg` only when the layout genuinely needs a larger target

### `pkg/components/ui/data`

Use `data` for grouped informational surfaces and tabular display.

Primary components:

- `Card.Root`, `Card.Header`, `Card.Title`, `Card.Description`, `Card.Content`, `Card.Footer`
- `Table.Root`, `Table.Header`, `Table.Body`, `Table.Row`, `Table.Head`, `Table.Cell`, `Table.Caption`

Rules:

- use cards for bounded tasks, summaries, and dialogs-in-page
- use tables for multi-column structured data, not for layout
- keep table actions compact and aligned to scan cleanly
- keep card content padded through card subcomponents, not extra wrapper div soup

### `pkg/components/ui/feedback`

Use `feedback` for messages and progress, not badges.

Primary components:

- `Alert`: inline page or form feedback
- `Toast`: transient or out-of-band feedback
- `Progress`: explicit progress display

Rules:

- alerts explain a situation in context
- toasts acknowledge an event and should stay brief
- destructive variants are for error or dangerous states, not generic emphasis
- use toast announcement roles as implemented; do not rebuild toast markup ad hoc

### `pkg/components/ui/layout`

Use `layout` for shell-level navigation and structural navigation patterns.

Primary components:

- `Navbar`
- `NavMenu`
- `Sidebar`

Rules:

- page shell navigation should be configured declaratively through `NavConfig`
- mobile drawer behavior should reuse `Sidebar`, not a second bespoke pattern
- dynamic auth or theme content should flow through nav slots, not hardcoded branches in many views

Examples:

- [`pkg/components/ui/layout/navmenu.go`](../pkg/components/ui/layout/navmenu.go)
- [`pkg/components/ui/layout/navbar.go`](../pkg/components/ui/layout/navbar.go)
- [`internal/view/layout/base.go`](../internal/view/layout/base.go)

---

## View Composition Guide

### `internal/view/layout/`

`internal/view/layout/base.go` is the application shell.

It is responsible for:

- document structure
- stylesheet and script inclusion
- nav rendering
- CSRF wiring for standard and HTMX requests
- the toast container

Do not duplicate shell concerns in page-level views.

### `internal/view/page/`

Use `internal/view/page/` for full-page constructors.

Rules:

- accept `view.Request` first
- wrap content with `req.Page(...)`
- compose page structure with shared components first, utilities second
- use utility classes for layout and spacing, not to recreate missing button/card/alert systems
- keep semantic structure correct even when the visual styling is intentionally restrained

Current page patterns:

- `auth.go`: centered auth cards with tight, task-focused forms
- `users.go`: page heading plus reusable HTMX region
- `home.go`: minimal landing page

### `internal/view/partial/`

Use `internal/view/partial/` for HTMX-replaceable fragments.

Rules:

- partials should preserve the same visual language as full pages
- partials should be structurally self-sufficient for the DOM region they replace
- prefer returning the same surface type after mutation as before mutation
- partial layouts should not become visually denser or noisier than the full page they live within

`userpartial` is the reference implementation:

- read-only row uses `Table.Row` and compact ghost/destructive actions
- edit row preserves table context, swaps in a form inline, and keeps controls compact

### `internal/view/errorpage/`

Errors should look like part of the application, not like fallback HTML.

Follow the existing pattern:

- a constrained card surface
- clear title and HTTP badge
- inline alert with the user-safe message
- one obvious escape action
- optional technical detail behind a disclosure in debug mode

---

## Practical Rules for New UI

### When a shared component exists, use it

Do not hand-roll:

- button styling
- badge styling
- card framing
- table anatomy
- alert and toast structure
- navbar or sidebar shells

Ad hoc utilities are acceptable for:

- page spacing
- responsive layout wrappers
- view-specific alignment
- one-off structural composition

### Prefer composition over variant explosion

If a screen needs a special arrangement, compose existing primitives first.
Only add a new component variant when the same visual pattern is reused in
multiple places.

### Make action hierarchy obvious

Action styling should reflect importance first, semantics second.

Default hierarchy:

- primary action: high-contrast filled button
- secondary action: outline or lower-emphasis filled treatment
- tertiary action: ghost or link treatment

Destructive actions do not automatically become primary. If the destructive
action is not the main intended action on the page, style it as a secondary or
tertiary destructive action and rely on confirmation at the point of no return.

### Keep forms readable

Most app forms in this repo should follow this structure:

- constrained width when standalone
- `uiform.Field` for each control
- `view.FormError(...)` for top-level validation state
- one obvious primary submit action
- quiet secondary navigation underneath

When labels are required, make them supportive and clear. When a value is
self-evident from context or formatting, do not add redundant label/value
noise in display UIs.

### Keep table actions quiet until needed

For tabular data:

- data should dominate, controls should support
- prefer `ghost` for edit/view actions
- reserve `destructive` for delete/danger actions
- keep row actions in a right-aligned compact group

### Balance text, icons, and borders

When one element feels too heavy:

- soften icon contrast before changing its size
- increase border weight slightly before making the color harsher
- de-emphasize competing content before over-emphasizing the target content

### Empty states, loading, and error states matter

Refactoring UI is right that polish comes from these states.

Every new surface should consider:

- what shows when there is no data
- what shows while work is happening
- what shows when the operation fails

If a page has no dedicated empty state yet, add one before adding decorative
complexity.

### Add color with accent borders

A thin colorful border is one of the simplest ways to make a neutral UI feel
more intentional. It requires no illustration skill and adds almost no visual
noise. Use it:

- across the top of a card or panel to distinguish it from its siblings
- along the left side of an alert or callout block (`border-l-4 border-primary`)
- as a short underline beneath a section heading
- as the active indicator on navigation items (bottom border on tabs, left
  border on sidebar items)
- as a single colored bar across the very top of the layout to inject brand
  color without affecting content legibility

Accent borders work because they use color structurally, not decoratively.
The color appears at the edge, not the face, so it does not fight text contrast.

### Use fewer borders

Borders are useful, but overuse creates noise.

Prefer:

- spacing for grouping
- contrast for hierarchy
- shadows for elevation
- borders for inputs, tables, and intentional separation

### Backgrounds should support, not distract

Most screens in this starter should lean on:

- `bg-background` for the page
- `bg-card` for elevated surfaces
- `bg-muted` or `bg-accent` sparingly for supporting distinction

Decorative backgrounds are acceptable only when they do not weaken
legibility or fight the app's restrained default personality.

### Use images carefully

When future screens include images:

- choose images with a clear job, not generic filler
- preserve text contrast on top of images
- give images an intended display size
- handle user-uploaded images defensively and expect inconsistent quality

#### Text over images requires deliberate contrast control

A photograph has both very light and very dark areas. No single text color
works across both. Solve this by reducing image dynamics, not by choosing a
text color:

#### Do not scale icons beyond their intended size

SVG icons drawn for 16–24px look chunky and lack visual detail when scaled to
48px or 96px — even though SVG itself is resolution-independent. The strokes
and proportions were designed at small sizes.

If a large icon is needed:

- use a larger icon from the same set drawn at that intended size
- enclose the small icon in a padded shape with a background color, keeping the
  actual icon near its intended size while filling the larger space

Similarly, large icons or screenshots shrunk down to small sizes look muddy and
unreadable. Use partial screenshots or a simplified illustration instead of
scaling the full view down.

#### Contain user-uploaded images

User-supplied images have unpredictable aspect ratios and quality. Contain them:

- render inside a fixed-size container
- use `object-fit: cover` (or `background-size: cover` for background images)
  to crop to fill, rather than letting the image dictate the layout

For circular or rounded thumbnails where the image background color may blend
into the page background, use an inner box shadow instead of a border:

```css
box-shadow: inset 0 0 0 1px hsla(0,0%,0%,.1);
```

Borders tend to clash with image colors; the inner shadow is nearly invisible
against most images while still providing separation when needed.

### De-emphasize labels relative to their values

In display UI (not form inputs), labels are secondary. The value is the
information the user came for; the label explains what the value means.

This means labels should be visually quieter than their values:

- use `text-xs text-muted-foreground` or `text-sm text-muted-foreground` for
  labels in cards, detail views, and stat blocks
- let the value carry the visual weight with normal foreground color and
  slightly larger or heavier type
- omit the label entirely when context makes the value self-evident

Avoid the reverse: a bold label and a muted value — that makes users read the
label twice and search for the data they actually want.

### Supercharge default elements

Before reaching for a new component, consider enriching what is already on the
page:

- **Bulleted lists**: replace generic bullets with relevant icons (`✓`, `→`, or
  domain-specific icons) using `pkg/components/ui/core.Icon`
- **Quotes and testimonials**: increase quotation mark size and soften their
  color — they become a visual element rather than punctuation
- **Links in body text**: a thick, colorful partial underline (using
  `text-decoration-color` or a `border-b-2` on `<a>`) is more distinctive than
  a plain underline
- **Form controls**: custom checkbox and radio styling using brand colors for
  the checked state makes forms feel polished without adding complexity
- **Radio groups for important choices**: if a radio group is a key decision on
  the screen, replace it with selectable card tiles

These are finishing details — apply them after hierarchy, spacing, and
accessibility are solid.

### Think outside component conventions

Do not assume the default shape of a component is the only option. Components
are surfaces — what you put on them is flexible:

- a dropdown does not have to be a plain list of links; it can use sections,
  multiple columns, icons, and supporting descriptions
- a table does not have to give each datum its own column; non-sortable
  secondary text (like a role below a name) can share a cell, reducing column
  count and improving readability
- a table cell is not limited to plain text — avatars, badges, and colored
  status labels add hierarchy without a separate component
- a modal-level choice does not have to use radio buttons; selectable cards
  present the same decision with more visual clarity

Apply judgment before breaking from established patterns, but do not avoid
improvements simply because the default form is familiar.

### Writing is part of the design

UI copy affects personality as much as color or type.

Default copy style for this starter:

- direct
- plain
- helpful
- not overly playful
- not legalistic unless the domain requires it

Choose words that reduce friction and match the visual restraint of the system.

---

## Decision Checklist

Before merging a UI change, confirm:

- the design starts from the feature, not from extra shell complexity
- a shared component was used where one already exists
- hierarchy comes from spacing, weight, and contrast before extra color
- semantic tokens were used instead of arbitrary palette classes
- widths are constrained where readability matters
- focus, invalid, and feedback states are visible
- action hierarchy is obvious without reading every label twice
- grouping is clear because spacing around groups is larger than spacing within them
- the mobile layout still works without inventing a second visual language
- HTMX partials match the full-page design language
- the screen has credible empty, loading, and error states where applicable
- text on colored or image backgrounds meets 4.5:1 contrast (4.5:1 normal, 3:1 large)
- labels in display UI are quieter than the values they describe
- numeric table columns are right-aligned
- headings use tighter tracking and leading than body text
- shadows reflect z-axis intent, not decoration; interaction changes that intent
- any new palette extension defines the full shade range before use

---

## Reference Map

Use these files as the practical source of truth:

- [`pkg/components/ui/core/button.go`](../pkg/components/ui/core/button.go)
- [`pkg/components/ui/core/badge.go`](../pkg/components/ui/core/badge.go)
- [`pkg/components/ui/core/label.go`](../pkg/components/ui/core/label.go)
- [`pkg/components/ui/data/card.go`](../pkg/components/ui/data/card.go)
- [`pkg/components/ui/data/table.go`](../pkg/components/ui/data/table.go)
- [`pkg/components/ui/feedback/alert.go`](../pkg/components/ui/feedback/alert.go)
- [`pkg/components/ui/feedback/toast.go`](../pkg/components/ui/feedback/toast.go)
- [`pkg/components/ui/feedback/progress.go`](../pkg/components/ui/feedback/progress.go)
- [`pkg/components/ui/layout/navbar.go`](../pkg/components/ui/layout/navbar.go)
- [`pkg/components/ui/layout/navmenu.go`](../pkg/components/ui/layout/navmenu.go)
- [`pkg/components/ui/layout/sidebar.go`](../pkg/components/ui/layout/sidebar.go)
- [`internal/view/layout/base.go`](../internal/view/layout/base.go)
- [`internal/view/page/auth.go`](../internal/view/page/auth.go)
- [`internal/view/page/users.go`](../internal/view/page/users.go)
- [`internal/view/errorpage/error.go`](../internal/view/errorpage/error.go)
- [`internal/view/partial/userpartial/user_form.go`](../internal/view/partial/userpartial/user_form.go)
- [`internal/view/partial/userpartial/user_row.go`](../internal/view/partial/userpartial/user_row.go)

When this guide and the code diverge, update the guide quickly. UI guidance is
only useful if it describes the UI that actually exists.
