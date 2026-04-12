# CLAUDE.md — Architectural Constitution

> This application is a Go web app built around server-rendered HTML, HTMX
> progressive enhancement, an `internal/`-only source tree, and PostgreSQL.
> It depends on the reusable `github.com/go-sum/*` modules (auth, componentry,
> kv, queue, security, send, server, session, site) as ordinary external
> dependencies. Versions are maintained in `.versions`.

---

## Behavioral Rules (always enforced)

- ONLY do what has been asked — recommend and get approval before any additions
- NEVER create documentation files (`*.md`) unless explicitly requested
- NEVER hardcode API keys, secrets, or credentials in source files
- NEVER commit secrets, credentials, or `.env` files
- ALWAYS validate user input at system boundaries; sanitize file paths (prevent `../` traversal)
- ALWAYS ensure implementations leverage the `github.com/go-sum/security` module
- ALWAYS run tests after making code changes
- ALWAYS trace ALL callers when refactoring Go config structs or YAML mappings
- ALWAYS account for HTML-encoded entities in test assertions for HTML output
- ALWAYS enforce exact-match test assertions — never substring matching
- ALWAYS use LSP (`mcp__gomcp__lsp_*`) ahead of Grep/Glob for Go code navigation
- FALLBACK to Grep only for non-code text or when `gomcp` MCP server is unavailable

---

## Guide Index
> Before writing code, depending on the requirement consult:
- [`DESIGN_GUIDE.md`](.decisions/DESIGN_GUIDE.md): current project architecture, composition root, layer rules,
  persistence ownership, routing, rendering, config, and testing patterns.
- [`UI_GUIDE.md`](.decisions/UI_GUIDE.md): visual design and UI composition guidance.
- [`ECHOV5_API_REFACTOR.md`](.decisions/ECHOV5_API_REFACTOR.md): Echo v5 handler signatures, breaking
  changes from v4, type-safe parameter extraction, and removed APIs.
<!-- monorepo-only-start -->
- If this repository contains a `pkg/` directory, ALSO read
  [`.decisions/MONOREPO_PACKAGES.md`](.decisions/MONOREPO_PACKAGES.md) before planning any change — it
  adds package-ownership rules and package CLI documentation that do not apply to app-only clones.
<!-- monorepo-only-end -->

---

## MCP Server — gomcp (LSP)

Registered in `.mcp.json`. Available in all agents. Prefer over Grep/Glob for Go.

| Tool | Use |
|------|-----|
| `mcp__gomcp__lsp_workspace_symbols` | Find types, functions, interfaces by name |
| `mcp__gomcp__lsp_find_references` | All callers / all implementors |
| `mcp__gomcp__lsp_definition` | Jump to any symbol definition |
| `mcp__gomcp__lsp_document_symbols` | Inventory all symbols in a file |
| `mcp__gomcp__ping` | Verify server availability |

---

## Development Phase Guide

Invoke the right agent for each phase. Each agent reads its paired rules file first.

| Phase | Agent | Rules | When |
|-------|-------|-------|------|
| Analysis & Design | `cc-plan` | `.claude/rules/r-plan.md` | Before any code — layer assignment, architecture |
| Implementation | `cc-dev` | `.claude/rules/r-code.md` | After plan approved — write code in correct layers |
| Testing | `cc-test` | `.claude/rules/r-test.md` | After implementation — happy-path + failure tests |
| Architecture Review | `cc-plan` | `.claude/rules/r-plan.md` | After tests pass — refactor planning |

Agent flow: `cc-plan` → `cc-dev` → `cc-test` → (if issues) back to `cc-plan`

Agents and rules live in `.claude/agents/` and `.claude/rules/`.
