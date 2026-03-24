package httpsec

import "net/http"

// IsSafeMethod reports whether method is treated as safe by RFC 9110.
func IsSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// IsUnsafeMethod reports whether method should be protected as state-changing.
func IsUnsafeMethod(method string) bool {
	return !IsSafeMethod(method)
}
