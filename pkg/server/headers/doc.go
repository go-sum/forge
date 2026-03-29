// Package headers provides typed parsing, construction, and round-trip
// serialisation for common HTTP headers: Accept-Language, Accept,
// Cache-Control, and Vary.
//
// All parsers accept raw header string values as returned by
// [net/http.Header.Get] and return typed values with structured accessor
// methods and [fmt.Stringer] implementations for round-trip serialisation.
//
// This package depends only on the Go standard library. It contains no
// framework types and is safe to import from any pkg/ package.
package headers
