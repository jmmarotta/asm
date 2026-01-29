package manifest

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestGitOriginVersions(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "one", Origin: "https://example.com/repo-a", Version: "v1.0.0"},
			{Name: "two", Origin: "/tmp/local"},
			{Name: "three", Origin: "https://example.com/repo-b", Version: "v2.0.0"},
		},
	}

	got := config.GitOriginVersions()
	if len(got) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(got))
	}
	if got["https://example.com/repo-a"] != "v1.0.0" {
		t.Fatalf("expected repo-a version v1.0.0, got %q", got["https://example.com/repo-a"])
	}
	if got["https://example.com/repo-b"] != "v2.0.0" {
		t.Fatalf("expected repo-b version v2.0.0, got %q", got["https://example.com/repo-b"])
	}
	if _, ok := got["/tmp/local"]; ok {
		t.Fatalf("expected path origin to be excluded")
	}
}

func TestResolveSkillPaths(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "local", Origin: "/tmp/local"},
			{Name: "remote", Origin: "https://example.com/repo", Subdir: "plugins/foo", Version: "v1.0.0"},
		},
	}

	originPaths := map[string]string{
		"https://example.com/repo": "/store/repo",
	}

	paths, err := config.ResolveSkillPaths(originPaths)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	if paths[0].Name != "local" || paths[0].Path != "/tmp/local" {
		t.Fatalf("expected local path /tmp/local, got %s (%s)", paths[0].Name, paths[0].Path)
	}

	expectedRemote := filepath.Join("/store/repo", "plugins", "foo")
	if paths[1].Name != "remote" || paths[1].Path != expectedRemote {
		t.Fatalf("expected remote path %s, got %s (%s)", expectedRemote, paths[1].Name, paths[1].Path)
	}
}

func TestResolveSkillPathsMissingOrigin(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "remote", Origin: "https://example.com/repo", Subdir: "plugins/foo", Version: "v1.0.0"},
		},
	}

	_, err := config.ResolveSkillPaths(map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
	var missing MissingOriginPathError
	if !errors.As(err, &missing) {
		t.Fatalf("expected MissingOriginPathError, got %T", err)
	}
	if missing.Origin != "https://example.com/repo" {
		t.Fatalf("expected https://example.com/repo, got %q", missing.Origin)
	}
	if missing.Skill != "remote" {
		t.Fatalf("expected skill remote, got %q", missing.Skill)
	}
}
