// Package gitsource clones git repositories that a configuration file uses as a
// source, and checks that they exist. Only Sync uses the network.
package gitsource

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Repo is a git repository at a ref, cloned into Dir.
type Repo struct {
	URL string `yaml:"url"`
	Ref string `yaml:"ref"` // tag, branch or commit; empty = the remote's default branch
	Dir string `yaml:"-"`   // clone location, set by SetDir
}

var slugUnsafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// SetDir places the clone in cacheDir. The directory name contains a hash of
// URL and ref, so after a ref change the clone is missing until Sync runs.
func (r *Repo) SetDir(cacheDir string) {
	name := strings.TrimSuffix(path.Base(strings.TrimRight(r.URL, "/")), ".git")
	sum := sha256.Sum256([]byte(r.URL + "@" + r.Ref))
	r.Dir = filepath.Join(cacheDir, slugUnsafe.ReplaceAllString(name, "-")+"-"+hex.EncodeToString(sum[:4]))
}

// Validate checks the configured fields.
func (r *Repo) Validate() error {
	if r.URL == "" {
		return fmt.Errorf("git.url is required")
	}
	return nil
}

func (r *Repo) String() string {
	if r.Ref == "" {
		return r.URL
	}
	return r.URL + "@" + r.Ref
}

// Require fails if the repository is not cloned yet; hint is the command that clones it.
func (r *Repo) Require(hint string) error {
	if _, err := os.Stat(r.Dir); err != nil {
		return fmt.Errorf("git source %s is not set up; run: %s", r, hint)
	}
	return nil
}

// Sync clones a missing repository, or fetches an existing one, and checks out
// the ref as a detached HEAD. git output goes to stderr.
func (r *Repo) Sync(ctx context.Context) error {
	if err := r.sync(ctx); err != nil {
		return fmt.Errorf("git source %s: %w", r, err)
	}
	return nil
}

func (r *Repo) sync(ctx context.Context) error {
	if _, err := os.Stat(r.Dir); err == nil {
		if err := runGit(ctx, r.Dir, "fetch", "--tags", "--force", "--prune", "origin"); err != nil {
			return err
		}
		return checkout(ctx, r.Dir, r.Ref)
	}
	// Clone aside and rename, so a failed checkout never leaves a usable-looking clone.
	tmp := r.Dir + ".tmp"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.Dir), 0o755); err != nil {
		return err
	}
	if err := runGit(ctx, "", "clone", "--", r.URL, tmp); err != nil {
		return err
	}
	if err := checkout(ctx, tmp, r.Ref); err != nil {
		os.RemoveAll(tmp)
		return err
	}
	return os.Rename(tmp, r.Dir)
}

// checkout detaches HEAD at ref, preferring the remote branch over a local name.
func checkout(ctx context.Context, dir, ref string) error {
	candidates := []string{"origin/HEAD"}
	if ref != "" {
		candidates = []string{"origin/" + ref, ref}
	}
	for _, c := range candidates {
		out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--verify", "--quiet", c+"^{commit}").Output()
		if err != nil {
			continue
		}
		return runGit(ctx, dir, "checkout", "--detach", "--force", "--quiet", strings.TrimSpace(string(out)))
	}
	return fmt.Errorf("ref %q not found", ref)
}

// runGit runs git in dir; stdout goes to stderr because the tools' stdout carries their protocol or answer.
func runGit(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stderr, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
