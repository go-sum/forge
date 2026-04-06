package pgstore

import (
	"embed"
	"fmt"
	"strings"
)

// SQL files are co-located with the adapter and embedded at compile time.
// schema.sql is the canonical source for the users table.
// queries.sql contains all DML, split by -- name: annotations.

//go:embed sql/schema.sql
var createTableSQL string

//go:embed sql/queries.sql
var rawQueriesSQL string

// AllSQL embeds every .sql file for tooling or inspection.
//
//go:embed sql/*.sql
var AllSQL embed.FS

// Parsed query variables, populated by init().
var (
	createUserSQL      string
	getUserByIDSQL     string
	getUserByEmailSQL  string
	updateUserEmailSQL string
)

func init() {
	queries := parseNamedQueries(rawQueriesSQL)
	mustGet := func(name string) string {
		q, ok := queries[name]
		if !ok {
			panic(fmt.Sprintf("pgstore: missing query %q in queries.sql", name))
		}
		return q
	}
	createUserSQL = mustGet("CreateUser")
	getUserByIDSQL = mustGet("GetUserByID")
	getUserByEmailSQL = mustGet("GetUserByEmail")
	updateUserEmailSQL = mustGet("UpdateUserEmail")
}

// parseNamedQueries splits a SQL file by -- name: annotations into a map
// keyed by query name. Lines before the first annotation are discarded.
func parseNamedQueries(raw string) map[string]string {
	queries := make(map[string]string)
	var currentName string
	var buf strings.Builder

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- name:") {
			if currentName != "" {
				queries[currentName] = buf.String()
			}
			// Parse "-- name: CreateUser :one" → "CreateUser"
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				currentName = parts[2]
			}
			buf.Reset()
			continue
		}
		if currentName != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	if currentName != "" {
		queries[currentName] = buf.String()
	}
	return queries
}
