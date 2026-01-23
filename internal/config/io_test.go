package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrefersJSONC(t *testing.T) {
	root := t.TempDir()
	jsoncPath := filepath.Join(root, "config.jsonc")
	jsonPath := filepath.Join(root, "config.json")

	if err := os.WriteFile(jsonPath, []byte(`{"sources": [{"name": "json", "type": "git", "origin": "x"}], "targets": []}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(jsoncPath, []byte(`// comment
{
  "sources": [{"name": "jsonc", "type": "git", "origin": "y"}],
  "targets": []
}`), 0o644); err != nil {
		t.Fatalf("write jsonc: %v", err)
	}

	loaded, _, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Sources) != 1 || loaded.Sources[0].Name != "jsonc" {
		t.Fatalf("expected jsonc config to load")
	}
}

func TestSaveAndLoadHomeExpansion(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "projects"), 0o755); err != nil {
		t.Fatalf("mkdir projects: %v", err)
	}
	t.Setenv("HOME", home)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.jsonc")
	input := Config{
		Sources: []Source{{
			Name:   "local",
			Type:   "path",
			Origin: filepath.Join(home, "projects"),
		}},
		Targets: []Target{{
			Name: "dotfiles",
			Path: filepath.Join(home, "dots"),
		}},
	}

	if err := Save(configPath, input); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), "$HOME/projects") {
		t.Fatalf("expected $HOME substitution")
	}
	if !strings.Contains(string(data), "$HOME/dots") {
		t.Fatalf("expected $HOME substitution for target path")
	}

	loaded, _, err := Load(configDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Sources[0].Origin != filepath.Join(home, "projects") {
		t.Fatalf("expected expanded origin, got %q", loaded.Sources[0].Origin)
	}
	if loaded.Targets[0].Path != filepath.Join(home, "dots") {
		t.Fatalf("expected expanded target path, got %q", loaded.Targets[0].Path)
	}
}
