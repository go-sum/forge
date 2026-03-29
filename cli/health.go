package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/go-sum/forge/internal/health"
)

func runHealth() {
	os.Exit(verifyHealth(os.Args[2:], os.Stdout, os.Stderr))
}

func verifyHealth(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("health", flag.ContinueOnError)
	fs.SetOutput(stderr)

	configDir := fs.String("config-dir", "config", "path to the config directory")
	httpURL := fs.String("http-url", "", "optional URL to verify over HTTP")
	jsonOut := fs.Bool("json", false, "write a JSON report")
	verbose := fs.Bool("verbose", false, "write every verification result")
	timeout := fs.Duration("timeout", 3*time.Second, "per-check timeout")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	report := health.Run(context.Background(), health.Options{
		ConfigDir: *configDir,
		HTTPURL:   *httpURL,
		Timeout:   *timeout,
	})

	if *jsonOut {
		enc := json.NewEncoder(stdout)
		if err := enc.Encode(report); err != nil {
			if err := writeLine(stderr, err); err != nil {
				return 1
			}
			return 1
		}
	} else if out := renderHealthHuman(report, *verbose); out != "" {
		if err := writeLine(stdout, out); err != nil {
			return 1
		}
	}

	if report.HasFailures() {
		return 1
	}
	return 0
}

func renderHealthHuman(report health.Report, verbose bool) string {
	lines := make([]string, 0, len(report.Checks))
	for _, check := range report.Checks {
		if !verbose && check.Status != health.StatusFail {
			continue
		}
		lines = append(lines, formatHealthResult(check, verbose))
	}
	return strings.Join(lines, "\n")
}

func formatHealthResult(result health.Result, verbose bool) string {
	if verbose {
		line := strings.ToUpper(string(result.Status)) + " " + result.Name
		if result.Message != "" {
			line += ": " + result.Message
		}
		return line
	}

	if result.Message == "" {
		return result.Name
	}
	return result.Name + ": " + result.Message
}

func writeLine(w io.Writer, value any) error {
	_, err := fmt.Fprintln(w, value)
	return err
}
