package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func runHashAirCSP() {
	airPath, err := exec.LookPath("air")
	if err != nil {
		fmt.Fprintln(os.Stderr, "air not found in PATH:", err)
		os.Exit(1)
	}

	ver, err := airVersion(airPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read air version:", err)
		os.Exit(1)
	}

	modCache, err := goModCache()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read GOMODCACHE:", err)
		os.Exit(1)
	}

	proxyJS := modCache + "/github.com/air-verse/air@" + ver + "/runner/proxy.js"
	data, err := os.ReadFile(proxyJS)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read proxy.js:", err)
		os.Exit(1)
	}

	sum := sha256.Sum256(data)
	hash := "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"

	fmt.Printf("air %s → %s\n", ver, hash)

	const cfgPath = "config/config.development.yaml"
	updated, err := updateConfig(cfgPath, hash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to update config:", err)
		os.Exit(1)
	}
	if updated {
		fmt.Println("Updated", cfgPath)
	} else {
		fmt.Println("Already up to date:", cfgPath)
	}
}

// airVersion reads the embedded module metadata from the air binary and
// returns the module version string (e.g. "v1.61.7").
func airVersion(airPath string) (string, error) {
	out, err := exec.Command("go", "version", "-m", airPath).Output()
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)
		// Target line: "\tmod\tgithub.com/air-verse/air\t<ver>\t..."
		if len(fields) >= 3 && fields[0] == "mod" && fields[1] == "github.com/air-verse/air" {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("module version not found in: %s", airPath)
}

// goModCache returns the Go module cache directory.
func goModCache() (string, error) {
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var hashRe = regexp.MustCompile(`'sha256-[^']+'`)

// updateConfig replaces the sha256 hash token on the line containing
// "# air proxy hot-reload" in the given YAML file. Returns true if the
// file was changed.
func updateConfig(path, newHash string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(raw), "\n")
	changed := false
	for i, line := range lines {
		if !strings.Contains(line, "# air proxy hot-reload") {
			continue
		}
		replaced := hashRe.ReplaceAllString(line, newHash)
		if replaced != line {
			lines[i] = replaced
			changed = true
		}
		break
	}

	if !changed {
		return false, nil
	}

	return true, os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}
