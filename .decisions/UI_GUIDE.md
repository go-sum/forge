# UI Development Guide

> A comprehensive design guide for building consistent, well-designed interfaces using our component library and Tailwind CSS.

---

## Quick Reference

### 20 Key Principles Checklist

**Visual Hierarchy**
- [ ] Use font weight and color (not just size) to create hierarchy
- [ ] Stick to 2-3 shades of text (primary, secondary, tertiary)
- [ ] De-emphasize competing elements instead of over-emphasizing primary ones
- [ ] Avoid grey text on colored backgrounds—reduce opacity instead

**Layout & Spacing**
- [ ] Use consistent spacing from Tailwind scale (`gap-2`, `gap-4`, `gap-6`)
- [ ] Don't feel obligated to fill the whole screen
- [ ] Dense layouts use `gap-2`/`p-2`, loose layouts use `gap-6`/`px-6 py-6`

**Typography**
- [ ] Limit type scale to 5-6 sizes (`text-xs` → `text-sm` → `text-base`)
- [ ] Use font weights for hierarchy (`font-medium`, `font-semibold`)
- [ ] Shorter line heights for headings (`leading-none`), taller for body (`leading-relaxed`)

**Color**
- [ ] Use semantic color tokens (`primary`, `destructive`, `success`, `warning`, `muted`)
- [ ] Define variants for feedback states (alerts, badges, buttons)
- [ ] Support dark mode with `dark:` utilities

**Depth & Interactive States**
- [ ] Use shadow scale consistently (`shadow-xs` for inputs, `shadow-sm` for cards)
- [ ] Always include focus-visible rings (`focus-visible:ring-ring/50`)
- [ ] Show hover states for interactive elements (`hover:bg-accent`)
- [ ] Indicate disabled with opacity (`disabled:opacity-50`)

**Accessibility**
- [ ] Use semantic color with ARIA states (`aria-invalid:border-destructive`)
- [ ] Include descriptive labels and error messages
- [ ] Ensure keyboard navigation with visible focus states
- [ ] Link labels to inputs with proper `for`/`id` or `aria-describedby`

### Component Decision Tree

```
Need to collect user input?
├─ Text input → Field + Input
├─ Long text → Field + Textarea
├─ Choice from list → Field + Select
├─ Boolean choice → Field + Checkbox/Switch
└─ Multiple choice → Field + RadioGroup

Need to group content?
├─ Card (with header/footer) → Card + Card.Header + Card.Content
├─ Alert/notification → Alert (with variant)
└─ Simple container → div with border/shadow

Need to display data?
├─ Tabular data → Table
├─ List of items → ul/ol with proper spacing
└─ Key-value pairs → dl or Card with fields

Need emphasis or status?
├─ Call-to-action → Button (variant="default")
├─ Secondary action → Button (variant="outline" or "ghost")
├─ Status indicator → Badge
└─ System feedback → Alert or Toast

Need navigation?
├─ Main navigation → NavMenu with config
├─ Dropdown actions → DropdownMenu
└─ Page sections → Anchor links with smooth scroll
```

### Tailwind Utility Quick Reference

**Spacing Scale** (use consistently throughout)
```
gap-1    = 0.25rem (4px)   - Very tight spacing
gap-2    = 0.5rem  (8px)   - Dense layouts, badges, tight groups
gap-4    = 1rem    (16px)  - Default spacing between related items
gap-6    = 1.5rem  (24px)  - Form fields, card sections, comfortable spacing
gap-8    = 2rem    (32px)  - Major sections
gap-12   = 3rem    (48px)  - Page-level spacing
```

**Common Patterns**
```tsx
// Forms
<Form class="space-y-6">             // Vertical spacing between fields

// Cards
<Card class="gap-6">                 // Internal section spacing
  <Card.Content class="px-6" />     // Horizontal padding

// Fields
<Field class="gap-2">                // Label to input spacing

// Buttons
<Button class="gap-2 px-4 py-2">    // Icon spacing + padding

// Tables
<Table.Cell class="p-2">            // Dense data display

// Alerts
<Alert class="px-4 py-3">           // Compact message padding
```

**Typography**
```
text-xs       - 0.75rem   - Helper text, badges
text-sm       - 0.875rem  - Body text, form labels, table cells
text-base     - 1rem      - Default body text

font-medium   - 500       - Labels, field labels, nav links
font-semibold - 600       - Card titles, headings
font-bold     - 700       - Major headings (use sparingly)

leading-none  - 1         - Tight headings, single-line text
leading-relaxed - 1.625   - Comfortable body text, descriptions
```

**Semantic Colors**
```tsx
// Primary actions
bg-primary text-primary-foreground hover:bg-primary/90

// Destructive actions
bg-destructive text-white hover:bg-destructive/90

// Success feedback
text-success bg-card

// Warning feedback
text-warning bg-card

// Muted/secondary content
text-muted-foreground

// Accent for hover states
hover:bg-accent hover:text-accent-foreground
```

**Shadows**
```
shadow-xs  - Subtle depth (inputs, subtle cards)
shadow-sm  - Standard elevation (cards, dropdowns)
shadow-md  - Emphasized elevation (modals, important UI)
shadow-xl  - Maximum elevation (drawers, overlays)
```

**Focus States** (always include for interactive elements)
```tsx
focus-visible:border-ring
focus-visible:ring-ring/50
focus-visible:ring-[3px]
```

**Responsive Breakpoints**
```
sm:   640px   - Small tablets
md:   768px   - Tablets
lg:   1024px  - Laptops
xl:   1280px  - Desktops
```

---

## 1. Starting from Scratch

### Start with a Feature, Not a Layout

**❌ Anti-pattern:**
```tsx
// Designing the entire page structure first
<div class="grid grid-cols-12 gap-4">
  <aside class="col-span-3">Sidebar goes here...</aside>
  <main class="col-span-9">Content goes here...</main>
</div>
```

**✅ Better approach:**
```tsx
// Start with the actual feature—a contact form
<Form method="POST">
  <Field label="Email" name="email" required>
    <Input type="email" />
  </Field>
  <Field label="Message" name="message" required>
    <Textarea rows={4} />
  </Field>
  <Button type="submit">Send Message</Button>
</Form>

// Layout emerges naturally from content needs
```

**Why this matters:**
When you start with a feature instead of a layout framework, you:
- Focus on solving the actual user need
- Avoid overcomplicating the design
- Let the content dictate structure naturally
- Build only what's necessary

**In practice:**
1. **Identify the core feature** - What is the user trying to accomplish?
2. **Build that piece** - Create the form, the data display, the action
3. **Add context gradually** - Then add headers, navigation, surrounding UI
4. **Refactor for reuse** - Extract patterns once you see them repeated

### Design in Grayscale First

Designing in grayscale forces you to:
- **Solve hierarchy with spacing, size, and weight** before adding color
- **Prevent using color as a crutch** for poor information architecture
- **Create designs that work for colorblind users** from the start

**✅ Grayscale-first workflow:**
```tsx
// Step 1: Build structure with grayscale only
<Card>
  <Card.Header>
    <Card.Title class="text-base font-semibold">          {/* Emphasis with weight */}
      Recent Activity
    </Card.Title>
    <Card.Description class="text-sm text-muted-foreground">  {/* Hierarchy with color shade */}
      Last 7 days
    </Card.Description>
  </Card.Header>
  <Card.Content>
    {/* Content here */}
  </Card.Content>
</Card>

// Step 2: Add semantic color only where it communicates meaning
<Alert variant="destructive">                              {/* Color = error state */}
  <Alert.Title>Error</Alert.Title>
  <Alert.Description>Invalid credentials</Alert.Description>
</Alert>
```

**When to add color:**
- **State communication** - errors (red), success (green), warnings (yellow)
- **Branding** - primary actions, logos, key UI elements
- **Emphasis** - to draw attention to critical actions
- **Never for decoration alone** - every color should have meaning

### Don't Design Too Much Upfront

Build iteratively. You cannot predict every edge case or requirement before implementation.

**❌ Avoid:**
- Designing every possible state (loading, error, empty, partial data, etc.) before writing code
- Creating pixel-perfect mockups for every screen variation
- Spending weeks in design tools before touching code

