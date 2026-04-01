package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <compose-base> <services> <command>",
		Short: "Run a command with docker compose services, cleaning up after",
		Long: `Ensures the named docker compose services are running, executes the
command, then stops any services that were started by this invocation.

Arguments:
  compose-base  Docker compose command prefix (e.g. "docker compose --profile test")
  services      Space-separated service names to ensure are running
  command       The shell command to execute`,
		Args: cobra.ExactArgs(3),
		RunE: runWithSvc,
	}
}

func runWithSvc(cmd *cobra.Command, args []string) error {
	composeBase := args[0]
	services := strings.Fields(args[1])
	userCmd := args[2]

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Record baseline running services.
	baseline, err := runningServices(ctx, composeBase)
	if err != nil {
		return fmt.Errorf("listing baseline services: %w", err)
	}
	slog.Info("baseline services", "services", mapKeys(baseline))

	// 2. Ensure all required services are running.
	var startedNew bool
	running, err := allServicesRunning(ctx, composeBase, services)
	if err != nil {
		return fmt.Errorf("checking services: %w", err)
	}
	if !running {
		if err := startServices(ctx, composeBase, services); err != nil {
			return err
		}
		startedNew = true
	} else {
		slog.Info("all services already running", "services", services)
	}

	// 3. Execute the user command.
	exitCode := executeCommand(ctx, userCmd)

	// 4. Cleanup: stop only services we started (use Background so cleanup
	//    completes even after signal cancellation).
	if startedNew {
		cleanupCtx := context.Background()
		current, err := runningServices(cleanupCtx, composeBase)
		if err != nil {
			slog.Error("failed to list services for cleanup", "error", err)
		} else {
			toStop := servicesToStop(baseline, current)
			if len(toStop) > 0 {
				if err := stopServices(cleanupCtx, composeBase, toStop); err != nil {
					slog.Error("failed to stop services", "error", err)
				}
			}
		}
	}

	// 5. Propagate exit code.
	os.Exit(exitCode)
	return nil // unreachable
}

// executeCommand runs a shell command and returns its exit code.
func executeCommand(ctx context.Context, shellCmd string) int {
	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return exitCodeFromError(err)
	}
	return 0
}

// exitCodeFromError extracts the exit code from an error.
// Returns 0 for nil, the process exit code for ExitError, or 1 for other errors.
func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

// servicesToStop returns services present in current but not in baseline.
func servicesToStop(baseline, current map[string]struct{}) []string {
	var stop []string
	for svc := range current {
		if _, ok := baseline[svc]; !ok {
			stop = append(stop, svc)
		}
	}
	return stop
}

func mapKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
