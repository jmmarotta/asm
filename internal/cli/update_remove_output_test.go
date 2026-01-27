package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestUpdateOutputIncludesOrigins(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	origin := t.TempDir()
	initTaggedRepo(t, origin, "v1.0.0")

	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Type:    "git",
			Origin:  origin,
			Version: "v1.0.0",
		}},
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

	expected := fmt.Sprintf("Installed: 1, Pruned: 0, Warnings: 0\nUpdated origins: %s\n", origin)
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
			{Name: "foo", Type: "path", Origin: skillRoot},
			{Name: "bar", Type: "path", Origin: otherRoot},
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
