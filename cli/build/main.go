package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: build <command>")
		fmt.Fprintln(os.Stderr, "commands: assets, sprites, all")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "assets":
		runBuildAssets()
	case "sprites":
		runBuildSVGSprites()
	case "all":
		runBuildAssets()
		runBuildSVGSprites()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runBuildAssets() {
	flags, err := parseAssetBuildFlags(os.Args[2:])
	if err != nil {
		os.Exit(1)
	}

	opts, err := resolveAssetBuildOptions(flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if err := buildAssets(opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runBuildSVGSprites() {
	if err := buildSVGSprites(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
