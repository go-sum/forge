package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:           "dev",
		Short:         "Start the development server via air",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			airCmd := commandContext(ctx, "air", "-c", ".air.toml")
			if err := airCmd.Start(); err != nil {
				return fmt.Errorf("starting air: %w", err)
			}

			if err := waitNamed("air", airCmd); err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
			return nil
		},
	}
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func commandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func waitNamed(name string, cmd *exec.Cmd) error {
	if err := cmd.Wait(); err != nil {
		if isExpectedExit(err) {
			return nil
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func isExpectedExit(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return false
	}
	return status.Signaled() && (status.Signal() == syscall.SIGTERM || status.Signal() == syscall.SIGKILL)
}
