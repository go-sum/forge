package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newExecCmd(cfg *config) *cobra.Command {
	var (
		jobs               int
		includes, excludes []string
		excludePkgs        []string
	)

	cmd := &cobra.Command{
		Use:   "exec [flags] -- <command> [args...]",
		Short: "Run a command in each workspace module directory",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			modules, err := parseWorkFile(cfg.workFile)
			if err != nil {
				return err
			}
			modules = filterModules(modules, includes, excludes)
			if len(modules) == 0 {
				fmt.Fprintln(os.Stderr, "no modules matched filters")
				return nil
			}
			return runExec(cmd.Context(), cfg, modules, args, jobs, excludePkgs)
		},
	}

	// cmd.Flags().IntVarP(&jobs, "jobs", "j", 4, "max parallel modules")
	cmd.Flags().IntVarP(&jobs, "jobs", "j", runtime.NumCPU(), "max parallel modules (default: num CPUs)")
	cmd.Flags().StringSliceVarP(&includes, "include", "i", nil, "include only modules matching substring")
	cmd.Flags().StringSliceVarP(&excludes, "exclude", "e", nil, "exclude modules matching substring")
	cmd.Flags().StringSliceVarP(&excludePkgs, "exclude-pkg", "x", nil, "exclude Go packages matching substring (replaces ./... with filtered list)")

	return cmd
}

// shellCmd joins args into a single shell command string.
func shellCmd(args []string) string {
	return strings.Join(args, " ")
}

// expandArgs replaces "./..." in args with a filtered package list when
// excludePkgs is non-empty. Runs "go list ./..." in the module directory,
// then removes any package whose import path contains an exclude substring.
func expandArgs(ctx context.Context, dir string, args []string, excludePkgs []string) ([]string, error) {
	if len(excludePkgs) == 0 {
		return args, nil
	}

	idx := -1
	for i, a := range args {
		if a == "./..." {
			idx = i
			break
		}
	}
	if idx == -1 {
		return args, nil
	}

	list := exec.CommandContext(ctx, "go", "list", "./...")
	list.Dir = dir
	out, err := list.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	var pkgs []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if matchesAny(line, excludePkgs) {
			continue
		}
		pkgs = append(pkgs, line)
	}

	expanded := make([]string, 0, len(args)-1+len(pkgs))
	expanded = append(expanded, args[:idx]...)
	expanded = append(expanded, pkgs...)
	expanded = append(expanded, args[idx+1:]...)
	return expanded, nil
}

// prefixWriter is an io.Writer that prefixes each complete line with a label.
// A shared mutex ensures lines from concurrent modules never interleave.
type prefixWriter struct {
	mu     *sync.Mutex
	out    io.Writer
	prefix []byte
	buf    []byte // partial line not yet terminated by \n
}

// newPrefixWriter creates a writer that prefixes each line with "[label]"
// padded to width characters.
func newPrefixWriter(mu *sync.Mutex, out io.Writer, label string, width int) *prefixWriter {
	prefix := fmt.Sprintf("%-*s ", width, fmt.Sprintf("[%s]", label))
	return &prefixWriter{
		mu:     mu,
		out:    out,
		prefix: []byte(prefix),
	}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	total := len(p)
	for len(p) > 0 {
		nl := bytes.IndexByte(p, '\n')
		if nl == -1 {
			// No newline — buffer the partial line.
			w.buf = append(w.buf, p...)
			break
		}

		// Complete line: buf + p[:nl].
		line := p[:nl]
		w.mu.Lock()
		w.out.Write(w.prefix)
		if len(w.buf) > 0 {
			w.out.Write(w.buf)
			w.buf = w.buf[:0]
		}
		w.out.Write(line)
		w.out.Write([]byte{'\n'})
		w.mu.Unlock()

		p = p[nl+1:]
	}
	return total, nil
}

// Flush writes any remaining partial line.
func (w *prefixWriter) Flush() {
	if len(w.buf) > 0 {
		w.mu.Lock()
		w.out.Write(w.prefix)
		w.out.Write(w.buf)
		w.out.Write([]byte{'\n'})
		w.mu.Unlock()
		w.buf = w.buf[:0]
	}
}

func runExec(ctx context.Context, cfg *config, modules []string, args []string, jobs int, excludePkgs []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Compute padding width for aligned prefixes.
	maxLen := 0
	for _, mod := range modules {
		if n := len(mod) + 2; n > maxLen { // +2 for [ ]
			maxLen = n
		}
	}

	var mu sync.Mutex
	errs := make([]error, len(modules))
	writers := make([]*prefixWriter, len(modules))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(jobs)

	for i, mod := range modules {
		pw := newPrefixWriter(&mu, os.Stdout, mod, maxLen)
		writers[i] = pw

		g.Go(func() error {
			dir := filepath.Join(cfg.repoRoot, mod)

			modArgs, err := expandArgs(gctx, dir, args, excludePkgs)
			if err != nil {
				fmt.Fprintf(pw, "error expanding packages: %v\n", err)
				errs[i] = err
				return nil
			}

			cmdStr := shellCmd(modArgs)
			cmd := exec.CommandContext(gctx, "sh", "-c", cmdStr)
			cmd.Dir = dir
			cmd.Stdout = pw
			cmd.Stderr = pw
			errs[i] = cmd.Run()
			return nil // never cancel siblings
		})
	}

	_ = g.Wait()

	// Flush any partial lines.
	for _, pw := range writers {
		pw.Flush()
	}

	var failed []string
	for i, err := range errs {
		if err != nil {
			failed = append(failed, modules[i])
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("FAIL: %s", strings.Join(failed, ", "))
	}
	return nil
}
