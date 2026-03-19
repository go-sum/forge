# Installing `pkg/components`

This component set is intended to be usable outside the starter app, but consumers must install three things explicitly:

1. CSS classes used by the components.
2. The delegated runtime JavaScript hooks used by interactive components.
3. The icon and sprite registry data used by icon-based components.

## 1. CSS contract

Your stylesheet must include the Tailwind utilities used by files under `pkg/components/` plus the custom rules from `static/css/tailwind.css` that support:

- `dialog::backdrop`
- accordion chevron rotation
- theme icon visibility
- checkbox/radio indicator backgrounds
- progress bar fill styling

## 2. JavaScript contract

Your page runtime must include the delegated handlers from `static/js/app.js` for:

- tabs keyboard and activation behavior
- dismissible alerts and toasts
- theme toggling
- dialog open/close hooks
- sidebar toggles via `data-sidebar-toggle` / `data-sidebar-close`
- dropdown outside-click closing

## 3. Registry installation

Use `starter/pkg/components/install` to populate either isolated registries or the package-global defaults.

### Default globals

```go
componentinstall.ApplyDefault(componentinstall.Config{
    PathFunc: assets.Path,
    IconOverrides: myOverrides,
})
```

### Isolated registries

```go
regs := componentinstall.New(componentinstall.Config{
    PathFunc: myPathFunc,
    IconOverrides: myOverrides,
})
```

When you use isolated registries, pair them with `starter/pkg/components/icons/render.PropsForRegistries(...)` instead of the global `PropsFor(...)` helpers.

## Sidebar instances

`layout.Sidebar` is instance-scoped. Give it an `ID` and apply `layout.ToggleAttrs(id)` to the button that opens it.
