package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// syncProdMod regenerates go.prod.mod from go.mod (stripping replace directives)
// and runs go mod tidy to verify all dependencies resolve.
// Note: git credentials must be configured before calling this function
// (e.g. via git config url.insteadOf) for private module resolution.
func syncProdMod(repoRoot string) error {
	goModPath := filepath.Join(repoRoot, "go.mod")
	prodModPath := filepath.Join(repoRoot, "go.prod.mod")

	if err := copyModStripReplace(goModPath, prodModPath); err != nil {
		return fmt.Errorf("write go.prod.mod: %w", err)
	}
	fmt.Fprintln(logWriter, "wrote go.prod.mod (replace directives removed)")

	// Run go mod tidy on the production mod file to resolve all dependencies.
	cmd := exec.Command("go", "mod", "tidy", "-modfile=go.prod.mod")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"GOWORK=off",
		"GONOSUMDB=github.com/go-sum/*",
		"GOPRIVATE=github.com/go-sum/*",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy -modfile=go.prod.mod: %w", err)
	}
	fmt.Fprintln(logWriter, "go.prod.mod tidied successfully")

	return nil
}

func newSyncCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Copy go.mod to go.prod.mod with replace directives removed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncProdMod(cfg.repoRoot)
		},
	}
}

// copyModStripReplace copies src go.mod to dst, removing all replace directives.
func copyModStripReplace(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	var out []string
	inReplace := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Start of multi-line replace block.
		if trimmed == "replace (" {
			inReplace = true
			continue
		}
		// End of multi-line replace block.
		if inReplace {
			if trimmed == ")" {
				inReplace = false
			}
			continue
		}
		// Single-line replace directive.
		if strings.HasPrefix(trimmed, "replace ") {
			continue
		}

		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(dst, []byte(strings.Join(out, "\n")+"\n"), 0644)
}
