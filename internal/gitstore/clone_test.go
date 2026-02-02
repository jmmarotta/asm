package gitstore

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

func TestResolveForRefPrefersRemoteBranch(t *testing.T) {
	originDir := t.TempDir()
	originRepo := initRepo(t, originDir)

	wt, err := originRepo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	writeFile(t, originDir, "README.md", "v1")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	commitA := commit(t, originRepo, wt, "init")

	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
		Create: true,
	}); err != nil {
		t.Fatalf("checkout main: %v", err)
	}

	cloneDir := filepath.Join(t.TempDir(), "clone")
	if err := EnsureRepo(cloneDir, originDir); err != nil {
		t.Fatalf("EnsureRepo: %v", err)
	}

	cloneRepo, err := git.PlainOpen(cloneDir)
	if err != nil {
		t.Fatalf("open clone: %v", err)
	}
	localMain, err := cloneRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if err != nil {
		t.Fatalf("read local main: %v", err)
	}
	if localMain.Hash().String() != commitA.String() {
		t.Fatalf("expected local main %s, got %s", commitA.String(), localMain.Hash())
	}

	if err := CheckoutRevision(cloneDir, commitA.String()); err != nil {
		t.Fatalf("checkout clone: %v", err)
	}

	writeFile(t, originDir, "README.md", "v2")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	commitB := commit(t, originRepo, wt, "update")

	if err := EnsureRepo(cloneDir, originDir); err != nil {
		t.Fatalf("EnsureRepo update: %v", err)
	}

	cloneRepo, err = git.PlainOpen(cloneDir)
	if err != nil {
		t.Fatalf("open clone updated: %v", err)
	}
	localMain, err = cloneRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if err != nil {
		t.Fatalf("read local main updated: %v", err)
	}
	if localMain.Hash().String() != commitA.String() {
		t.Fatalf("expected local main to stay at %s, got %s", commitA.String(), localMain.Hash())
	}
	remoteMain, err := cloneRepo.Reference(plumbing.NewRemoteReferenceName("origin", "main"), true)
	if err != nil {
		t.Fatalf("read remote main: %v", err)
	}
	if remoteMain.Hash().String() != commitB.String() {
		t.Fatalf("expected remote main %s, got %s", commitB.String(), remoteMain.Hash())
	}

	resolved, err := ResolveForRefAt(cloneDir, "main")
	if err != nil {
		t.Fatalf("resolve main: %v", err)
	}
	if resolved.Rev != commitB.String() {
		t.Fatalf("expected main to resolve to %s, got %s", commitB.String(), resolved.Rev)
	}

	resolvedLocal, err := ResolveForRefAt(cloneDir, "refs/heads/main")
	if err != nil {
		t.Fatalf("resolve refs/heads/main: %v", err)
	}
	if resolvedLocal.Rev != commitA.String() {
		t.Fatalf("expected refs/heads/main to resolve to %s, got %s", commitA.String(), resolvedLocal.Rev)
	}
}
