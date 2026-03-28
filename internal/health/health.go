package health

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sum/forge/config"
	"github.com/go-sum/server/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTimeout = 3 * time.Second

type Status string

const (
	StatusPass Status = "pass"
	StatusFail Status = "fail"
	StatusSkip Status = "skip"
)

type Result struct {
	Name       string        `json:"name"`
	Status     Status        `json:"status"`
	Message    string        `json:"message,omitempty"`
	DurationMS int64         `json:"duration_ms"`
	Duration   time.Duration `json:"-"`
}

type Report struct {
	Status string   `json:"status"`
	Checks []Result `json:"checks"`
}

type Options struct {
	ConfigDir string
	HTTPURL   string
	Timeout   time.Duration
}

// DBQuerier is satisfied by *pgxpool.Pool and test fakes.
type DBQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type verificationState struct {
	opts Options
	cfg  *config.Config
	db   *pgxpool.Pool
}

// Run executes all health checks in sequence and returns a report.
// Each check records its own Result; dependent checks skip automatically
// when an upstream check did not pass.
func Run(ctx context.Context, opts Options) Report {
	opts = withDefaults(opts)
	state := &verificationState{opts: opts}
	defer closePool(state.db)

	report := Report{Status: "ok"}

	run := func(name string, fn func(context.Context, *verificationState) Result) {
		start := time.Now()
		result := fn(ctx, state)
		result.Name = name
		result.Duration = time.Since(start)
		result.DurationMS = result.Duration.Milliseconds()
		report.Checks = append(report.Checks, result)
		if result.Status == StatusFail {
			report.Status = "error"
		}
	}

	run("assertLoadConfig", assertLoadConfig)
	run("assertDSNConfigured", assertDSNConfigured)
	run("assertConnectDB", assertConnectDB)
	run("assertUsersSchema", assertUsersSchema)
	if opts.HTTPURL != "" {
		run("assertHTTPRequest", assertHTTPRequest)
	}

	return report
}

func (r Report) HasFailures() bool {
	for _, check := range r.Checks {
		if check.Status == StatusFail {
			return true
		}
	}
	return false
}

// VerifyRequiredRelations checks that the required auth relations exist.
func VerifyRequiredRelations(ctx context.Context, db DBQuerier) error {
	var usersTable string
	err := db.QueryRow(ctx, `
SELECT
    COALESCE(to_regclass('public.users')::text, '')
`).Scan(&usersTable)
	if err != nil {
		return fmt.Errorf("verify schema: %w", err)
	}

	missing := make([]string, 0, 1)
	if usersTable == "" {
		missing = append(missing, "users")
	}
	if len(missing) == 0 {
		return nil
	}

	return fmt.Errorf(
		"missing required relations %s; run `make db-apply` to apply db/sql/schema.sql",
		strings.Join(missing, ", "),
	)
}

func withDefaults(opts Options) Options {
	if strings.TrimSpace(opts.ConfigDir) == "" {
		opts.ConfigDir = "config"
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	return opts
}

func assertLoadConfig(_ context.Context, state *verificationState) Result {
	cfg, err := config.LoadFrom(state.opts.ConfigDir, os.Getenv("APP_ENV"))
	if err != nil {
		return failResult(fmt.Errorf("load config from %s: %w", state.opts.ConfigDir, err))
	}
	state.cfg = cfg
	return passResult("loaded config from " + state.opts.ConfigDir)
}

func assertDSNConfigured(_ context.Context, state *verificationState) Result {
	if state.cfg == nil {
		return skipResult("skipped because config did not load")
	}
	if state.cfg.DSN() == "" {
		return failResult(fmt.Errorf("database.url is not set in config"))
	}
	return passResult("DSN is configured")
}

func assertConnectDB(ctx context.Context, state *verificationState) Result {
	if state.cfg == nil || state.cfg.DSN() == "" {
		return skipResult("skipped because database DSN is not available")
	}

	checkCtx, cancel := context.WithTimeout(ctx, state.opts.Timeout)
	defer cancel()

	pool, err := database.Connect(checkCtx, state.cfg.DSN())
	if err != nil {
		return failResult(err)
	}

	closePool(state.db)
	state.db = pool
	return passResult("connected to " + redactDSN(state.cfg.DSN()))
}

func assertUsersSchema(ctx context.Context, state *verificationState) Result {
	if state.db == nil {
		return skipResult("skipped because database connectivity failed")
	}

	checkCtx, cancel := context.WithTimeout(ctx, state.opts.Timeout)
	defer cancel()

	if err := VerifyRequiredRelations(checkCtx, state.db); err != nil {
		return failResult(err)
	}

	return passResult("required relations are present")
}

func assertHTTPRequest(ctx context.Context, state *verificationState) Result {
	checkCtx, cancel := context.WithTimeout(ctx, state.opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, state.opts.HTTPURL, nil)
	if err != nil {
		return failResult(fmt.Errorf("build request: %w", err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return failResult(fmt.Errorf("GET %s: %w", state.opts.HTTPURL, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return failResult(fmt.Errorf("GET %s: status %d", state.opts.HTTPURL, resp.StatusCode))
	}

	return passResult("reachable at " + state.opts.HTTPURL)
}

func passResult(msg string) Result {
	return Result{Status: StatusPass, Message: msg}
}

func failResult(err error) Result {
	return Result{Status: StatusFail, Message: err.Error()}
}

func skipResult(msg string) Result {
	return Result{Status: StatusSkip, Message: msg}
}

func closePool(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

func redactDSN(dsn string) string {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return "database"
	}

	host := cfg.ConnConfig.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.ConnConfig.Port
	if port == 0 {
		port = 5432
	}
	name := cfg.ConnConfig.Database
	if name == "" {
		name = "postgres"
	}

	return host + ":" + strconv.Itoa(int(port)) + "/" + name
}
