package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sum/forge/config"
	"github.com/jackc/pgx/v5"
)

// TestRunConfigFailureSkipsDependentChecks verifies that a bad config dir
// causes the config assertion to fail and all downstream checks to skip.
func TestRunConfigFailureSkipsDependentChecks(t *testing.T) {
	prev := config.App
	t.Cleanup(func() { config.App = prev })

	report := Run(context.Background(), Options{
		ConfigDir: filepath.Join(t.TempDir(), "missing"),
	})

	if !report.HasFailures() {
		t.Fatal("Run() unexpectedly succeeded")
	}
	if got := report.Status; got != "error" {
		t.Fatalf("report.Status = %q, want %q", got, "error")
	}
	// assertLoadConfig + assertDSNConfigured + assertConnectDB + assertUsersSchema
	if got := len(report.Checks); got != 4 {
		t.Fatalf("len(report.Checks) = %d, want 4", got)
	}

	if c := report.Checks[0]; c.Name != "assertLoadConfig" || c.Status != StatusFail {
		t.Fatalf("assertLoadConfig = %+v", c)
	}
	if c := report.Checks[1]; c.Name != "assertDSNConfigured" || c.Status != StatusSkip {
		t.Fatalf("assertDSNConfigured = %+v, want skip", c)
	}
	if c := report.Checks[2]; c.Name != "assertConnectDB" || c.Status != StatusSkip {
		t.Fatalf("assertConnectDB = %+v, want skip", c)
	}
	if c := report.Checks[3]; c.Name != "assertUsersSchema" || c.Status != StatusSkip {
		t.Fatalf("assertUsersSchema = %+v, want skip", c)
	}
}

// TestRunDatabaseFailureAndHTTPSuccessAreIndependent verifies that a bad DSN
// fails DB connectivity, skips required relations, but the HTTP assertion
// still runs and passes independently.
func TestRunDatabaseFailureAndHTTPSuccessAreIndependent(t *testing.T) {
	prev := config.App
	t.Cleanup(func() { config.App = prev })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dir := writeHealthConfig(t, "postgres://user:pass@127.0.0.1:1/starter?sslmode=disable&connect_timeout=1")

	report := Run(context.Background(), Options{
		ConfigDir: dir,
		HTTPURL:   server.URL,
		Timeout:   2 * time.Second,
	})

	if !report.HasFailures() {
		t.Fatal("Run() unexpectedly succeeded")
	}
	// assertLoadConfig + assertDSNConfigured + assertConnectDB + assertUsersSchema + assertHTTPRequest
	if got := len(report.Checks); got != 5 {
		t.Fatalf("len(report.Checks) = %d, want 5", got)
	}
	checkWant := []struct {
		name   string
		status Status
	}{
		{"assertLoadConfig", StatusPass},
		{"assertDSNConfigured", StatusPass},
		{"assertConnectDB", StatusFail},
		{"assertUsersSchema", StatusSkip},
		{"assertHTTPRequest", StatusPass},
	}
	for i, want := range checkWant {
		c := report.Checks[i]
		if c.Name != want.name || c.Status != want.status {
			t.Fatalf("Checks[%d] = {Name:%q Status:%q}, want {Name:%q Status:%q}",
				i, c.Name, c.Status, want.name, want.status)
		}
	}
}

func TestVerifyRequiredRelations(t *testing.T) {
	t.Run("passes when both relations exist", func(t *testing.T) {
		db := fakeQuerier{row: fakeRow{values: []string{"users", "passwords"}}}
		if err := VerifyRequiredRelations(context.Background(), db); err != nil {
			t.Fatalf("VerifyRequiredRelations() error = %v", err)
		}
	})

	t.Run("reports missing relations", func(t *testing.T) {
		db := fakeQuerier{row: fakeRow{values: []string{"", "passwords"}}}
		err := VerifyRequiredRelations(context.Background(), db)
		if err == nil {
			t.Fatal("VerifyRequiredRelations() error = nil")
		}
		if !strings.Contains(err.Error(), "missing required relations users") {
			t.Fatalf("VerifyRequiredRelations() error = %v", err)
		}
	})
}

type fakeQuerier struct {
	row fakeRow
}

func (f fakeQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	return f.row
}

type fakeRow struct {
	values []string
	err    error
}

func (f fakeRow) Scan(dest ...any) error {
	if f.err != nil {
		return f.err
	}
	for i := range dest {
		ptr, ok := dest[i].(*string)
		if !ok {
			return errors.New("dest must be *string")
		}
		*ptr = f.values[i]
	}
	return nil
}

func writeHealthConfig(t *testing.T, dsn string) string {
	t.Helper()

	dir := t.TempDir()
	files := map[string]string{
		"config.yaml": `app:
  env: development
  name: starter
database:
  url: ` + dsn + `
log:
  level: info
server:
  host: 0.0.0.0
  port: 8080
  graceful_timeout: 10
  csp: "default-src 'self'; script-src 'self'; style-src 'self'"
  csrf_cookie_name: _csrf
csp_hashes:
  always: []
  dev_only: []
auth:
  jwt:
    secret: "12345678901234567890123456789012"
    issuer: starter
    token_duration: 86400
  session:
    name: _session
    auth_key: "12345678901234567890123456789012"
    encrypt_key: "12345678901234567890123456789012"
    max_age: 86400
    secure: false
`,
		"site.yaml": `site:
  title: starter
`,
		"nav.yaml": `nav:
  brand:
    label: Starter
    href: /
  sections: []
`,
	}

	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("WriteFile(%s) error = %v", name, err)
		}
	}

	return dir
}