**✅ Instead:**
1. **Design the happy path** - The main success scenario
2. **Implement it** - Build working code
3. **Discover edge cases** - See what breaks, what's missing
4. **Design solutions** - Address real problems, not hypothetical ones
5. **Iterate** - Repeat this cycle

**Example iteration:**
```tsx
// Iteration 1: Basic form
<Form method="POST">
  <Field label="Email" name="email">
    <Input type="email" />
  </Field>
  <Button type="submit">Subscribe</Button>
</Form>

// Iteration 2: Add error handling (discovered need during implementation)
<Form method="POST">
  <Form.Error error={formErrors?.form} />
  <Field label="Email" name="email" error={formErrors?.email} required>
    <Input type="email" />
  </Field>
  <Button type="submit">Subscribe</Button>
</Form>

// Iteration 3: Add success state (discovered need during testing)
<Form method="POST">
  <Form.Success message={successMessage} />
  <Form.Error error={formErrors?.form} />
  {/* ... rest of form ... */}
</Form>
```

---

## 2. Visual Hierarchy

Hierarchy is the most important aspect of interface design. Without clear hierarchy, users don't know where to look or what to do.

### Size Isn't Everything

**The problem:** Developers often reach for font size first when trying to create emphasis.

**❌ Creates scaling problems:**
```tsx
<h1 class="text-6xl">Page Title</h1>          {/* Too large */}
<h2 class="text-4xl">Section</h2>             {/* Still too large */}
<p class="text-2xl">Important text</p>        {/* Running out of scale */}
<p class="text-base">Normal text</p>          {/* Finally normal */}
```

**✅ Use weight and color instead:**
```tsx
<h1 class="text-2xl font-bold">Page Title</h1>                    {/* Weight for emphasis */}
<h2 class="text-lg font-semibold">Section</h2>                    {/* Slightly smaller, still strong */}
<p class="text-base font-medium">Important text</p>               {/* Medium weight = subtle emphasis */}
<p class="text-base text-muted-foreground">Secondary text</p>     {/* Color for de-emphasis */}
```

**In our component library:**
```tsx
// Button variants use color + weight, not size
<Button variant="default">Primary Action</Button>         {/* Full color emphasis */}
<Button variant="outline">Secondary</Button>              {/* Border only, less emphasis */}
<Button variant="ghost">Tertiary</Button>                 {/* Minimal emphasis */}

// All same size (text-sm), different emphasis levels
```

### The 2-3 Shade Rule

Limit text colors to create clear hierarchy:
1. **Primary text** - `text-foreground` - Main content, headings
2. **Secondary text** - `text-muted-foreground` - Supporting content, labels
3. **Tertiary text** - `text-muted-foreground` with reduced opacity - Timestamps, meta info

**✅ Example from Card component:**
```tsx
<Card>
  <Card.Header>
    <Card.Title>New user registered</Card.Title>              {/* Primary: font-semibold */}
    <Card.Description>2 minutes ago</Card.Description>        {/* Secondary: text-muted-foreground */}
  </Card.Header>
  <Card.Content>
    <p>John Doe (john@example.com) created an account.</p>    {/* Primary text */}
  </Card.Content>
</Card>
```

**❌ Avoid:**
```tsx
// Too many shades creates visual noise
<p class="text-gray-900">Primary</p>
<p class="text-gray-800">...</p>
<p class="text-gray-700">...</p>
<p class="text-gray-600">...</p>
<p class="text-gray-500">...</p>
<p class="text-gray-400">Way too many options</p>
```

### Emphasize by De-emphasizing

Instead of making important elements bigger/bolder, make competing elements smaller/lighter.

**❌ Over-emphasizing:**
```tsx
<div>
  <p class="text-sm text-muted-foreground">Label</p>
  <p class="text-3xl font-bold text-primary">$1,234.56</p>  {/* Oversized */}
</div>
```

**✅ De-emphasize the label instead:**
```tsx
<div>
  <p class="text-xs text-muted-foreground uppercase tracking-wide">Revenue</p>  {/* Smaller, lighter */}
  <p class="text-2xl font-semibold">$1,234.56</p>                              {/* Normal emphasis */}
</div>
```

**Example from Alert component:**
```tsx
// Title emphasized, description de-emphasized
<Alert variant="destructive">
  <Alert.Title class="font-medium">Error</Alert.Title>              {/* Medium weight */}
  <Alert.Description class="text-sm text-destructive/90">           {/* Smaller, slightly transparent */}
    Your session has expired. Please log in again.
  </Alert.Description>
</Alert>
```

### Don't Use Grey Text on Colored Backgrounds

**❌ This reduces contrast and looks washed out:**
```tsx
<div class="bg-primary p-4">
  <p class="text-gray-400">This is hard to read</p>  {/* Poor contrast */}
</div>
```

**✅ Reduce opacity of white text instead:**
```tsx
<div class="bg-primary p-4">
  <h3 class="text-primary-foreground font-semibold">Title</h3>      {/* Full opacity */}
  <p class="text-primary-foreground/80">Supporting text</p>         {/* 80% opacity */}
  <p class="text-primary-foreground/60">Metadata</p>                {/* 60% opacity */}
</div>
```

**In our Button component:**
```tsx
// Primary button uses white text with semantic foreground color
<Button variant="default" class="bg-primary text-primary-foreground">
  Save Changes
</Button>

// Not: class="bg-primary text-gray-300"  ← Avoid this
```

### Labels are a Last Resort

Before adding a label, consider if the context makes it obvious.

**❌ Unnecessary labels:**
```tsx
<div>
  <span class="text-sm text-muted-foreground">Name:</span>
  <span>John Doe</span>
</div>
<div>
  <span class="text-sm text-muted-foreground">Email:</span>
  <span>john@example.com</span>
</div>
```

**✅ Clear without labels:**
```tsx
<div>
  <p class="font-medium">John Doe</p>
  <p class="text-sm text-muted-foreground">john@example.com</p>  {/* Context makes it obvious */}
</div>
```

