package version

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/module"
)

func newTestRepo(t *testing.T) *git.Repository {
	t.Helper()

	root := t.TempDir()
	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	path := filepath.Join(root, "README.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if _, err := worktree.Add("README.md"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := worktree.Commit("init", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}

	return repo
}

func TestResolveForRefEmptyRefIgnoresEOF(t *testing.T) {
	repo := newTestRepo(t)

	resolved, err := ResolveForRef(repo, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.Rev == "" {
		t.Fatalf("expected revision")
	}
}

func TestResolveForRefMissingPrefixIgnoresEOF(t *testing.T) {
	repo := newTestRepo(t)

	_, err := ResolveForRef(repo, "deadbeefdead")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "commit \"deadbeefdead\" not found") {
		t.Fatalf("expected commit not found, got %v", err)
	}
}

func TestResolveForVersionPseudoVersion(t *testing.T) {
	repo := newTestRepo(t)

	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	short := commit.Hash.String()
	if len(short) > shortHashLength {
		short = short[:shortHashLength]
	}
	version := module.PseudoVersion("v0", "", commit.Committer.When.UTC(), short)

	rev, err := ResolveForVersion(repo, version)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if rev != commit.Hash.String() {
		t.Fatalf("expected %s, got %s", commit.Hash.String(), rev)
	}
}
