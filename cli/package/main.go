package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// logWriter is where informational messages are written.
var logWriter io.Writer = os.Stderr

// config holds global CLI configuration shared by all subcommands.
type config struct {
	owner    string
	dryRun   bool
	repoRoot string
}

func main() {
	cfg := &config{}

	root := &cobra.Command{
		Use:   "package",
		Short: "Manage subtree-split packages in the forge monorepo",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitRepoRoot()
			if err != nil {
				return fmt.Errorf("not inside a git repository: %w", err)
			}
			cfg.repoRoot = root
			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&cfg.owner, "owner", "go-sum", "GitHub repository owner")
	root.PersistentFlags().BoolVar(&cfg.dryRun, "dry-run", false, "print actions without executing")

	root.AddCommand(
		newListCmd(cfg),
		newPushCmd(cfg),
		newReleaseCmd(cfg),
		newStatusCmd(cfg),
		newSyncCmd(cfg),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// gitRepoRoot returns the root of the current git repository.
func gitRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
