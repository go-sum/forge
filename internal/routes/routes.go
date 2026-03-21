// Package routes defines the canonical URL path patterns and builder functions
// for the application. Both the transport layer (handler registration) and the
// view layer (href attributes, form actions) import from here, so neither
// depends on the other.
package routes

import "fmt"

// Route path patterns — used verbatim with Echo's router registration.
// Use the builder functions below when constructing concrete URLs with IDs or
// query parameters.
const (
	Health     = "/health"
	Components = "/_components"
	Home       = "/"

	Login    = "/login"
	Register = "/register"
	Logout   = "/logout"

	Users    = "/users"
	UserEdit = "/users/:id/edit"
	UserRow  = "/users/:id/row"
	UserByID = "/users/:id"
)

// UserPath returns the resource URL for a specific user.
func UserPath(id string) string { return "/users/" + id }

// UserEditPath returns the inline-edit URL for a user.
func UserEditPath(id string) string { return "/users/" + id + "/edit" }

// UserRowPath returns the read-only row URL for a user.
func UserRowPath(id string) string { return "/users/" + id + "/row" }

// UserListPage returns the paginated user-list URL for the given page number.
func UserListPage(page int) string { return fmt.Sprintf("%s?page=%d", Users, page) }
