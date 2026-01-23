package cli

import (
	"path/filepath"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestRemoveDefaultsToGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	globalConfigDir := filepath.Join(home, ".config", "asm")
	saveConfig(t, globalConfigDir, config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: "/global"}},
	})

	localConfigDir := filepath.Join(repo, ".asm")
	saveConfig(t, localConfigDir, config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: "/local"}},
	})

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"remove", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("remove: %v", err)
	}

	globalLoaded, _, err := config.Load(globalConfigDir)
	if err != nil {
		t.Fatalf("load global: %v", err)
	}
	if len(globalLoaded.Sources) != 0 {
		t.Fatalf("expected global source removed")
	}

	localLoaded, _, err := config.Load(localConfigDir)
	if err != nil {
		t.Fatalf("load local: %v", err)
	}
	if len(localLoaded.Sources) != 1 {
		t.Fatalf("expected local source to remain")
	}
}

func TestUpdateLocalNoop(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	localConfigDir := filepath.Join(repo, ".asm")
	saveConfig(t, localConfigDir, config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: repo}},
	})

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"update", "--local"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("update: %v", err)
	}
}
