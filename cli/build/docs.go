package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-sum/componentry/assetconfig"
	"github.com/spf13/cobra"
)

func newDocsCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Build Hugo documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := assetconfig.Load(configPath)
			if err != nil {
				return err
			}
			return buildDocs(cfg.Paths)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", assetconfig.DefaultConfigPath, "path to assets config file")

	return cmd
}

func buildDocs(paths assetconfig.Paths) error {
	const docsSourceDir = ".docs"

	publicRoot, err := filepath.Abs(paths.PublicRoot())
	if err != nil {
		return fmt.Errorf("resolve public root %s: %w", paths.PublicRoot(), err)
	}
	outputDir := filepath.Join(publicRoot, "doc")
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("remove %s: %w", outputDir, err)
	}
	if err := os.MkdirAll(publicRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", publicRoot, err)
	}

	args := []string{
		"--source", docsSourceDir,
		"--destination", outputDir,
		"--quiet",
	}
	if err := command("hugo", args...).Run(); err != nil {
		return fmt.Errorf("hugo: %w", err)
	}
	return nil
}
