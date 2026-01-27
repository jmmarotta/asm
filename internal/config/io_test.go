package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrefersJSONC(t *testing.T) {
	root := t.TempDir()
	jsoncPath := filepath.Join(root, "skills.jsonc")
	jsonPath := filepath.Join(root, "skills.json")

	if err := os.WriteFile(jsonPath, []byte(`{"skills": [{"name": "json", "type": "git", "origin": "x", "version": "v1.0.0"}]}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(jsoncPath, []byte(`// comment
{
  "skills": [{"name": "jsonc", "type": "git", "origin": "y", "version": "v1.0.0"}]
}`), 0o644); err != nil {
		t.Fatalf("write jsonc: %v", err)
	}

	path, err := FindManifestPath(root)
	if err != nil {
		t.Fatalf("FindManifestPath: %v", err)
	}
	if path != jsoncPath {
		t.Fatalf("expected jsonc path, got %s", path)
	}
}

func TestSaveAndLoadRelativePaths(t *testing.T) {
	root := t.TempDir()
	local := filepath.Join(root, "local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatalf("mkdir local: %v", err)
	}
	configPath := filepath.Join(root, "skills.jsonc")
	input := Config{
		Skills: []Skill{{
			Name:   "local",
			Type:   "path",
			Origin: local,
		}},
		Replace: map[string]string{
			"https://example.com/repo": local,
		},
	}

	if err := Save(configPath, input); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "\"origin\": \"local\"") {
		t.Fatalf("expected relative origin, got %s", string(data))
	}
	if !strings.Contains(string(data), "\"https://example.com/repo\": \"local\"") {
		t.Fatalf("expected relative replace path")
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Skills[0].Origin != local {
		t.Fatalf("expected expanded origin, got %q", loaded.Skills[0].Origin)
	}
	if loaded.Replace["https://example.com/repo"] != local {
		t.Fatalf("expected expanded replace path, got %q", loaded.Replace["https://example.com/repo"])
	}
}
