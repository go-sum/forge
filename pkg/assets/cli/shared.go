package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

func command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func closeOnReturn(errp *error, closer io.Closer, subject string, args ...any) {
	if closeErr := closer.Close(); closeErr != nil && *errp == nil {
		*errp = fmt.Errorf("close "+subject+": %w", append(args, closeErr)...)
	}
}

// resolveVersion returns the version from the {NAME}_VERSION environment
// variable (e.g. HTMX_VERSION), falling back to the value in .assets.yaml.
func resolveVersion(name, defaultVersion string) string {
	if v := strings.TrimSpace(os.Getenv(strings.ToUpper(name) + "_VERSION")); v != "" {
		return v
	}
	return defaultVersion
}
