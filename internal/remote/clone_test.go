package remote

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func TestEnsureRepoClonesAndUpdates(t *testing.T) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	writeFile(t, repoDir, "plugins/foo/SKILL.md", "v1")
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	commit(t, repo, wt, "init")

	if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("feature/foo"), Create: true}); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	writeFile(t, repoDir, "plugins/foo/SKILL.md", "branch")
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add branch: %v", err)
	}
	commit(t, repo, wt, "branch")

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if err := EnsureRepo(cloneDir, repoDir, "feature/foo"); err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(cloneDir, "plugins", "foo", "SKILL.md"))
	if err != nil {
		t.Fatalf("read clone: %v", err)
	}
	if string(contents) != "branch" {
		t.Fatalf("expected branch, got %q", string(contents))
	}

	if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("feature/foo")}); err != nil {
		t.Fatalf("checkout branch: %v", err)
	}
	writeFile(t, repoDir, "plugins/foo/SKILL.md", "v2")
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	commit(t, repo, wt, "update")

	if err := EnsureRepo(cloneDir, repoDir, "feature/foo"); err != nil {
		t.Fatalf("EnsureRepo update: %v", err)
	}

	updated, err := os.ReadFile(filepath.Join(cloneDir, "plugins", "foo", "SKILL.md"))
	if err != nil {
		t.Fatalf("read clone update: %v", err)
	}
	if string(updated) != "v2" {
		t.Fatalf("expected v2, got %q", string(updated))
	}
}
