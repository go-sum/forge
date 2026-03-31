package main

import (
	"fmt"
	"time"

	"github.com/splitsh/lite/splitter"
)

// splitSubtree performs a subtree split for the given prefix using splitsh/lite
// and returns the resulting commit SHA.
func splitSubtree(repoRoot, prefix string) (string, error) {
	config := &splitter.Config{
		Path:       repoRoot,
		Origin:     "HEAD",
		Prefixes:   []*splitter.Prefix{splitter.NewPrefix(prefix, "", nil)},
		GitVersion: "latest",
	}

	result := &splitter.Result{}

	if err := splitter.Split(config, result); err != nil {
		return "", fmt.Errorf("split %s: %w", prefix, err)
	}

	if result.Head() == nil {
		return "", fmt.Errorf("split %s: no commits produced", prefix)
	}

	fmt.Fprintf(logWriter, "  split %s: %d created, %d traversed in %s\n",
		prefix, result.Created(), result.Traversed(), result.Duration(time.Millisecond))

	return result.Head().String(), nil
}