**✅ Use labels in forms (where they're necessary for accessibility):**
```tsx
<Field label="Email address" name="email" required>  {/* Label is important here */}
  <Input type="email" />
</Field>
```

**When labels are good:**
- Forms (for accessibility and clarity)
- Settings pages (to explain options)
- Data with no inherent formatting (like "Temperature: 72°F")

**When labels are bad:**
- Email addresses (format is obvious)
- Dates (format indicates meaning)
- Names in profile contexts (position indicates meaning)

---

## 3. Layout & Spacing with Tailwind

Consistent spacing is crucial for professional-looking interfaces. Our component library uses Tailwind's spacing scale systematically.

### Establish a Spacing Scale

**Our standard scale:**
```
gap-1  (4px)   - Badges, very tight groups
gap-2  (8px)   - Fields (label to input), tight layouts, table cells
gap-4  (16px)  - Related items, nav menu items
gap-6  (24px)  - Form fields, card sections (our most common)
gap-8  (32px)  - Major sections
gap-12 (48px)  - Page-level spacing
```

**✅ Component examples:**
```tsx
// Form: space-y-6 (24px between fields)
<Form class="space-y-6">
  <Field />
  <Field />
  <Button />
</Form>

// Card: gap-6 (24px between sections), px-6 (24px horizontal padding)
<Card class="gap-6">
  <Card.Header class="px-6" />  {/* Applies to all Card children */}
  <Card.Content class="px-6" />
  <Card.Footer class="px-6" />
</Card>

// Field: gap-2 (8px label to input)
<Field class="gap-2">
  <label />
  <input />
  <error />
</Field>

// Table: p-2 (8px dense data)
<Table.Cell class="p-2">...</Table.Cell>

// Button: gap-2 (8px icon to text)
<Button class="gap-2">
  <Icon />
  Save
</Button>
```

**❌ Avoid:**
```tsx
// Arbitrary spacing breaks consistency
<div class="mt-3 mb-5 px-7">  {/* No, use scale values */}
```

**✅ Stick to the scale:**
```tsx
<div class="my-6 px-6">  {/* Yes, uses scale values */}
```

### You Don't Have to Fill the Whole Screen

Empty space is not wasted space. Generous spacing makes interfaces feel less cluttered.

**❌ Cramming everything together:**
```tsx
<div class="container mx-auto px-4">
  <Form class="space-y-2">              {/* Too tight */}
    <Field class="gap-1">               {/* No breathing room */}
      <Input class="py-0.5 px-2" />     {/* Cramped */}
    </Field>
  </Form>
</div>
```

**✅ Let it breathe:**
```tsx
<div class="container max-w-md mx-auto px-4 py-12">  {/* Contained width, generous padding */}
  <Form class="space-y-6">                           {/* Comfortable field spacing */}
    <Field class="gap-2">                            {/* Proper label spacing */}
      <Input class="py-1 px-3" />                    {/* Standard input padding */}
    </Field>
  </Form>
</div>
```

**Use max-width to avoid overstretching:**
```tsx
// Forms and text content should be constrained
<div class="max-w-md">  {/* 448px max */}
  <Form>...</Form>
</div>

<div class="max-w-prose">  {/* ~65ch for readable text */}
  <p>Long form content...</p>
</div>

<div class="max-w-7xl">  {/* Wide dashboard layouts */}
  <Table>...</Table>
</div>
```

### Dense vs. Loose Layouts

Different content types need different spacing densities.

**Dense (gap-2, p-2):**
- Tables - data needs to be scannable
- Badges - compact labels
- Inline forms - search bars, filters

```tsx
<Table>
  <Table.Cell class="p-2">Dense data</Table.Cell>  {/* 8px padding */}
</Table>

<Badge class="px-2 py-0.5">Status</Badge>  {/* Compact */}
```

**Standard (gap-4, px-4 py-2):**
- Buttons - comfortable click targets
- Dropdown items - easy to tap
- Navigation links

```tsx
<Button class="px-4 py-2 gap-2">Action</Button>

<NavMenu.Item>Link</NavMenu.Item>  {/* Comfortable spacing */}
```

**Loose (gap-6, px-6 py-6):**
- Forms - need clear field separation
- Cards - section breathing room
- Page sections - major content blocks

```tsx
<Form class="space-y-6">...</Form>

<Card class="gap-6 py-6">
  <Card.Header class="px-6" />
  <Card.Content class="px-6" />
</Card>
```

### Responsive Spacing

Increase spacing on larger screens for better use of available space.

```tsx
// Forms: tighter on mobile, more spacious on desktop
<Form class="space-y-4 md:space-y-6">
  <Field />
  <Field />
</Form>

// Cards: adjust padding
<Card class="px-4 py-4 sm:px-6 sm:py-6">
  <Card.Content />
</Card>

// Sections: scale up spacing
<section class="space-y-8 lg:space-y-12">
  <div />
  <div />
</section>
```

---

## 4. Typography System

Good typography creates clear hierarchy and readability without requiring large size variations.

### Establish a Type Scale

**Our scale (limited to prevent chaos):**
```
text-xs    - 0.75rem  (12px)  - Badge labels, helper text, table meta
text-sm    - 0.875rem (14px)  - Body text, form labels, buttons, table cells
text-base  - 1rem     (16px)  - Default body text, card titles (with weight)
text-lg    - 1.125rem (18px)  - Section headings
text-xl    - 1.25rem  (20px)  - Page titles
text-2xl   - 1.5rem   (24px)  - Hero headings (use sparingly)
```

**❌ Avoid going beyond this scale:**
```tsx
<h1 class="text-6xl">This is way too large</h1>  {/* Rarely needed */}
```

**✅ Use weight for additional hierarchy:**
```tsx
<h1 class="text-2xl font-bold">Major Heading</h1>              {/* Largest size */}
<h2 class="text-xl font-semibold">Section Heading</h2>         {/* Slightly smaller */}
<h3 class="text-base font-semibold">Card Title</h3>            {/* Base size + weight */}
<p class="text-sm font-medium">Label</p>                       {/* Small + medium weight */}
<p class="text-sm text-muted-foreground">Helper text</p>       {/* Small + muted */}
```

### Font Weights for Hierarchy

**Our weights:**
```
font-medium   (500)  - Field labels, table headers, nav links
font-semibold (600)  - Card titles, section headings, alert titles
font-bold     (700)  - Major page headings (use sparingly)
```

**✅ Component examples:**
```tsx
// Field component: font-medium for labels
<Field.Label class="font-medium text-sm">Email</Field.Label>

// Card component: font-semibold for titles
<Card.Title class="font-semibold">Recent Activity</Card.Title>

// Alert component: font-medium for title
<Alert.Title class="font-medium">Error</Alert.Title>

// Button component: font-medium for all buttons
<Button class="font-medium">Save Changes</Button>
```

**Hierarchy in practice:**
```tsx
<Card>
  <Card.Header>
    <Card.Title class="font-semibold">Dashboard</Card.Title>        {/* 600 weight */}
    <Card.Description class="text-muted-foreground">              {/* Regular weight */}
      Overview of your metrics
    </Card.Description>
  </Card.Header>
  <Card.Content>
    <p class="text-sm font-medium">Active Users</p>                {/* 500 weight */}
    <p class="text-2xl font-semibold">1,234</p>                    {/* 600 weight, larger */}
    <p class="text-sm text-muted-foreground">+12% from last week</p>  {/* Regular, muted */}
  </Card.Content>
</Card>
```

### Line Height is Proportional

**Rule:** Shorter line heights for headings, taller for body text.

```
leading-none     (1)      - Single-line headings, tight labels
leading-relaxed  (1.625)  - Comfortable body text, descriptions
```

**✅ In components:**
```tsx
// Field labels: leading-none (single line, tight)
<Field.Label class="text-sm leading-none">Email Address</Field.Label>

// Card descriptions: leading-relaxed (readable, comfortable)
<Card.Description class="text-sm leading-relaxed">
  This is a longer description that needs comfortable line spacing
  for readability across multiple lines.
</Card.Description>

// Alert descriptions: leading-relaxed (readable error messages)
<Alert.Description class="text-sm [&_p]:leading-relaxed">
  <p>Your session has expired. Please log in again to continue.</p>
</Alert.Description>
```

**❌ Avoid:**
```tsx
// Heading with tall line height (looks disconnected)
<h2 class="text-2xl leading-relaxed">Section Title</h2>

// Body text with tight line height (hard to read)
<p class="text-base leading-none">
  This paragraph has text that is too tightly spaced for
  comfortable reading across multiple lines.
</p>
```

### Responsive Text Sizing

Our Input component demonstrates responsive text sizing:

```tsx
// workspaces/componentry/src/ui/server/input.tsx:35
<input class="text-base md:text-sm" />
```

**Why:** Mobile browsers zoom in on inputs with font-size < 16px (prevents auto-zoom on focus)

**Pattern:**
- Mobile: `text-base` (16px) prevents zoom
- Desktop: `md:text-sm` (14px) matches rest of UI

**Apply to forms:**
```tsx
<Input class="text-base md:text-sm" type="email" />
<Textarea class="text-base md:text-sm" />
<Select class="text-base md:text-sm" />
```

---

## 5. Color System with Tailwind

Our component library uses semantic color tokens for maintainability and theme support.

### Semantic Color Tokens

Instead of using `bg-blue-500` or `text-red-600`, we use meaningful tokens:

**Primary colors:**
```tsx
bg-primary                  // Primary brand color
text-primary-foreground     // Text on primary background
text-primary                // Primary color text
hover:bg-primary/90         // Primary hover (90% opacity)
```

**Feedback colors:**
```tsx
// Destructive (errors, dangerous actions)
bg-destructive
text-destructive
border-destructive
ring-destructive/20

// Success (confirmations, completed states)
text-success
bg-success (rare, usually just text)

// Warning (caution, important notices)
text-warning
bg-warning (rare, usually just text)
```

**Neutral colors:**
```tsx
// Backgrounds
bg-background     // Page background
bg-card           // Card/panel background
bg-input          // Input background

// Foreground
text-foreground         // Default text
text-muted-foreground   // De-emphasized text
text-card-foreground    // Text on card background

// Interactive states
bg-accent                    // Hover background for interactive elements
text-accent-foreground       // Text on accent background
hover:bg-accent             // Common hover pattern
```

**Borders:**
```tsx
border-input      // Input borders
border-border     // General borders
border-ring       // Focus ring color
```

### Component Examples

**Button variants demonstrate semantic color usage:**
```tsx
// workspaces/componentry/src/ui/server/button.tsx:14-21
<Button variant="default">
  // bg-primary text-primary-foreground hover:bg-primary/90
</Button>

<Button variant="destructive">
  // bg-destructive text-white hover:bg-destructive/90
</Button>

<Button variant="outline">
  // border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground
</Button>

<Button variant="ghost">
  // hover:bg-accent hover:text-accent-foreground
</Button>
```

**Alert variants for feedback:**
```tsx
// workspaces/componentry/src/ui/server/alert.tsx:16-20
<Alert variant="default">
  // bg-card text-card-foreground
</Alert>

<Alert variant="destructive">
  // text-destructive bg-card
</Alert>

<Alert variant="success">
  // text-success bg-card
</Alert>

<Alert variant="warning">
  // text-warning bg-card
</Alert>
```

### Dark Mode Support

All components support dark mode using Tailwind's `dark:` prefix.

**Patterns in components:**
```tsx
// Input component (workspaces/componentry/src/ui/server/input.tsx:33)
<Input class="
  bg-transparent
  dark:bg-input/30                              // Subtle background in dark mode

  border-input
  dark:border-input                             // Consistent border

  aria-invalid:ring-destructive/20
  dark:aria-invalid:ring-destructive/40         // Stronger ring in dark mode
" />

// Button outline variant (workspaces/componentry/src/ui/server/button.tsx:18)
<Button variant="outline" class="
  bg-background
  dark:bg-input/30                              // Slightly tinted in dark mode

  border
  dark:border-input                             // Semantic border color

  hover:bg-accent
  dark:hover:bg-input/50                        // Distinct hover in dark mode
" />
```

**Best practices:**
1. **Use semantic tokens** - They automatically adapt to dark mode
2. **Test both modes** - Design components to work in light and dark
3. **Adjust opacity in dark mode** - Often need stronger opacity (e.g., `/20` → `/40`)
4. **Background tints** - Use `dark:bg-input/30` for subtle depth

### Don't Let Lightness Kill Your Saturation

When creating lighter versions of colors, don't just increase lightness—it makes colors washed out.

**❌ Avoid:**
```tsx
// These approaches create washed-out, grey-ish colors
<div class="bg-blue-100">Looks grey, not blue</div>
<div class="bg-red-200 opacity-50">Even worse</div>
```

**✅ Use opacity on saturated colors:**
```tsx
// Maintains saturation while lightening
<div class="bg-primary/10">Light but still colorful</div>
<div class="bg-destructive/20">Light red that looks red</div>
<div class="bg-success/15">Light green that looks green</div>
```

**In our components:**
```tsx
// Focus rings use saturated color with opacity
focus-visible:ring-ring/50          // 50% opacity maintains color

// Hover states on colored backgrounds
hover:bg-primary/90                 // Slightly transparent, keeps saturation

// Destructive ring states
aria-invalid:ring-destructive/20    // Light but visibly red
dark:aria-invalid:ring-destructive/40  // Stronger in dark mode
```

---

## 6. Depth & Elevation

Even flat designs need depth to create visual hierarchy and indicate interactivity.

### Shadow Scale

**Our scale:**
```
shadow-xs  - Subtle depth (inputs, minimal elevation)
shadow-sm  - Standard elevation (cards, dropdown menus)
shadow-md  - Emphasized elevation (modals, important panels)
shadow-xl  - Maximum elevation (drawers, full-screen overlays)
```

**✅ Component usage:**
```tsx
// Input: shadow-xs (subtle depth, not floating)
<Input class="shadow-xs border" />

// Card: shadow-sm (clearly elevated from background)
<Card class="shadow-sm border" />

// NavMenu mobile button: shadow-md (important, always visible)
<summary class="shadow-md bg-background">Menu</summary>

// NavMenu drawer: shadow-xl (overlays entire page)
<div class="shadow-xl bg-card">Drawer content</div>
```

**When to use shadows:**
- **Inputs** - `shadow-xs` creates subtle depth perception
- **Cards/Panels** - `shadow-sm` indicates they're surfaces on the background
- **Dropdowns/Popovers** - `shadow-md` shows they float above content
- **Modals/Drawers** - `shadow-xl` maximum elevation for overlays

**When to use borders instead:**
- **Tables** - borders are better for data grids
- **Outlined buttons** - `Button variant="outline"` uses border + `shadow-xs`
- **Separated sections** - borders show division within same elevation

### Focus Rings for Depth

Focus rings add depth to interactive elements and improve accessibility.

**Our standard focus state:**
```tsx
focus-visible:border-ring              // Change border color to ring color
focus-visible:ring-ring/50             // Add ring at 50% opacity
focus-visible:ring-[3px]               // 3px ring width
```

**✅ Used in all interactive components:**
```tsx
// Button (workspaces/componentry/src/ui/server/button.tsx:10)
<Button class="
  outline-none
  focus-visible:border-ring
  focus-visible:ring-ring/50
  focus-visible:ring-[3px]
" />

// Input (workspaces/componentry/src/ui/server/input.tsx:27-28)
<Input class="
  outline-none
  focus-visible:border-ring
  focus-visible:ring-[3px]
  focus-visible:ring-ring/50
" />
```

**Invalid state rings:**
```tsx
// Show error state with ring color
aria-invalid:border-destructive           // Red border
aria-invalid:ring-destructive/20          // Light red ring (light mode)
dark:aria-invalid:ring-destructive/40     // Stronger red ring (dark mode)
```

**Why `focus-visible` instead of `focus`:**
- `focus-visible` only shows ring for keyboard navigation
- Mouse clicks don't show ring (cleaner for click interactions)
- Better UX: keyboard users get visual indicator, mouse users don't see unnecessary outlines

### Borders vs. Shadows

Use both strategically to create clear hierarchy.

**Shadows for elevation:**
```tsx
// Card floats above background
<Card class="shadow-sm border rounded-xl" />

// Button feels clickable
<Button variant="outline" class="shadow-xs border" />
```

**Borders for separation:**
```tsx
// Table rows separated by borders (not shadows)
<Table.Row class="border-b" />

// Card sections divided by borders
<Card.Header class="border-b" />
<Card.Footer class="border-t" />
```

**Combine both:**
```tsx
// Card has elevation (shadow) and definition (border)
<Card class="shadow-sm border" />

// Input has subtle shadow and clear boundary
<Input class="shadow-xs border" />
```

**❌ Avoid double-emphasis:**
```tsx
// Don't use heavy shadow AND heavy border
<div class="shadow-xl border-4 border-black">  // Too much
```

### Emulate a Light Source

Shadows should come from a consistent light source (usually top-down).

Tailwind's default shadows follow this principle:
- Shadows appear below elements (light from above)
- Larger elevations = larger, softer shadows

**✅ Stick to Tailwind defaults:**
```tsx
<div class="shadow-sm">Standard elevation</div>
<div class="shadow-md">Higher elevation</div>
```

**❌ Avoid custom shadows that break the light source:**
```tsx
// Don't create inconsistent shadow directions
<div class="shadow-[0_-4px_6px_rgba(0,0,0,0.1)]">  // Shadow above? Confusing.
```

---

## 7. Responsive Design Patterns

Our components are built mobile-first using Tailwind's responsive breakpoints.

### Breakpoints

```
sm:   640px   - Small tablets (portrait)
md:   768px   - Tablets (landscape) / small laptops
lg:   1024px  - Laptops / desktops
xl:   1280px  - Large desktops
2xl:  1536px  - Extra large screens
```

**Mobile-first approach:**
```tsx
// Base styles = mobile, add breakpoints for larger screens
<div class="space-y-4 md:space-y-6 lg:space-y-8">
  // 16px on mobile, 24px on tablets, 32px on desktop
</div>
```

### Navigation: NavMenu Component

The `NavMenu` component demonstrates responsive navigation patterns:

**Mobile (< lg):**
- Hamburger menu button (fixed top-right)
- Drawer opens from right side
- Vertical navigation stack
- Details/summary for native open/close

**Desktop (≥ lg):**
- Horizontal menu bar
- Inline navigation items
- Dropdown menus for sub-items

```tsx
// workspaces/componentry/src/ui/server/nav-menu.tsx:86-122
<div>
  {/* Mobile: Drawer with details/summary */}
  <details class="group lg:hidden">
    <summary class="fixed top-4 right-4 z-60">
      <Icon name="menu" class="group-open:hidden" />
      <Icon name="x" class="hidden group-open:block" />
    </summary>

    <div class="fixed inset-0 z-40 bg-black/50" />      {/* Backdrop */}
    <div class="fixed inset-y-0 right-0 z-50 w-80">    {/* Drawer */}
      {/* Vertical navigation */}
    </div>
  </details>

  {/* Desktop: Horizontal menu */}
  <nav class="hidden lg:flex lg:flex-row lg:items-center lg:gap-6">
    {/* Horizontal navigation */}
  </nav>
</div>
```

### Forms: Responsive Layouts

**Mobile:**
- Single column
- Full-width inputs
- Larger touch targets (text-base prevents zoom)

**Desktop:**
- Can use grid for side-by-side fields
- Smaller text (md:text-sm)
- More generous spacing

```tsx
// Basic form (stacks on all sizes)
<Form class="space-y-6">
  <Field label="Email" name="email">
    <Input type="email" class="text-base md:text-sm" />
  </Field>
  <Field label="Password" name="password">
    <Input type="password" class="text-base md:text-sm" />
  </Field>
</Form>

// Advanced: Side-by-side fields on desktop
<Form class="space-y-6">
  <div class="grid gap-6 md:grid-cols-2">
    <Field label="First name" name="firstName">
      <Input class="text-base md:text-sm" />
    </Field>
    <Field label="Last name" name="lastName">
      <Input class="text-base md:text-sm" />
    </Field>
  </div>
  <Field label="Email" name="email">
    <Input type="email" class="text-base md:text-sm" />
  </Field>
</Form>
```

### Cards: Responsive Padding

**Mobile:**
- Tighter padding (px-4)
- Smaller gaps (gap-4)

**Desktop:**
- More generous padding (sm:px-6)
- Larger gaps (sm:gap-6)

```tsx
<Card class="gap-4 px-4 py-4 sm:gap-6 sm:px-6 sm:py-6">
  <Card.Header class="px-0">  {/* px-0 because Card already has padding */}
    <Card.Title>Dashboard</Card.Title>
  </Card.Header>
  <Card.Content class="px-0">
    {/* Content */}
  </Card.Content>
</Card>
```

### Tables: Responsive Strategies

**Option 1: Horizontal scroll (best for many columns)**
```tsx
<div class="w-full overflow-x-auto">
  <Table>
    <Table.Header>
      <Table.Row>
        <Table.Head>Name</Table.Head>
        <Table.Head>Email</Table.Head>
        <Table.Head>Status</Table.Head>
        <Table.Head>Role</Table.Head>
        <Table.Head>Created</Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {/* Rows... */}
    </Table.Body>
  </Table>
</div>
```

**Option 2: Hide columns on mobile (best for 3-4 columns)**
```tsx
<Table>
  <Table.Header>
    <Table.Row>
      <Table.Head>Name</Table.Head>
      <Table.Head class="hidden sm:table-cell">Email</Table.Head>
      <Table.Head>Status</Table.Head>
      <Table.Head class="hidden md:table-cell">Created</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    <Table.Row>
      <Table.Cell>John Doe</Table.Cell>
      <Table.Cell class="hidden sm:table-cell">john@example.com</Table.Cell>
      <Table.Cell><Badge variant="success">Active</Badge></Table.Cell>
      <Table.Cell class="hidden md:table-cell">2024-01-15</Table.Cell>
    </Table.Row>
  </Table.Body>
</Table>
```

**Option 3: Stacked layout for mobile (best for detailed data)**
```tsx
// Mobile: Card-like layout
// Desktop: Table layout
<div class="space-y-4 md:space-y-0">
  {items.map((item) => (
    <div class="border rounded-lg p-4 md:hidden">
      <div class="font-medium">{item.name}</div>
      <div class="text-sm text-muted-foreground">{item.email}</div>
      <Badge variant="success">{item.status}</Badge>
    </div>
  ))}

  <Table class="hidden md:table">
    {/* Full table for desktop */}
  </Table>
</div>
```

### Container Widths

Use max-width classes to control content width:

```tsx
// Forms (narrow for readability)
<div class="max-w-md mx-auto">
  <Form>...</Form>
</div>

// Text content (optimal reading width)
<div class="max-w-prose mx-auto">
  <article>...</article>
</div>

// Dashboards (use available space)
<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
  <Dashboard>...</Dashboard>
</div>
```

---

## 8. Accessibility Guidelines

Our components include accessibility features by default. Here's how to use them correctly.

### ARIA Patterns in Components

**Field component auto-links everything:**

```tsx
// workspaces/componentry/src/ui/server/field.tsx:46-69
<Field label="Email" name="email" error={errors?.email} description="We'll never share" required>
  <Input type="email" />
</Field>

// Automatically becomes:
<div>
  <label for="email-abc123">
    Email
    <span class="text-destructive">*</span>
  </label>

  <div id="email-abc123-description">
    We'll never share
  </div>

  <input
    id="email-abc123"
    name="email"
    type="email"
    required
    aria-describedby="email-abc123-description email-abc123-error"
    aria-invalid="true"
  />

  <div id="email-abc123-error">
    Please enter a valid email
  </div>
</div>
```

**Benefits:**
- Screen readers announce label, description, and error
- Keyboard navigation works correctly
- Error state is programmatically indicated
- Required fields are marked visually and semantically

### Focus Management

All interactive components use `focus-visible` for keyboard-only focus indicators:

```tsx
// Shows ring only for keyboard navigation
<Button class="
  outline-none                          // Remove browser default
  focus-visible:border-ring             // Keyboard: change border
  focus-visible:ring-ring/50            // Keyboard: show ring
  focus-visible:ring-[3px]              // Keyboard: 3px ring
" />

// Mouse clicks don't show ring, keyboard Tab does
```

**Why this matters:**
- Keyboard users need visible focus indicators
- Mouse users don't want outlines after clicking
- `focus-visible` provides the best of both

### Color Contrast Requirements

**WCAG AA requires:**
- Normal text (< 18px): 4.5:1 contrast
- Large text (≥ 18px): 3:1 contrast
- UI components: 3:1 contrast

**Our semantic tokens meet these requirements:**
```tsx
// High contrast combinations
text-foreground on bg-background           // Body text: ✓ Passes
text-primary-foreground on bg-primary      // Button text: ✓ Passes
text-destructive on bg-background          // Error text: ✓ Passes

// Muted text still meets minimums
text-muted-foreground on bg-background     // Secondary text: ✓ Passes (typically ~4.6:1)
```

**❌ Avoid:**
```tsx
// Low contrast text
<p class="text-gray-400">Hard to read</p>  // ✗ Fails on white background

// Grey text on colored backgrounds
<div class="bg-primary">
  <p class="text-gray-300">Poor contrast</p>  // ✗ Fails
</div>
```

**✅ Use proper contrast:**
```tsx
// High contrast text
<p class="text-foreground">Easy to read</p>

// Proper foreground colors on colored backgrounds
<div class="bg-primary">
  <p class="text-primary-foreground">High contrast</p>  // ✓ Passes
</div>
```

### Keyboard Navigation Best Practices

**All interactive elements must be keyboard accessible:**

```tsx
// ✓ Native buttons are keyboard accessible
<Button type="button" onClick={handleClick}>Click me</Button>

// ✓ Links for navigation
<a href="/dashboard">Dashboard</a>

// ✗ Divs are not keyboard accessible
<div onClick={handleClick}>Don't do this</div>

// ✓ If you must use div, add keyboard support
<div
  role="button"
  tabindex="0"
  onClick={handleClick}
  onKeyDown={(e) => e.key === 'Enter' && handleClick()}
>
  Only if necessary
</div>
```

**Navigation order:**
- Tab moves forward through interactive elements
- Shift+Tab moves backward
- Enter activates buttons and links
- Space activates buttons
- Arrow keys for radio groups, select dropdowns

**Form keyboard patterns:**
```tsx
<Form>
  <Field label="Email" name="email">
    <Input type="email" />  {/* Tab to focus */}
  </Field>

  <Field label="Country" name="country">
    <Select>  {/* Tab to focus, arrow keys to select */}
      <option>USA</option>
      <option>Canada</option>
    </Select>
  </Field>

  <Field name="newsletter">
    <Checkbox />  {/* Tab to focus, Space to toggle */}
  </Field>

  <Button type="submit">  {/* Tab to focus, Enter to submit */}
    Subscribe
  </Button>
</Form>
```

### Screen Reader Considerations

**Use semantic HTML:**
```tsx
// ✓ Semantic elements announce their role
<nav>Navigation</nav>
<main>Main content</main>
<aside>Sidebar</aside>
<header>Site header</header>
<footer>Site footer</footer>

// ✗ Divs require explicit roles
<div>Unknown structure</div>
```

**Alert component uses `role="alert"` for announcements:**
```tsx
// workspaces/componentry/src/ui/server/alert.tsx:30
<Alert variant="destructive">
  // Becomes: <div role="alert">
  <Alert.Title>Error</Alert.Title>
  <Alert.Description>Your session expired</Alert.Description>
</Alert>

// Screen reader automatically announces when Alert appears
```

**Form errors announce automatically:**
```tsx
<Field label="Email" name="email" error="Invalid email format">
  <Input type="email" aria-invalid="true" />
</Field>

// Screen reader announces:
// "Email, required, edit text, invalid, Invalid email format"
```

**Provide text alternatives:**
```tsx
// ✓ Icons with labels
<Button>
  <Icon name="save" />
  Save Changes  {/* Text label */}
</Button>

// ✓ Icon-only buttons with aria-label
<Button aria-label="Close dialog">
  <Icon name="x" />
</Button>

// ✗ Icon-only without label
<Button>
  <Icon name="settings" />  {/* Screen reader doesn't know what this does */}
</Button>
```

### Required Fields

**Visual and semantic indicators:**
```tsx
// Field component marks required fields with asterisk
<Field label="Email" name="email" required>
  <Input type="email" />
</Field>

// Renders:
<label>
  Email
  <span class="ml-1 text-destructive">*</span>  {/* Visual indicator */}
</label>
<input required />  {/* Semantic indicator for browsers/screen readers */}
```

**Best practice: Indicate optional instead of required**
```tsx
// When most fields are required, mark the few optional ones
<Field label="Email" name="email" required>
  <Input type="email" />
</Field>

<Field label="Phone (optional)" name="phone">
  <Input type="tel" />
</Field>
```

---

## 9. Component-Specific Guidelines

Detailed guidance for each major component type.

### Buttons

**Variants communicate intent:**
```tsx
// Primary actions (one per section)
<Button variant="default">Save Changes</Button>

// Secondary actions (multiple allowed)
<Button variant="outline">Cancel</Button>
<Button variant="ghost">Learn More</Button>

// Destructive actions (deletion, removal)
<Button variant="destructive">Delete Account</Button>

// Links styled as text
<Button variant="link">Terms of Service</Button>
```

**Sizing:**
```tsx
<Button size="sm">Small</Button>       // height: 32px, px-3
<Button size="default">Default</Button> // height: 36px, px-4
<Button size="lg">Large</Button>       // height: 40px, px-6

// Icon buttons
<Button size="icon"><Icon name="settings" /></Button>        // 36×36px
<Button size="icon-sm"><Icon name="x" /></Button>            // 32×32px
<Button size="icon-lg"><Icon name="menu" /></Button>         // 40×40px
```

**Icon placement:**
```tsx
// Icon + text (gap-2 auto-applied)
<Button>
  <Icon name="save" />
  Save Changes
</Button>

// Icon on right
<Button>
  Continue
  <Icon name="arrow-right" />
</Button>

// Icon only (use aria-label)
<Button size="icon" aria-label="Delete item">
  <Icon name="trash" />
</Button>
```

**Loading states:**
```tsx
<Button disabled>
  <Icon name="loader-2" class="animate-spin" />
  Saving...
</Button>
```

**❌ Anti-patterns:**
```tsx
// Too many primary buttons (confusing hierarchy)
<div>
  <Button variant="default">Save</Button>
  <Button variant="default">Publish</Button>
  <Button variant="default">Share</Button>
</div>

// Use outline/ghost for secondary actions
<div>
  <Button variant="default">Publish</Button>
  <Button variant="outline">Save Draft</Button>
  <Button variant="ghost">Preview</Button>
</div>
```

### Forms & Fields

**Use Field component for all inputs:**
```tsx
// ✓ Auto-links label, error, description
<Form method="POST" class="space-y-6">
  <Field label="Email" name="email" error={errors?.email} required>
    <Input type="email" />
  </Field>

  <Field label="Message" name="message" description="Max 500 characters">
    <Textarea rows={4} />
  </Field>

  <Button type="submit">Send</Button>
</Form>
```

**Form-level errors:**
```tsx
<Form method="POST" class="space-y-6">
  <Form.Error error={errors?.form} title="Submission Failed" />

  <Field label="Email" name="email" error={errors?.email}>
    <Input type="email" />
  </Field>
</Form>
```

**Success messages:**
```tsx
<Form.Success
  title="Message sent!"
  description="We'll get back to you within 24 hours."
/>
```

**Validation patterns:**
```tsx
// Client-side validation (HTML5)
<Input type="email" required />
<Input type="url" required />
<Input type="number" min={0} max={100} />
<Input pattern="[0-9]{5}" title="5-digit ZIP code" />

// Server-side validation (display errors)
<Field label="Email" name="email" error={serverErrors?.email}>
  <Input type="email" aria-invalid={!!serverErrors?.email} />
</Field>
```

**❌ Anti-patterns:**
```tsx
// Missing labels (bad for accessibility)
<Input placeholder="Enter email" />  // Placeholder is not a label

// Should be:
<Field label="Email" name="email">
  <Input placeholder="you@example.com" />
</Field>

// Not grouping related fields
<Input name="street" />
<Input name="city" />
<Input name="zip" />

// Should use fieldset:
<fieldset class="space-y-4">
  <legend class="font-semibold">Address</legend>
  <Field label="Street" name="street"><Input /></Field>
  <Field label="City" name="city"><Input /></Field>
  <Field label="ZIP" name="zip"><Input /></Field>
</fieldset>
```

### Cards

**Basic structure:**
```tsx
<Card>
  <Card.Header>
    <Card.Title>Card Title</Card.Title>
    <Card.Description>Optional description</Card.Description>
  </Card.Header>
  <Card.Content>
    {/* Main content */}
  </Card.Content>
  <Card.Footer>
    <Button>Action</Button>
  </Card.Footer>
</Card>
```

**Header with action:**
```tsx
<Card>
  <Card.Header>
    <Card.Title>Recent Activity</Card.Title>
    <Card.Description>Last 7 days</Card.Description>
    <Card.Action>
      <Button variant="ghost" size="icon-sm">
        <Icon name="more-horizontal" />
      </Button>
    </Card.Action>
  </Card.Header>
  <Card.Content>
    {/* Activity list */}
  </Card.Content>
</Card>
```

**Divided sections:**
```tsx
<Card>
  <Card.Header class="border-b">
    <Card.Title>Settings</Card.Title>
  </Card.Header>
  <Card.Content>
    {/* Settings form */}
  </Card.Content>
  <Card.Footer class="border-t">
    <Button>Save Changes</Button>
  </Card.Footer>
</Card>
```

**❌ Anti-patterns:**
```tsx
// Nested cards (creates confusion)
<Card>
  <Card.Content>
    <Card>  {/* Don't nest cards */}
      <Card.Content>Content</Card.Content>
    </Card>
  </Card.Content>
</Card>

// Cards for everything (overuse)
<Card><Card.Content>Single word</Card.Content></Card>

// Use cards for grouping meaningful content:
<Card>
  <Card.Header>
    <Card.Title>User Profile</Card.Title>
  </Card.Header>
  <Card.Content>
    {/* Multiple fields, cohesive content */}
  </Card.Content>
</Card>
```

### Tables

**Basic structure:**
```tsx
<Table>
  <Table.Header>
    <Table.Row>
      <Table.Head>Name</Table.Head>
      <Table.Head>Email</Table.Head>
      <Table.Head>Status</Table.Head>
      <Table.Head class="text-right">Actions</Table.Head>
    </Table.Row>
  </Table.Header>
  <Table.Body>
    {users.map((user) => (
      <Table.Row key={user.id}>
        <Table.Cell class="font-medium">{user.name}</Table.Cell>
        <Table.Cell>{user.email}</Table.Cell>
        <Table.Cell>
          <Badge variant={user.active ? 'success' : 'secondary'}>
            {user.active ? 'Active' : 'Inactive'}
          </Badge>
        </Table.Cell>
        <Table.Cell class="text-right">
          <Button variant="ghost" size="sm">Edit</Button>
        </Table.Cell>
      </Table.Row>
    ))}
  </Table.Body>
</Table>
```

**Sortable headers:**
```tsx
<Table.Head>
  <button
    onClick={() => handleSort('name')}
    class="flex items-center gap-1 hover:text-foreground"
  >
    Name
    <Icon name={sortDirection === 'asc' ? 'arrow-up' : 'arrow-down'} class="size-3" />
  </button>
</Table.Head>
```

**Empty state:**
```tsx
<Table>
  <Table.Header>{/* Headers */}</Table.Header>
  <Table.Body>
    {users.length === 0 ? (
      <Table.Row>
        <Table.Cell colSpan={4} class="h-24 text-center">
          <p class="text-muted-foreground">No users found</p>
        </Table.Cell>
      </Table.Row>
    ) : (
      users.map((user) => (/* Rows */))
    )}
  </Table.Body>
</Table>
```

**❌ Anti-patterns:**
```tsx
// Inconsistent cell padding (breaks alignment)
<Table.Cell class="p-4">Name</Table.Cell>
<Table.Cell class="p-2">Email</Table.Cell>  // Different padding

// Use consistent padding:
<Table.Cell class="p-2">Name</Table.Cell>
<Table.Cell class="p-2">Email</Table.Cell>

// Actions column not aligned
<Table.Cell>
  <Button size="sm">Edit</Button>
</Table.Cell>

// Right-align actions:
<Table.Cell class="text-right">
  <Button size="sm">Edit</Button>
</Table.Cell>
```

### Navigation (NavMenu)

**Basic setup:**
```tsx
// app/config/nav.config.ts
export const NAV_CONFIG: NavMenuConfig = {
  sections: [
    {
      items: [
        { label: 'Home', href: '/' },
        { label: 'About', href: '/about' },
        {
          label: 'Products',
          items: [
            { label: 'All Products', href: '/products' },
            { type: 'separator' },
            {
              label: 'Categories',
              items: [
                { label: 'Electronics', href: '/products/electronics' },
                { label: 'Clothing', href: '/products/clothing' },
              ]
            }
          ]
        }
      ]
    }
  ]
}

// app/root.tsx
<NavMenu config={NAV_CONFIG}>
  <NavMenu.Actions>
    <Button variant="outline" size="sm">Sign In</Button>
  </NavMenu.Actions>
</NavMenu>
```

**Active link styling:**
```tsx
// Automatically styled with data-nav-link attribute
<NavMenu.Item href="/dashboard" active={currentPath === '/dashboard'}>
  Dashboard
</NavMenu.Item>

// CSS:
[data-nav-link][active] { color: var(--primary); }
```

**❌ Anti-patterns:**
```tsx
// Too many top-level items (overwhelming)
{ items: [
  { label: 'Home' },
  { label: 'About' },
  { label: 'Products' },
  { label: 'Services' },
  { label: 'Blog' },
  { label: 'Team' },
  { label: 'Careers' },
  { label: 'Contact' },
  { label: 'Support' },
] }

// Group related items:
{
  label: 'Company',
  items: [
    { label: 'About', href: '/about' },
    { label: 'Team', href: '/team' },
    { label: 'Careers', href: '/careers' },
  ]
}
```

### Alerts & Feedback

**Alert variants:**
```tsx
// Default (neutral info)
<Alert>
  <Alert.Title>Info</Alert.Title>
  <Alert.Description>Your changes have been saved.</Alert.Description>
</Alert>

// Success
<Alert variant="success">
  <Alert.Title>Success</Alert.Title>
  <Alert.Description>Account created successfully!</Alert.Description>
</Alert>

// Destructive (errors)
<Alert variant="destructive">
  <Alert.Title>Error</Alert.Title>
  <Alert.Description>Failed to save changes. Please try again.</Alert.Description>
</Alert>

// Warning
<Alert variant="warning">
  <Alert.Title>Warning</Alert.Title>
  <Alert.Description>Your subscription expires in 3 days.</Alert.Description>
</Alert>
```

**With icons:**
```tsx
<Alert variant="destructive">
  <Icon name="alert-circle" />
  <Alert.Title>Error</Alert.Title>
  <Alert.Description>Invalid credentials</Alert.Description>
</Alert>
```

**Dismissible alerts:**
```tsx
<Alert>
  <Alert.Title>Announcement</Alert.Title>
  <Alert.Description>New features available!</Alert.Description>
  <Button
    variant="ghost"
    size="icon-sm"
    onClick={dismiss}
    class="absolute top-2 right-2"
  >
    <Icon name="x" />
  </Button>
</Alert>
```

**Badge for status:**
```tsx
<Badge variant="default">Active</Badge>
<Badge variant="secondary">Pending</Badge>
<Badge variant="destructive">Suspended</Badge>
<Badge variant="outline">Draft</Badge>
```

---

## 10. Finishing Touches

Small details that elevate designs from good to great.

### Empty States

Every list, table, or data view needs an empty state.

**❌ Poor empty state:**
```tsx
<Table>
  <Table.Body>
    {users.length === 0 && <p>No data</p>}
  </Table.Body>
</Table>
```

**✅ Good empty state:**
```tsx
<Table>
  <Table.Body>
    {users.length === 0 ? (
      <Table.Row>
        <Table.Cell colSpan={4} class="h-32 text-center">
          <div class="flex flex-col items-center gap-2">
            <Icon name="users" class="size-8 text-muted-foreground" />
            <p class="font-medium">No users yet</p>
            <p class="text-muted-foreground text-sm">
              Get started by inviting team members
            </p>
            <Button variant="outline" size="sm" class="mt-2">
              <Icon name="plus" />
              Invite User
            </Button>
          </div>
        </Table.Cell>
      </Table.Row>
    ) : (
      users.map((user) => (/* User rows */))
    )}
  </Table.Body>
</Table>
```

**Empty state components:**
```tsx
// Reusable empty state
function EmptyState({
  icon,
  title,
  description,
  action,
}: {
  icon: IconName
  title: string
  description?: string
  action?: JSX.Element
}) {
  return (
    <div class="flex flex-col items-center justify-center gap-3 py-12 text-center">
      <Icon name={icon} class="size-12 text-muted-foreground" />
      <div>
        <p class="font-semibold text-base">{title}</p>
        {description && (
          <p class="text-muted-foreground text-sm mt-1">{description}</p>
        )}
      </div>
      {action}
    </div>
  )
}

// Usage:
<EmptyState
  icon="inbox"
  title="No messages"
  description="When someone sends you a message, it will appear here"
  action={<Button variant="outline">Compose Message</Button>}
/>
```

### Loading States with Skeleton

Use the Skeleton component for loading placeholders:

```tsx
// Instead of spinners, show content structure
<Card>
  <Card.Header>
    <Skeleton class="h-6 w-32" />  {/* Title placeholder */}
    <Skeleton class="h-4 w-48" />  {/* Description placeholder */}
  </Card.Header>
  <Card.Content class="space-y-3">
    <Skeleton class="h-4 w-full" />
    <Skeleton class="h-4 w-3/4" />
    <Skeleton class="h-4 w-5/6" />
  </Card.Content>
</Card>
```

**Table loading:**
```tsx
<Table>
  <Table.Header>{/* Real headers */}</Table.Header>
  <Table.Body>
    {isLoading
      ? Array.from({ length: 5 }).map((_, i) => (
          <Table.Row key={i}>
            <Table.Cell><Skeleton class="h-4 w-24" /></Table.Cell>
            <Table.Cell><Skeleton class="h-4 w-32" /></Table.Cell>
            <Table.Cell><Skeleton class="h-4 w-16" /></Table.Cell>
          </Table.Row>
        ))
      : users.map((user) => (/* Real rows */))}
  </Table.Body>
</Table>
```

### Use Fewer Borders

Borders aren't the only way to separate content. Consider alternatives:

**❌ Border overload:**
```tsx
<div class="border">
  <div class="border-b p-4">Section 1</div>
  <div class="border-b p-4">Section 2</div>
  <div class="border-b p-4">Section 3</div>
  <div class="p-4">Section 4</div>
</div>
```

**✅ Use spacing, background color, or shadows:**
```tsx
// Option 1: Just spacing
<div class="space-y-6">
  <div>Section 1</div>
  <div>Section 2</div>
  <div>Section 3</div>
</div>

// Option 2: Alternating backgrounds
<div>
  <div class="p-4">Section 1</div>
  <div class="p-4 bg-muted/30">Section 2</div>
  <div class="p-4">Section 3</div>
</div>

// Option 3: Shadow for elevation
<div class="space-y-4">
  <Card class="shadow-sm">Section 1</Card>
  <Card class="shadow-sm">Section 2</Card>
</div>
```

**When borders are good:**
- Tables (to separate data clearly)
- Form inputs (to show boundaries)
- Cards (subtle definition)

**When borders are bad:**
- Everywhere (creates visual noise)
- List items (spacing is often better)
- Navigation items (use background on hover instead)

### Supercharge Defaults

Add subtle details to enhance default states:

```tsx
// Basic button (fine)
<Button>Click me</Button>

// Enhanced with transition
<Button class="transition-all duration-150">Click me</Button>

// Basic input (fine)
<Input />

// Enhanced with selection color
<Input class="selection:bg-primary selection:text-primary-foreground" />

// Basic link (works)
<a href="/page">Link</a>

// Enhanced with underline animation
<a href="/page" class="underline-offset-4 hover:underline transition-all">
  Link
</a>
```

**Subtle improvements:**
```tsx
// Smooth transitions
transition-all           // All properties
transition-colors        // Just colors (more performant)
transition-transform     // Just transforms

duration-150            // Fast (clicks, hovers)
duration-300            // Medium (modals, dropdowns)

// Focus enhancements
focus-visible:outline-none
focus-visible:ring-2
focus-visible:ring-offset-2

// Hover effects
hover:scale-105          // Subtle lift
hover:-translate-y-0.5   // Subtle rise
hover:shadow-md          // Elevation change
```

### Think Outside the Database

Display data in user-friendly formats, not raw database values.

**❌ Database formats:**
```tsx
<p>Status: active</p>
<p>Created: 2024-01-15T10:30:00Z</p>
<p>Price: 1299</p>
```

**✅ User-friendly formats:**
```tsx
<Badge variant="success">Active</Badge>
<p class="text-sm text-muted-foreground">Created Jan 15, 2024</p>
<p class="text-2xl font-semibold">$12.99</p>
```

**Format helpers:**
```tsx
// Dates
new Date('2024-01-15').toLocaleDateString('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric'
})  // "Jan 15, 2024"

// Relative time
function timeAgo(date: Date) {
  const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000)
  if (seconds < 60) return 'just now'
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

// Currency
new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD'
}).format(12.99)  // "$12.99"

// File sizes
function formatBytes(bytes: number) {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}
```

---

## Appendices

### A. Component-to-Principle Map

**Button** → Visual Hierarchy (variants), Interactive States, Accessibility
**Form** → Layout & Spacing, Accessibility
**Field** → Accessibility (ARIA), Typography, Spacing
**Input** → Interactive States, Depth (shadows), Responsive Design, Accessibility
**Card** → Layout & Spacing, Depth (elevation), Visual Hierarchy
**Alert** → Color (semantic variants), Visual Hierarchy, Accessibility
**Table** → Typography (dense), Spacing (compact), Responsive Design
**NavMenu** → Responsive Design, Interactive States, Accessibility
**Badge** → Typography (compact), Color (semantic), Visual Hierarchy

### B. Tailwind Utility Cheatsheet

**Spacing (use these values consistently)**
```
p-2, px-2, py-2, gap-2    = 8px   (dense: tables, badges)
p-4, px-4, py-4, gap-4    = 16px  (standard: buttons, related items)
p-6, px-6, py-6, gap-6    = 24px  (loose: forms, cards)
p-8, gap-8                = 32px  (sections)
p-12, gap-12              = 48px  (page-level)
```

**Typography**
```
text-xs      + font-medium    = Labels, badges
text-sm      + font-medium    = Form labels, nav items
text-sm      + font-semibold  = Card titles
text-base    + font-semibold  = Section headings
text-lg      + font-semibold  = Page headings
```

**Colors (semantic tokens)**
```
Primary:      bg-primary, text-primary-foreground
Destructive:  bg-destructive, text-destructive
Success:      text-success, bg-card (alerts)
Warning:      text-warning, bg-card (alerts)
Muted:        text-muted-foreground
Accent:       hover:bg-accent, hover:text-accent-foreground
Background:   bg-background, bg-card, bg-input
Foreground:   text-foreground, text-card-foreground
Borders:      border-input, border-border, border-ring
```

**Interactive States**
```
Focus:     focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]
Hover:     hover:bg-primary/90, hover:bg-accent, hover:text-accent-foreground
Disabled:  disabled:opacity-50 disabled:pointer-events-none
Invalid:   aria-invalid:border-destructive aria-invalid:ring-destructive/20
```

**Shadows**
```
shadow-xs   - Inputs, subtle cards
shadow-sm   - Standard cards, dropdowns
shadow-md   - Modals, emphasized panels
shadow-xl   - Drawers, overlays
```

**Responsive Breakpoints**
```
sm:   640px  - Small tablets
md:   768px  - Tablets
lg:   1024px - Laptops
xl:   1280px - Desktops
```

### C. Design Token Reference

**Spacing Scale**
```typescript
const spacing = {
  '1':  '0.25rem',  // 4px
  '2':  '0.5rem',   // 8px
  '3':  '0.75rem',  // 12px
  '4':  '1rem',     // 16px
  '6':  '1.5rem',   // 24px
  '8':  '2rem',     // 32px
  '12': '3rem',     // 48px
}
```

**Typography Scale**
```typescript
const fontSize = {
  'xs':   ['0.75rem',   { lineHeight: '1rem' }],     // 12px
  'sm':   ['0.875rem',  { lineHeight: '1.25rem' }],  // 14px
  'base': ['1rem',      { lineHeight: '1.5rem' }],   // 16px
  'lg':   ['1.125rem',  { lineHeight: '1.75rem' }],  // 18px
  'xl':   ['1.25rem',   { lineHeight: '1.75rem' }],  // 20px
  '2xl':  ['1.5rem',    { lineHeight: '2rem' }],     // 24px
}

const fontWeight = {
  'medium':   500,
  'semibold': 600,
  'bold':     700,
}
```

**Color Tokens (semantic)**
```typescript
const colors = {
  // Primary brand
  primary: 'var(--primary)',
  'primary-foreground': 'var(--primary-foreground)',

  // Feedback
  destructive: 'var(--destructive)',
  success: 'var(--success)',
  warning: 'var(--warning)',

  // Neutral
  background: 'var(--background)',
  foreground: 'var(--foreground)',
  card: 'var(--card)',
  'card-foreground': 'var(--card-foreground)',

  // Muted
  muted: 'var(--muted)',
  'muted-foreground': 'var(--muted-foreground)',

  // Interactive
  accent: 'var(--accent)',
  'accent-foreground': 'var(--accent-foreground)',

  // Borders
  border: 'var(--border)',
  input: 'var(--input)',
  ring: 'var(--ring)',
}
```

**Shadow Scale**
```typescript
const boxShadow = {
  xs: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  sm: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
}
```

---

## Summary

This guide provides a comprehensive framework for building consistent, accessible, and well-designed interfaces using our component library and Tailwind CSS. Key takeaways:

1. **Start with features, not layouts** - Build what's needed, when it's needed
2. **Visual hierarchy** - Use weight, color, and spacing (not just size)
3. **Consistent spacing** - Stick to the scale (2, 4, 6, 8, 12)
4. **Limited typography** - 5-6 sizes, 3 weights maximum
5. **Semantic colors** - Use tokens that communicate meaning
6. **Thoughtful depth** - Shadows and borders serve different purposes
7. **Mobile-first responsive** - Build up from mobile, not down from desktop
8. **Accessibility by default** - Our components include ARIA, focus states, and semantic HTML
9. **Component composition** - Use compound components (Card.Header, Form.Field, etc.)
10. **Polish the details** - Empty states, loading states, and subtle transitions

Remember: **Good design is invisible**. Users shouldn't notice your design choices—they should just find the interface easy to use and pleasant to look at.

Refer to the component files in `workspaces/componentry/src/ui/server/` for implementation details and real-world examples of these principles in action.
