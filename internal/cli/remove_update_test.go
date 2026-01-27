package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestRemovePrunesSymlink(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillRoot := filepath.Join(t.TempDir(), "foo")
	touchSkill(t, skillRoot)

	if err := config.Save(filepath.Join(repo, "skills.jsonc"), config.Config{
		Skills: []config.Skill{{
			Name:   "foo",
			Type:   "path",
			Origin: skillRoot,
		}},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	skillsDir := filepath.Join(repo, "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills: %v", err)
	}
	linkPath := filepath.Join(skillsDir, "foo")
	if err := os.Symlink(skillRoot, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"remove", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}

	loaded, err := config.Load(filepath.Join(repo, "skills.jsonc"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(loaded.Skills) != 0 {
		t.Fatalf("expected no skills, got %d", len(loaded.Skills))
	}
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Fatalf("expected symlink removed")
	}
}

func TestUpdateNoSkillsNoop(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	if err := config.Save(filepath.Join(repo, "skills.jsonc"), config.Config{}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"update"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}
}
