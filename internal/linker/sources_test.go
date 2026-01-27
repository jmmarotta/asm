package linker

import (
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestSourcesFromSkillPaths(t *testing.T) {
	paths := []manifest.SkillPath{
		{Name: "alpha", Path: "/tmp/alpha"},
		{Name: "beta", Path: "/tmp/beta"},
	}

	sources := SourcesFromSkillPaths(paths)
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].Name != "alpha" || sources[0].Path != "/tmp/alpha" {
		t.Fatalf("expected alpha source, got %s (%s)", sources[0].Name, sources[0].Path)
	}
	if sources[1].Name != "beta" || sources[1].Path != "/tmp/beta" {
		t.Fatalf("expected beta source, got %s (%s)", sources[1].Name, sources[1].Path)
	}
}

func TestSourcesFromConfig(t *testing.T) {
	config := manifest.Config{
		Skills: []manifest.Skill{
			{Name: "local", Type: "path", Origin: "/tmp/local"},
			{Name: "remote", Type: "git", Origin: "origin-a", Subdir: "plugins/foo"},
		},
	}

	originPaths := map[string]string{"origin-a": "/store/repo"}

	sources, err := SourcesFromConfig(config, originPaths)
	if err != nil {
		t.Fatalf("sources: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0].Name != "local" || sources[0].Path != "/tmp/local" {
		t.Fatalf("expected local source, got %s (%s)", sources[0].Name, sources[0].Path)
	}
	if sources[1].Name != "remote" || sources[1].Path != "/store/repo/plugins/foo" {
		t.Fatalf("expected remote source, got %s (%s)", sources[1].Name, sources[1].Path)
	}
}

func TestSourcesFromConfigMissingOrigin(t *testing.T) {
	config := manifest.Config{
		Skills: []manifest.Skill{{Name: "remote", Type: "git", Origin: "origin-a"}},
	}

	_, err := SourcesFromConfig(config, map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
