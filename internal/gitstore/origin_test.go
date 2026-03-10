package gitstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestResolveOriginRevisionReplaceMissing(t *testing.T) {
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
	if _, err := repo.CreateTag("v1.0.0", commitHash, nil); err != nil {
		t.Fatalf("tag: %v", err)
	}

	storeDir := t.TempDir()
	replacePath := filepath.Join(t.TempDir(), "missing")
	lock := map[manifest.LockKey]string{}

	resolution, err := ResolveOriginRevision(storeDir, repoDir, "v1.0.0", replacePath, lock, true)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolution.UsingReplace {
		t.Fatalf("expected fallback to store")
	}
	if !strings.Contains(resolution.Warning, "replace path missing") {
		t.Fatalf("expected replace missing warning, got %q", resolution.Warning)
	}
	if resolution.Path != RepoPath(storeDir, repoDir) {
		t.Fatalf("expected store path, got %s", resolution.Path)
	}
}

func TestResolveOriginRevisionReplaceNotUsable(t *testing.T) {
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
	if _, err := repo.CreateTag("v1.0.0", commitHash, nil); err != nil {
		t.Fatalf("tag: %v", err)
	}

	storeDir := t.TempDir()
	replacePath := t.TempDir()
	lock := map[manifest.LockKey]string{}

	resolution, err := ResolveOriginRevision(storeDir, repoDir, "v1.0.0", replacePath, lock, true)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolution.UsingReplace {
		t.Fatalf("expected fallback to store")
	}
	if !strings.Contains(resolution.Warning, "not usable") {
		t.Fatalf("expected replace not usable warning, got %q", resolution.Warning)
	}
	if resolution.Path != RepoPath(storeDir, repoDir) {
		t.Fatalf("expected store path, got %s", resolution.Path)
	}
}

func TestApplyOriginResolutionReplaceMismatch(t *testing.T) {
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
	writeFile(t, repoDir, "README.md", "v2")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	commit(t, repo, wt, "update")

	resolution := OriginResolution{
		Path:         repoDir,
		Rev:          commitHash.String(),
		UsingReplace: true,
		LockChanged:  false,
	}

	warning, err := ApplyOriginResolution(resolution)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if warning == "" {
		t.Fatalf("expected warning")
	}
	if !strings.Contains(warning, "expected") {
		t.Fatalf("expected mismatch warning, got %q", warning)
	}
}

func TestApplyOriginResolutionReplaceMatchNoWarning(t *testing.T) {
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

	resolution := OriginResolution{
		Path:         repoDir,
		Rev:          commitHash.String(),
		UsingReplace: true,
		LockChanged:  false,
	}

	warning, err := ApplyOriginResolution(resolution)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
}

func TestApplyOriginResolutionCheckout(t *testing.T) {
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
	writeFile(t, repoDir, "README.md", "v2")
	if _, err := wt.Add("README.md"); err != nil {
		t.Fatalf("add update: %v", err)
	}
	commit(t, repo, wt, "update")

	resolution := OriginResolution{
		Path:         repoDir,
		Rev:          commitHash.String(),
		UsingReplace: false,
		LockChanged:  false,
	}

	warning, err := ApplyOriginResolution(resolution)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if warning != "" {
		t.Fatalf("expected no warning, got %q", warning)
	}
	head, err := HeadHash(repoDir)
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if head != commitHash.String() {
		t.Fatalf("expected head %s, got %s", commitHash.String(), head)
	}
}

func TestResolveOriginBatchUsesPrefetchedPseudoVersion(t *testing.T) {
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
	commit(t, repo, wt, "init")

	resolved, err := ResolveForRefAt(repoDir, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	origin := filepath.Join(t.TempDir(), "missing-origin")
	lock := map[manifest.LockKey]string{}
	result, err := ResolveOriginBatch(t.TempDir(), []OriginRequest{{
		Origin:  origin,
		Version: resolved.Version,
		Prefetched: &PrefetchedOrigin{
			Path: repoDir,
			Rev:  resolved.Rev,
		},
	}}, lock, ResolveBatchOptions{Strict: true, MaxParallel: 2})
	if err != nil {
		t.Fatalf("resolve batch: %v", err)
	}
	if got := result.Paths[origin]; got != repoDir {
		t.Fatalf("expected repo path %s, got %s", repoDir, got)
	}
	if !result.LockChanged {
		t.Fatalf("expected lock change")
	}
	if got := lock[manifest.LockKey{Origin: origin, Version: resolved.Version}]; got != resolved.Rev {
		t.Fatalf("expected lock rev %s, got %s", resolved.Rev, got)
	}
}

func TestResolveOriginBatchDoesNotReusePrefetchedSemver(t *testing.T) {
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
	if _, err := repo.CreateTag("v1.0.0", commitHash, nil); err != nil {
		t.Fatalf("tag: %v", err)
	}

	origin := filepath.Join(t.TempDir(), "missing-origin")
	if err := os.RemoveAll(origin); err != nil {
		t.Fatalf("remove missing origin: %v", err)
	}
	_, err = ResolveOriginBatch(t.TempDir(), []OriginRequest{{
		Origin:  origin,
		Version: "v1.0.0",
		Prefetched: &PrefetchedOrigin{
			Path: repoDir,
			Rev:  commitHash.String(),
		},
	}}, map[manifest.LockKey]string{}, ResolveBatchOptions{Strict: true, MaxParallel: 2})
	if err == nil {
		t.Fatalf("expected batch resolve error")
	}
	if !strings.Contains(err.Error(), "missing-origin") {
		t.Fatalf("expected missing origin error, got %v", err)
	}
}
