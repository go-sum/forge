package main

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

// Manifest describes the clone rules loaded from manifest.yaml.
type Manifest struct {
	Exclude       []string      `yaml:"exclude"`
	Rename        []RenameRule  `yaml:"rename"`
	ModuleRewrite ModuleRewrite `yaml:"moduleRewrite"`
}

// RenameRule describes a single file rename operation.
type RenameRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// ModuleRewrite holds the source module path to be replaced.
type ModuleRewrite struct {
	From string `yaml:"from"`
}

// loadManifest reads and parses the YAML manifest at the given path.
func loadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("loadManifest: read %s: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("loadManifest: parse %s: %w", path, err)
	}
	return m, nil
}
