package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v68/github"
)

const tokenEnvVar = "GITHUB_ACCESS_TOKEN"

// ghClient wraps the go-github client with convenience methods.
type ghClient struct {
	client *github.Client
	token  string
	owner  string
	dryRun bool
}

// newGHClient creates a GitHub client using the GITHUB_ACCESS_TOKEN env var.
func newGHClient(owner string, dryRun bool) (*ghClient, error) {
	token := os.Getenv(tokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("%s is required", tokenEnvVar)
	}

	client := github.NewClient(nil).WithAuthToken(token)
	return &ghClient{
		client: client,
		token:  token,
		owner:  owner,
		dryRun: dryRun,
	}, nil
}

// repoExists checks whether a mirror repo exists on GitHub.
func (g *ghClient) repoExists(ctx context.Context, repo string) (bool, error) {
	_, resp, err := g.client.Repositories.Get(ctx, g.owner, repo)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("check repo %s/%s: %w", g.owner, repo, err)
	}
	return true, nil
}

// ensureRepoExists verifies the mirror repo exists, returning a helpful error if not.
func ensureRepoExists(ctx context.Context, gh *ghClient, pkg Package) error {
	if gh.dryRun {
		return nil
	}

	exists, err := gh.repoExists(ctx, pkg.MirrorRepo)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("mirror repo %s/%s does not exist on GitHub\n  create it at: https://github.com/organizations/%s/repositories/new?name=%s",
			gh.owner, pkg.MirrorRepo, gh.owner, pkg.MirrorRepo)
	}
	return nil
}

// createRelease creates a GitHub Release on the mirror repo.
func (g *ghClient) createRelease(ctx context.Context, repo, version, prefix string) error {
	notes := fmt.Sprintf("Released from go-sum/forge subtree %s", prefix)

	if g.dryRun {
		fmt.Fprintf(logWriter, "  [dry-run] would create release %s on %s/%s\n", version, g.owner, repo)
		return nil
	}

	_, _, err := g.client.Repositories.CreateRelease(ctx, g.owner, repo, &github.RepositoryRelease{
		TagName: github.Ptr(version),
		Name:    github.Ptr(version),
		Body:    github.Ptr(notes),
	})
	if err != nil {
		return fmt.Errorf("create release %s on %s/%s: %w", version, g.owner, repo, err)
	}
	return nil
}

// getRef returns the SHA of a ref on the mirror repo, or empty string if not found.
func (g *ghClient) getRef(ctx context.Context, repo, refName string) (string, error) {
	ref, resp, err := g.client.Git.GetRef(ctx, g.owner, repo, refName)
	if err != nil {
		if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 409) {
			return "", nil
		}
		return "", fmt.Errorf("get ref %s on %s/%s: %w", refName, g.owner, repo, err)
	}
	return ref.GetObject().GetSHA(), nil
}

// pushGit pushes a local SHA to the mirror repo via git push.
// refs are full ref names like "refs/heads/main" or "refs/tags/v0.1.0".
func pushGit(repoRoot, token, owner, repo, sha string, refs []string, dryRun bool) error {
	remoteURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)

	refSpecs := make([]string, len(refs))
	for i, ref := range refs {
		refSpecs[i] = sha + ":" + ref
	}

	if dryRun {
		safeURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
		fmt.Fprintf(logWriter, "  [dry-run] would push %s to %s %s\n", sha[:12], safeURL, strings.Join(refs, " "))
		return nil
	}

	args := append([]string{"push", remoteURL}, refSpecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push to %s/%s: %w", owner, repo, err)
	}
	return nil
}
