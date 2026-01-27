package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestAddLocalPathCreatesManifestAndSymlink(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillRoot := filepath.Join(t.TempDir(), "foo")
	touchSkill(t, skillRoot)

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"add", skillRoot})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	loaded, err := config.Load(filepath.Join(repo, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(loaded.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(loaded.Skills))
	}
	skill := loaded.Skills[0]
	if skill.Type != "path" {
		t.Fatalf("expected path type, got %s", skill.Type)
	}
	if skill.Origin != skillRoot {
		t.Fatalf("expected origin %q, got %q", skillRoot, skill.Origin)
	}

	assertSymlink(t, filepath.Join(repo, "skills", skill.Name), skillRoot)
}

func TestAddGitRepoReusesIdentityAndUpdatesVersion(t *testing.T) {
	repoRoot := t.TempDir()
	origin := "https://example.com/acme/skills"
	repo := initGitRepoWithSkills(t, repoRoot, origin, []string{"alpha", "beta"}, time.Now().Add(-time.Minute))

	manifestRoot := t.TempDir()
	setWorkingDir(t, manifestRoot)

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"add", repoRoot})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	loaded, err := config.Load(filepath.Join(manifestRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(loaded.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(loaded.Skills))
	}
	version1 := loaded.Skills[0].Version
	if version1 == "" {
		t.Fatalf("expected version")
	}
	for _, skill := range loaded.Skills {
		if skill.Origin != origin {
			t.Fatalf("expected origin %q, got %q", origin, skill.Origin)
		}
		if skill.Version != version1 {
			t.Fatalf("expected version %q, got %q", version1, skill.Version)
		}
	}
	if err := assertNames(loaded.Skills, "alpha", "beta"); err != nil {
		t.Fatalf("names: %v", err)
	}

	readme := filepath.Join(repoRoot, "README.md")
	if err := os.WriteFile(readme, []byte("update"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	commitPaths(t, repo, "update", time.Now(), "README.md")

	cmd, _, _ = newTestCommand()
	cmd.SetArgs([]string{"add", repoRoot})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add again: %v", err)
	}

	loaded, err = config.Load(filepath.Join(manifestRoot, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(loaded.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(loaded.Skills))
	}
	version2 := loaded.Skills[0].Version
	if version2 == version1 {
		t.Fatalf("expected version to change")
	}
	for _, skill := range loaded.Skills {
		if skill.Version != version2 {
			t.Fatalf("expected version %q, got %q", version2, skill.Version)
		}
	}
	if err := assertNames(loaded.Skills, "alpha", "beta"); err != nil {
		t.Fatalf("names: %v", err)
	}
}

func initGitRepoWithSkills(t *testing.T, root string, origin string, names []string, when time.Time) *git.Repository {
	t.Helper()

	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	paths := make([]string, 0, len(names))
	for _, name := range names {
		touchSkill(t, filepath.Join(root, "skills", name))
		paths = append(paths, filepath.Join("skills", name, "SKILL.md"))
	}

	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{origin},
	}); err != nil {
		t.Fatalf("create origin: %v", err)
	}

	commitPaths(t, repo, "init", when, paths...)
	return repo
}

func commitPaths(t *testing.T, repo *git.Repository, message string, when time.Time, paths ...string) {
	t.Helper()

	worktree, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	for _, path := range paths {
		if _, err := worktree.Add(path); err != nil {
			t.Fatalf("add %s: %v", path, err)
		}
	}
	if _, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  when,
		},
		Committer: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  when,
		},
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func assertNames(skills []config.Skill, names ...string) error {
	set := make(map[string]struct{}, len(skills))
	for _, skill := range skills {
		set[skill.Name] = struct{}{}
	}
	for _, name := range names {
		if _, ok := set[name]; !ok {
			return fmt.Errorf("missing %s", name)
		}
	}
	if len(set) != len(names) {
		return fmt.Errorf("expected %d skills, got %d", len(names), len(set))
	}
	return nil
}
