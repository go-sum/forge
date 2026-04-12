package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type verifyOptions struct {
	source string
}

func newVerifyCmd() *cobra.Command {
	var opts verifyOptions

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify that the starter template clones and builds cleanly",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&opts.source, "source", "", "forge repository root (default: auto-detect)")

	return cmd
}

// runVerify clones into a temp directory, then runs go build and go vet.
func runVerify(opts verifyOptions, w io.Writer) error {
	source := opts.source
	if source == "" {
		var err error
		source, err = findSourceRoot()
		if err != nil {
			return err
		}
	}

	tmpDir, err := os.MkdirTemp("", "forge-verify-*")
	if err != nil {
		return fmt.Errorf("verify: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(w, "Cloning into %s...\n", tmpDir)

	cloneOpts := cloneOptions{
		source: source,
		target: tmpDir,
		module: "example.com/verify",
	}
	// The temp dir exists but is empty — remove it so clone doesn't reject it,
	// then re-create it so clone can write into it.
	if err := os.Remove(tmpDir); err != nil {
		return fmt.Errorf("verify: remove temp dir: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("verify: recreate temp dir: %w", err)
	}

	if err := runClone(cloneOpts, w); err != nil {
		return fmt.Errorf("verify: clone failed: %w", err)
	}

	allPassed := true

	// Step: go build ./cmd/server
	fmt.Fprintf(w, "\n[1/2] go build ./cmd/server ... ")
	buildCmd := exec.Command("go", "build", "./cmd/server")
	buildCmd.Dir = tmpDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "FAIL\n")
		if len(buildOut) > 0 {
			fmt.Fprintf(w, "%s\n", buildOut)
		}
		allPassed = false
	} else {
		fmt.Fprintf(w, "PASS\n")
	}

	// Step: go vet ./...
	fmt.Fprintf(w, "[2/2] go vet ./...          ... ")
	vetCmd := exec.Command("go", "vet", "./...")
	vetCmd.Dir = tmpDir
	vetOut, err := vetCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(w, "FAIL\n")
		if len(vetOut) > 0 {
			fmt.Fprintf(w, "%s\n", vetOut)
		}
		allPassed = false
	} else {
		fmt.Fprintf(w, "PASS\n")
	}

	fmt.Fprintln(w)
	if !allPassed {
		fmt.Fprintf(w, "Result: FAIL\n")
		return fmt.Errorf("verify: one or more checks failed")
	}
	fmt.Fprintf(w, "Result: PASS\n")
	return nil
}
