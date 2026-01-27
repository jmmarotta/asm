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

	sum := map[manifest.SumKey]string{
		{Origin: repoDir, Version: version}: commitHash.String(),
	}

	_, _, err := ResolveRevision(repoDir, repoDir, version, sum, true)
	if err == nil {
		t.Fatalf("expected moved tag error")
	}
	if !strings.Contains(err.Error(), "moved") {
		t.Fatalf("expected moved tag error, got %v", err)
	}
}

func TestResolveRevisionMovedTagNonStrictUpdatesSum(t *testing.T) {
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

	key := manifest.SumKey{Origin: repoDir, Version: version}
	sum := map[manifest.SumKey]string{key: commitHash.String()}

	rev, changed, err := ResolveRevision(repoDir, repoDir, version, sum, false)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !changed {
		t.Fatalf("expected sum update")
	}
	if rev != updated.String() {
		t.Fatalf("expected %s, got %s", updated.String(), rev)
	}
	if sum[key] != updated.String() {
		t.Fatalf("expected sum update, got %s", sum[key])
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
