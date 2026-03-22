package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: cli <command>")
		fmt.Fprintln(os.Stderr, "commands: build-assets, build-sprites, dev, hash-air-csp, health")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build-assets":
		runBuildAssets()
	case "build-sprites":
		runBuildSVGSprites()
	case "dev":
		runDev()
	case "hash-air-csp":
		runHashAirCSP()
	case "health":
		runHealth()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
