package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestUpdateOutputSkipsPinnedSemverByDefault(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	originPath := t.TempDir()
	initTaggedRepo(t, originPath, "v1.0.0")
	origin := "https://example.com/repo"

	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Origin:  origin,
			Version: "v1.0.0",
		}},
		Replace: map[string]string{origin: originPath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	expected := "Installed: 1, Pruned: 0, Warnings: 0\n"
	if stdout.String() != expected {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRemoveOutputIncludesSummary(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillRoot := filepath.Join(t.TempDir(), "foo")
	otherRoot := filepath.Join(t.TempDir(), "bar")
	touchSkill(t, skillRoot)
	touchSkill(t, otherRoot)

	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{
			{Name: "foo", Origin: skillRoot},
			{Name: "bar", Origin: otherRoot},
		},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	skillsDir := filepath.Join(repo, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	if err := os.Symlink(skillRoot, filepath.Join(skillsDir, "foo")); err != nil {
		t.Fatalf("symlink foo: %v", err)
	}
	if err := os.Symlink(otherRoot, filepath.Join(skillsDir, "bar")); err != nil {
		t.Fatalf("symlink bar: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"remove", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	expected := fmt.Sprintf("Installed: 0, Pruned: 1, Warnings: 0\nRemoved: foo (%s)\n", skillRoot)
	if stdout.String() != expected {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRemoveMultipleOutputs(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillRoot := filepath.Join(t.TempDir(), "foo")
	otherRoot := filepath.Join(t.TempDir(), "bar")
	touchSkill(t, skillRoot)
	touchSkill(t, otherRoot)

	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{
			{Name: "foo", Origin: skillRoot},
			{Name: "bar", Origin: otherRoot},
		},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	skillsDir := filepath.Join(repo, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	if err := os.Symlink(skillRoot, filepath.Join(skillsDir, "foo")); err != nil {
		t.Fatalf("symlink foo: %v", err)
	}
	if err := os.Symlink(otherRoot, filepath.Join(skillsDir, "bar")); err != nil {
		t.Fatalf("symlink bar: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"remove", "foo", "bar"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	expected := fmt.Sprintf("No skills found.\nRemoved: foo (%s)\nRemoved: bar (%s)\n", skillRoot, otherRoot)
	if stdout.String() != expected {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRemoveNameAndOriginOutputsWithoutDuplicates(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	originPath := filepath.Join(repo, "origin")
	fooPath := filepath.Join(originPath, "skills", "foo")
	barPath := filepath.Join(originPath, "skills", "bar")
	touchSkill(t, fooPath)
	touchSkill(t, barPath)

	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{
			{Name: "foo", Origin: originPath, Subdir: "skills/foo"},
			{Name: "bar", Origin: originPath, Subdir: "skills/bar"},
		},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"remove", "foo", "--origin", originPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	expectedOrigin := originPath
	if resolved, err := filepath.EvalSymlinks(originPath); err == nil {
		expectedOrigin = resolved
	}
	expected := fmt.Sprintf("No skills found.\nRemoved: foo (%s)\nRemoved: bar (%s)\n", expectedOrigin, expectedOrigin)
	if stdout.String() != expected {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}
}

func TestRemoveOriginPrunesRemoteState(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	origin := "https://example.com/repo"
	config := manifest.Config{
		Skills: []manifest.Skill{
			{Name: "foo", Origin: origin, Subdir: "skills/foo", Version: "v1.0.0"},
			{Name: "bar", Origin: origin, Subdir: "skills/bar", Version: "v1.0.0"},
		},
		Replace: map[string]string{origin: filepath.Join(t.TempDir(), "replace")},
	}
	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), config); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	lockPath := filepath.Join(repo, "skills-lock.json")
	lock := map[manifest.LockKey]string{{Origin: origin, Version: "v1.0.0"}: "deadbeef"}
	if err := manifest.SaveLockWithSkills(lockPath, lock, config.Skills); err != nil {
		t.Fatalf("save lock: %v", err)
	}

	storePath := gitstore.RepoPath(filepath.Join(repo, ".asm", "store"), origin)
	if err := os.MkdirAll(storePath, 0o755); err != nil {
		t.Fatalf("mkdir store: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"remove", "--origin", origin + ".git@v1.0.0"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}

	expected := "No skills found.\n" +
		fmt.Sprintf("Removed: bar (%s)\n", origin) +
		fmt.Sprintf("Removed: foo (%s)\n", origin) +
		fmt.Sprintf("Pruned store: %s\n", origin)
	if stdout.String() != expected {
		t.Fatalf("unexpected stdout:\n%s", stdout.String())
	}

	loaded, err := manifest.Load(filepath.Join(repo, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(loaded.Skills) != 0 {
		t.Fatalf("expected no skills, got %d", len(loaded.Skills))
	}
	if len(loaded.Replace) != 0 {
		t.Fatalf("expected replace removed, got %v", loaded.Replace)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatalf("expected lock removed, got %v", err)
	}
	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		t.Fatalf("expected store pruned, got %v", err)
	}
}

func initTaggedRepo(t *testing.T, root string, tag string) {
	t.Helper()

	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	readme := filepath.Join(root, "README.md")
	if err := os.WriteFile(readme, []byte("init"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	commitPaths(t, repo, "init", time.Now(), "README.md")
	head, err := repo.Head()
	if err != nil {
		t.Fatalf("head: %v", err)
	}
	if _, err := repo.CreateTag(tag, head.Hash(), nil); err != nil {
		t.Fatalf("tag: %v", err)
	}
}
