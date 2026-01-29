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
			{Name: "local", Origin: "/tmp/local"},
			{Name: "remote", Origin: "https://example.com/repo", Subdir: "plugins/foo", Version: "v1.0.0"},
		},
	}

	originPaths := map[string]string{"https://example.com/repo": "/store/repo"}

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
		Skills: []manifest.Skill{{Name: "remote", Origin: "https://example.com/repo", Version: "v1.0.0"}},
	}

	_, err := SourcesFromConfig(config, map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
