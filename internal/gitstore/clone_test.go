package gitstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
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

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if err := EnsureRepo(cloneDir, repoDir); err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(cloneDir, "plugins", "foo", "SKILL.md"))
	if err != nil {
		t.Fatalf("read clone: %v", err)
	}
	if string(contents) != "v1" {
		t.Fatalf("expected v1, got %q", string(contents))
	}

	writeFile(t, repoDir, "plugins/foo/SKILL.md", "v2")
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	newCommit := commit(t, repo, wt, "update")

	if err := EnsureRepo(cloneDir, repoDir); err != nil {
		t.Fatalf("EnsureRepo update: %v", err)
	}

	cloneRepo, err := git.PlainOpen(cloneDir)
	if err != nil {
		t.Fatalf("open clone: %v", err)
	}
	if _, err := cloneRepo.CommitObject(newCommit); err != nil {
		t.Fatalf("expected fetched commit: %v", err)
	}
}
