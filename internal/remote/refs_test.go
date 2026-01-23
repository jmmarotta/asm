package remote

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestListRemoteRefs(t *testing.T) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)

	writeFile(t, repoDir, "plugins/foo/SKILL.md", "# skill")
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	hash := commit(t, repo, wt, "init")
	if _, err := repo.CreateTag("v0.0.1", hash, nil); err != nil {
		t.Fatalf("tag: %v", err)
	}
	if err := wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("feature/foo"), Create: true}); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	writeFile(t, repoDir, "plugins/foo/SKILL.md", "branch")
	if _, err := wt.Add("plugins/foo/SKILL.md"); err != nil {
		t.Fatalf("add branch: %v", err)
	}
	commit(t, repo, wt, "branch")

	refs, err := ListRemoteRefs(repoDir)
	if err != nil {
		t.Fatalf("ListRemoteRefs: %v", err)
	}
	if _, ok := refs.Branches["feature/foo"]; !ok {
		t.Fatalf("expected branch feature/foo")
	}
	if _, ok := refs.Tags["v0.0.1"]; !ok {
		t.Fatalf("expected tag v0.0.1")
	}
}

func initRepo(t *testing.T, dir string) *git.Repository {
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	return repo
}

func commit(t *testing.T, repo *git.Repository, wt *git.Worktree, message string) plumbing.Hash {
	hash, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	return hash
}

func writeFile(t *testing.T, repoDir string, relativePath string, contents string) {
	path := filepath.Join(repoDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
