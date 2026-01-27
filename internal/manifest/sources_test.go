package manifest

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestGitOriginVersions(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "one", Type: "git", Origin: "origin-a", Version: "v1.0.0"},
			{Name: "two", Type: "path", Origin: "/tmp/local"},
			{Name: "three", Type: "git", Origin: "origin-b", Version: "v2.0.0"},
		},
	}

	got := config.GitOriginVersions()
	if len(got) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(got))
	}
	if got["origin-a"] != "v1.0.0" {
		t.Fatalf("expected origin-a version v1.0.0, got %q", got["origin-a"])
	}
	if got["origin-b"] != "v2.0.0" {
		t.Fatalf("expected origin-b version v2.0.0, got %q", got["origin-b"])
	}
	if _, ok := got["/tmp/local"]; ok {
		t.Fatalf("expected path origin to be excluded")
	}
}

func TestResolveSkillPaths(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "local", Type: "path", Origin: "/tmp/local"},
			{Name: "remote", Type: "git", Origin: "origin-a", Subdir: "plugins/foo"},
		},
	}

	originPaths := map[string]string{
		"origin-a": "/store/repo",
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
			{Name: "remote", Type: "git", Origin: "origin-a", Subdir: "plugins/foo"},
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
	if missing.Origin != "origin-a" {
		t.Fatalf("expected origin-a, got %q", missing.Origin)
	}
	if missing.Skill != "remote" {
		t.Fatalf("expected skill remote, got %q", missing.Skill)
	}
}
