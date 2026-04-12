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

// config holds CLI-wide state resolved in PersistentPreRunE.
type config struct {
	repoRoot string
	workFile string
}

func main() {
	cfg := &config{}

	root := &cobra.Command{
		Use:   "workspace",
		Short: "Run commands across Go workspace modules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitRepoRoot()
			if err != nil {
				return fmt.Errorf("not inside a git repository: %w", err)
			}
			cfg.repoRoot = root
			cfg.workFile = filepath.Join(root, "go.work")
			return nil
		},
		SilenceUsage: true,
	}

	root.AddCommand(
		newExecCmd(cfg),
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

// parseWorkFile reads a go.work file and returns the list of module paths.
func parseWorkFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open go.work: %w", err)
	}
	defer f.Close()

	var modules []string
	var inUseBlock bool

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if fields[0] == "use" && len(fields) >= 2 && fields[1] == "(" {
			inUseBlock = true
			continue
		}
		if inUseBlock && fields[0] == ")" {
			inUseBlock = false
			continue
		}
		if inUseBlock {
			modules = append(modules, fields[0])
			continue
		}
		if fields[0] == "use" && len(fields) >= 2 {
			modules = append(modules, fields[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read go.work: %w", err)
	}
	return modules, nil
}

// filterModules applies include/exclude substring filters to a list of module paths.
// If includes is non-empty, only modules matching at least one include pattern are kept.
// Then any module matching an exclude pattern is removed.
func filterModules(modules []string, includes, excludes []string) []string {
	var result []string
	for _, mod := range modules {
		if len(includes) > 0 && !matchesAny(mod, includes) {
			continue
		}
		if matchesAny(mod, excludes) {
			continue
		}
		result = append(result, mod)
	}
	return result
}

func matchesAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
