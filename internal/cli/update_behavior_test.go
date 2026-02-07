package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestUpdateDefaultAdvancesPseudoVersion(t *testing.T) {
	repoRoot := t.TempDir()
	setWorkingDir(t, repoRoot)

	origin := "https://example.com/acme/skills"
	originPath, initial, latest := setupUpdateRepo(t, "skills/foo", "")

	if err := manifest.Save(filepath.Join(repoRoot, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Origin:  origin,
			Subdir:  "skills/foo",
			Version: initial.Version,
		}},
		Replace: map[string]string{origin: originPath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	loaded, err := manifest.Load(filepath.Join(repoRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if loaded.Skills[0].Version != latest.Version {
		t.Fatalf("expected version %q, got %q", latest.Version, loaded.Skills[0].Version)
	}

	lock, err := manifest.LoadLock(filepath.Join(repoRoot, "skills-lock.json"))
	if err != nil {
		t.Fatalf("load lock: %v", err)
	}
	if lock[manifest.LockKey{Origin: origin, Version: latest.Version}] != latest.Rev {
		t.Fatalf("expected lock rev %q for %q", latest.Rev, latest.Version)
	}
	if _, ok := lock[manifest.LockKey{Origin: origin, Version: initial.Version}]; ok {
		t.Fatalf("expected old lock key removed")
	}
}

func TestUpdateNameAdvancesPinnedSemver(t *testing.T) {
	repoRoot := t.TempDir()
	setWorkingDir(t, repoRoot)

	origin := "https://example.com/acme/skills"
	originPath, _, latest := setupUpdateRepo(t, "skills/foo", "v1.0.0")

	if err := manifest.Save(filepath.Join(repoRoot, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Origin:  origin,
			Subdir:  "skills/foo",
			Version: "v1.0.0",
		}},
		Replace: map[string]string{origin: originPath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"update", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	loaded, err := manifest.Load(filepath.Join(repoRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if loaded.Skills[0].Version != latest.Version {
		t.Fatalf("expected version %q, got %q", latest.Version, loaded.Skills[0].Version)
	}
}

func TestUpdateOriginSelectorAdvancesPinnedSemver(t *testing.T) {
	repoRoot := t.TempDir()
	setWorkingDir(t, repoRoot)

	origin := "https://example.com/acme/skills"
	originPath, _, latest := setupUpdateRepo(t, "skills/foo", "v1.0.0")

	if err := manifest.Save(filepath.Join(repoRoot, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Origin:  origin,
			Subdir:  "skills/foo",
			Version: "v1.0.0",
		}},
		Replace: map[string]string{origin: originPath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"update", origin})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	loaded, err := manifest.Load(filepath.Join(repoRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if loaded.Skills[0].Version != latest.Version {
		t.Fatalf("expected version %q, got %q", latest.Version, loaded.Skills[0].Version)
	}
}

func TestUpdateOriginPathSelectorAdvancesPinnedSemver(t *testing.T) {
	repoRoot := t.TempDir()
	setWorkingDir(t, repoRoot)

	origin := "https://example.com/acme/skills"
	originPath, _, latest := setupUpdateRepo(t, "skills/aglit-workflow", "v1.0.0")

	if err := manifest.Save(filepath.Join(repoRoot, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "aglit-workflow",
			Origin:  origin,
			Subdir:  "skills/aglit-workflow",
			Version: "v1.0.0",
		}},
		Replace: map[string]string{origin: originPath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"update", origin, "--path", "skills/aglit-workflow"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}

	loaded, err := manifest.Load(filepath.Join(repoRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if loaded.Skills[0].Version != latest.Version {
		t.Fatalf("expected version %q, got %q", latest.Version, loaded.Skills[0].Version)
	}
}

func setupUpdateRepo(t *testing.T, subdir string, tag string) (string, gitstore.Resolved, gitstore.Resolved) {
	t.Helper()

	originPath := t.TempDir()
	repo, err := git.PlainInit(originPath, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	now := time.Now().Add(-2 * time.Minute)
	touchSkill(t, filepath.Join(originPath, filepath.FromSlash(subdir)))
	commitPaths(t, repo, "init", now, filepath.Join(filepath.FromSlash(subdir), "SKILL.md"))

	if tag != "" {
		head, err := repo.Head()
		if err != nil {
			t.Fatalf("head: %v", err)
		}
		if _, err := repo.CreateTag(tag, head.Hash(), nil); err != nil {
			t.Fatalf("create tag: %v", err)
		}
	}

	initial, err := gitstore.ResolveForRefAt(originPath, "")
	if err != nil {
		t.Fatalf("resolve initial: %v", err)
	}

	readme := filepath.Join(originPath, "README.md")
	if err := os.WriteFile(readme, []byte("update"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	commitPaths(t, repo, "update", now.Add(time.Minute), "README.md")

	latest, err := gitstore.ResolveForRefAt(originPath, "")
	if err != nil {
		t.Fatalf("resolve latest: %v", err)
	}

	return originPath, initial, latest
}
