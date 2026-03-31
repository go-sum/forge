package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newSyncCmd(cfg *config) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Copy go.mod to go.prod.mod with replace directives removed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			goModPath := filepath.Join(cfg.repoRoot, "go.mod")
			prodModPath := filepath.Join(cfg.repoRoot, "go.prod.mod")

			if err := copyModStripReplace(goModPath, prodModPath); err != nil {
				return fmt.Errorf("write go.prod.mod: %w", err)
			}

			fmt.Fprintln(logWriter, "wrote go.prod.mod (replace directives removed)")
			return nil
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
