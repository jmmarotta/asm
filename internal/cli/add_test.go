package cli

import (
	"path/filepath"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/source"
)

func TestAddLocalMultiSkillCollision(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	touchSkill(t, filepath.Join(repo, "plugins", "foo"))

	localConfigDir := filepath.Join(repo, ".asm")
	saveConfig(t, localConfigDir, config.Config{
		Sources: []config.Source{{
			Name:   "foo",
			Type:   "path",
			Origin: "/tmp/other",
		}},
	})

	setWorkingDir(t, repo)
	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"add", "--local", repo})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("add: %v", err)
	}

	loaded, _, err := config.Load(localConfigDir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	names := make(map[string]struct{})
	for _, source := range loaded.Sources {
		names[source.Name] = struct{}{}
	}

	if _, ok := names["foo"]; !ok {
		t.Fatalf("expected existing foo")
	}
	expected := source.AuthorForLocalPath(repo) + "/foo"
	if _, ok := names[expected]; !ok {
		all := make([]string, 0, len(names))
		for name := range names {
			all = append(all, name)
		}
		t.Fatalf("expected %s collision name, got %v", expected, all)
	}
}
