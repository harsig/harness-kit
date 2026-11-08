package gitsource

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// upstream creates a repository with tags v1 and v2 of a.txt on branch main.
func upstream(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	for _, v := range []string{"v1", "v2"} {
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte(v), 0o644); err != nil {
			t.Fatal(err)
		}
		git(t, dir, "add", ".")
		git(t, dir, "commit", "-q", "-m", v)
		git(t, dir, "tag", v)
	}
	return dir
}

func read(t *testing.T, r *Repo) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(r.Dir, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSync(t *testing.T) {
	url := upstream(t)
	cache := t.TempDir()
	ctx := context.Background()

	tests := []struct{ ref, want string }{
		{ref: "v1", want: "v1"},
		{ref: "", want: "v2"},     // default branch
		{ref: "main", want: "v2"}, // branch
	}
	for _, tt := range tests {
		r := &Repo{URL: url, Ref: tt.ref}
		r.SetDir(cache)
		if err := r.Require("setup"); err == nil {
			t.Fatalf("ref %q: Require before Sync succeeded", tt.ref)
		}
		if err := r.Sync(ctx); err != nil {
			t.Fatalf("ref %q: %v", tt.ref, err)
		}
		if err := r.Require("setup"); err != nil {
			t.Fatalf("ref %q: %v", tt.ref, err)
		}
		if got := read(t, r); got != tt.want {
			t.Errorf("ref %q: content %q, want %q", tt.ref, got, tt.want)
		}
	}
}

func TestSyncUpdatesExistingClone(t *testing.T) {
	url := upstream(t)
	r := &Repo{URL: url, Ref: "main"}
	r.SetDir(t.TempDir())
	ctx := context.Background()
	if err := r.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(url, "a.txt"), []byte("v3"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, url, "commit", "-qam", "v3")
	if err := r.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	if got := read(t, r); got != "v3" {
		t.Errorf("content %q, want v3", got)
	}
}

func TestSyncUnknownRef(t *testing.T) {
	r := &Repo{URL: upstream(t), Ref: "nope"}
	r.SetDir(t.TempDir())
	err := r.Sync(context.Background())
	if err == nil || !strings.Contains(err.Error(), `ref "nope" not found`) {
		t.Fatalf("err = %v, want ref not found", err)
	}
	if _, statErr := os.Stat(r.Dir); statErr == nil {
		t.Error("a failed Sync left a clone behind")
	}
}

func TestSetDir(t *testing.T) {
	a := &Repo{URL: "git@github.com:acme/skills.git", Ref: "v1"}
	b := &Repo{URL: "git@github.com:acme/skills.git", Ref: "v2"}
	a.SetDir("/c")
	b.SetDir("/c")
	if !strings.HasPrefix(a.Dir, "/c/skills-") {
		t.Errorf("Dir = %q, want /c/skills-<hash>", a.Dir)
	}
	if a.Dir == b.Dir {
		t.Error("different refs share a clone directory")
	}
}
