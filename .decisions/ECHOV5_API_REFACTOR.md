# ECHOV5_API_REFACTOR.md — Echo v5 Critical API Rules

> Transport-specific rules for Echo v5. Use alongside
> [`DESIGN_GUIDE.md`](./DESIGN_GUIDE.md) for architecture and ownership, and
> [`PATTERNS_PRINCIPLES.md`](./PATTERNS_PRINCIPLES.md) for code-structure and
> maintainability rules.

---

## Handler Signatures — BREAKING CHANGE from v4

```go
// ✅ v5: pointer to concrete struct
func MyHandler(c *echo.Context) error

// ❌ v4 (WRONG — will not compile): interface
func MyHandler(c echo.Context) error
```

---

## Import Paths

```go
import (
    "github.com/labstack/echo/v5"
    "github.com/labstack/echo/v5/middleware"
)
```

---

## Type-Safe Parameter Extraction

```go
// Path parameters
id, err := echo.PathParam[int](c, "id")           // returns (int, error)
id, err := echo.PathParamOr[int](c, "id", 0)      // with default
raw := c.Param("id")                               // string, no conversion

// Query parameters
page, err := echo.QueryParam[int](c, "page")
page, err := echo.QueryParamOr[int](c, "page", 1)
ids, err  := echo.QueryParams[int](c, "ids")       // []int from repeated param

// Form values (RENAMED from FormParam* in v4)
name := c.FormValue("name")                         // string
age, err := echo.FormValue[int](c, "age")
tags, err := echo.FormValues[string](c, "tags")

// Type-safe context get
user, err := echo.ContextGet[User](c, "user")
```

---

## Request Body Binding

```go
echo.BindBody(c, &payload)
echo.BindHeaders(c, &headers)
echo.BindQueryParams(c, &query)
```

---

## Response Methods

```go
c.String(http.StatusOK, "text")
c.JSON(http.StatusOK, data)
c.HTML(http.StatusOK, "<h1>Hi</h1>")
c.NoContent(http.StatusNoContent)
c.Redirect(http.StatusSeeOther, "/target")
c.File("path/to/file")
```

`c.Response()` returns `http.ResponseWriter` directly. Use
`c.UnwrapResponse()` for the Echo response wrapper.

---

## Error Handling

```go
// HTTPError.Message is now string (was interface{} in v4)
e.HTTPErrorHandler = echo.DefaultHTTPErrorHandler(cfg.Debug)

// Pre-defined errors
echo.ErrBadRequest          // 400
echo.ErrUnauthorized        // 401
echo.ErrForbidden           // 403
echo.ErrNotFound            // 404
echo.ErrMethodNotAllowed    // 405
echo.ErrInternalServerError // 500
```

---

## Removed APIs (v4 → v5)

`ParamNames()`, `ParamValues()`, `SetParamNames()`, `SetParamValues()` —
removed from Context. Use type-safe extraction instead.
