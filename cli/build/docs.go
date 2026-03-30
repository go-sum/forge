package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-sum/componentry/assetconfig"
)

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
