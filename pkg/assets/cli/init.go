package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

//go:embed all:template
var templateFS embed.FS

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Scaffold a .assets.yaml configuration file",
		Long: `Scaffold a default .assets.yaml in the current working directory.

The scaffolded file includes commented sections for JS downloads, JS bundles,
CSS compilation, self-hosted fonts, and SVG sprites, ready to edit and build with:

  go run github.com/go-sum/assets/cli assets
  go run github.com/go-sum/assets/cli sprites`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initAssets()
		},
	}
}

func initAssets() error {
	const target = ".assets.yaml"
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s already exists", target)
	}

	err := fs.WalkDir(templateFS, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		return fmt.Errorf("scaffold %s: %w", target, err)
	}

	fmt.Printf("created %s\n", target)
	fmt.Println("next steps:")
	fmt.Println("  edit .assets.yaml to configure your JS, CSS, fonts, and sprites")
	fmt.Println("  go run github.com/go-sum/assets/cli assets")
	fmt.Println("  go run github.com/go-sum/assets/cli sprites")
	return nil
}
