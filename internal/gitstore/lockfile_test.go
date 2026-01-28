package gitstore

import (
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestResolveRevisionMovedTagStrict(t *testing.T) {
	version := "v1.0.0"
	repoDir, repo, wt, commitHash := createTaggedRepo(t, version)

	writeFile(t, repoDir, "README.md", "v2")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	updated := commit(t, repo, wt, "update")
	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(version), updated)); err != nil {
		t.Fatalf("move tag: %v", err)
	}

	lock := map[manifest.LockKey]string{
		{Origin: repoDir, Version: version}: commitHash.String(),
	}

	_, _, err := ResolveRevision(repoDir, repoDir, version, lock, true)
	if err == nil {
		t.Fatalf("expected moved tag error")
	}
	if !strings.Contains(err.Error(), "moved") {
		t.Fatalf("expected moved tag error, got %v", err)
	}
}

func TestResolveRevisionMovedTagNonStrictUpdatesLock(t *testing.T) {
	version := "v1.0.0"
	repoDir, repo, wt, commitHash := createTaggedRepo(t, version)

	writeFile(t, repoDir, "README.md", "v2")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	updated := commit(t, repo, wt, "update")
	if err := repo.Storer.SetReference(plumbing.NewHashReference(plumbing.NewTagReferenceName(version), updated)); err != nil {
		t.Fatalf("move tag: %v", err)
	}

	key := manifest.LockKey{Origin: repoDir, Version: version}
	lock := map[manifest.LockKey]string{key: commitHash.String()}

	rev, changed, err := ResolveRevision(repoDir, repoDir, version, lock, false)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !changed {
		t.Fatalf("expected lock update")
	}
	if rev != updated.String() {
		t.Fatalf("expected %s, got %s", updated.String(), rev)
	}
	if lock[key] != updated.String() {
		t.Fatalf("expected lock update, got %s", lock[key])
	}
}

func createTaggedRepo(t *testing.T, version string) (string, *git.Repository, *git.Worktree, plumbing.Hash) {
	repoDir := t.TempDir()
	repo := initRepo(t, repoDir)
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	writeFile(t, repoDir, "README.md", "v1")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	commitHash := commit(t, repo, wt, "init")
	if _, err := repo.CreateTag(version, commitHash, nil); err != nil {
		t.Fatalf("tag: %v", err)
	}

	return repoDir, repo, wt, commitHash
}
