package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

// runningServices returns the set of currently running service names.
func runningServices(ctx context.Context, compose string) (map[string]struct{}, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", compose+" ps --status running --services 2>/dev/null")
	out, err := cmd.Output()
	if err != nil {
		return make(map[string]struct{}), nil
	}
	services := make(map[string]struct{})
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if s := strings.TrimSpace(line); s != "" {
			services[s] = struct{}{}
		}
	}
	return services, nil
}

// allServicesRunning checks if all named services have running containers.
func allServicesRunning(ctx context.Context, compose string, services []string) (bool, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c",
		compose+" ps --status running -q "+strings.Join(services, " ")+" 2>/dev/null")
	out, err := cmd.Output()
	if err != nil {
		return false, nil
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// startServices starts the listed services and waits for health checks.
func startServices(ctx context.Context, compose string, services []string) error {
	slog.Info("starting services", "services", services)
	cmd := exec.CommandContext(ctx, "sh", "-c",
		compose+" up -d --wait "+strings.Join(services, " "))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("starting services %v: %w", services, err)
	}
	return nil
}

// stopServices force-removes the listed services.
func stopServices(ctx context.Context, compose string, services []string) error {
	slog.Info("stopping services started by this run", "services", services)
	cmd := exec.CommandContext(ctx, "sh", "-c",
		compose+" rm -fs "+strings.Join(services, " "))
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stopping services %v: %w", services, err)
	}
	return nil
}
