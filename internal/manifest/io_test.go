package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPrefersJSONC(t *testing.T) {
	root := t.TempDir()
	jsoncPath := filepath.Join(root, "skills.jsonc")
	jsonPath := filepath.Join(root, "skills.json")

	if err := os.WriteFile(jsonPath, []byte(`{"skills": [{"name": "json", "origin": "https://example.com/repo", "version": "v1.0.0"}]}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := os.WriteFile(jsoncPath, []byte(`// comment
{
	"skills": [{"name": "jsonc", "origin": "https://example.com/repo", "version": "v1.0.0"}]
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

func TestSaveDoesNotEmitTypeField(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "skills.jsonc")
	input := Config{
		Skills: []Skill{{
			Name:    "remote",
			Origin:  "https://example.com/repo",
			Version: "v1.0.0",
		}},
	}

	if err := Save(configPath, input); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(data), "\"type\"") {
		t.Fatalf("expected no type field, got %s", string(data))
	}
}

func TestLoadNormalizesFileURIOriginsAndReplace(t *testing.T) {
	root := t.TempDir()
	local := filepath.Join(root, "local")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatalf("mkdir local: %v", err)
	}
	configPath := filepath.Join(root, "skills.jsonc")
	fileURI := "file://" + filepath.ToSlash(local)
	payload := `{"skills": [{"name": "local", "origin": "` + fileURI + `"}], "replace": {"https://example.com/repo": "` + fileURI + `"}}`
	if err := os.WriteFile(configPath, []byte(payload), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Skills[0].Origin != local {
		t.Fatalf("expected file uri origin %q, got %q", local, loaded.Skills[0].Origin)
	}
	if loaded.Replace["https://example.com/repo"] != local {
		t.Fatalf("expected file uri replace %q, got %q", local, loaded.Replace["https://example.com/repo"])
	}
}

func TestLoadIgnoresLegacyTypeField(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "skills.jsonc")
	if err := os.WriteFile(configPath, []byte(`{"skills": [{"name": "legacy", "type": "git", "origin": "https://example.com/repo", "version": "v1.0.0"}]}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(loaded.Skills))
	}
	if loaded.Skills[0].Origin != "https://example.com/repo" {
		t.Fatalf("expected origin https://example.com/repo, got %q", loaded.Skills[0].Origin)
	}
}

func TestSaveSortsSkillsAlphabetically(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "skills.jsonc")

	input := Config{
		Skills: []Skill{
			{Name: "zeta", Origin: "/tmp/zeta"},
			{Name: "alpha", Origin: "/tmp/alpha"},
			{Name: "mid", Origin: "/tmp/mid"},
		},
	}

	if err := Save(configPath, input); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Skills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(loaded.Skills))
	}
	if loaded.Skills[0].Name != "alpha" {
		t.Fatalf("expected first skill alpha, got %q", loaded.Skills[0].Name)
	}
	if loaded.Skills[1].Name != "mid" {
		t.Fatalf("expected second skill mid, got %q", loaded.Skills[1].Name)
	}
	if loaded.Skills[2].Name != "zeta" {
		t.Fatalf("expected third skill zeta, got %q", loaded.Skills[2].Name)
	}
}

func TestSaveLockWithSkillsOrdersByOriginVersionSubdirAndName(t *testing.T) {
	root := t.TempDir()
	lockPath := filepath.Join(root, "skills-lock.json")

	entries := map[LockKey]string{
		{Origin: "https://example.com/repo-a", Version: "v1.0.0"}: "aaaa1111",
		{Origin: "https://example.com/repo-a", Version: "v2.0.0"}: "bbbb2222",
		{Origin: "https://example.com/repo-b", Version: "v1.0.0"}: "cccc3333",
	}
	skills := []Skill{
		{Name: "zeta", Origin: "https://example.com/repo-a", Version: "v1.0.0", Subdir: "plugins/z"},
		{Name: "alpha", Origin: "https://example.com/repo-a", Version: "v1.0.0", Subdir: "plugins/a"},
		{Name: "root", Origin: "https://example.com/repo-a", Version: "v1.0.0"},
		{Name: "beta", Origin: "https://example.com/repo-a", Version: "v2.0.0", Subdir: "plugins/b"},
	}

	if err := SaveLockWithSkills(lockPath, entries, skills); err != nil {
		t.Fatalf("SaveLockWithSkills: %v", err)
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}

	var parsed lockFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal lock: %v", err)
	}

	if len(parsed.Entries) != 5 {
		t.Fatalf("expected 5 lock entries, got %d", len(parsed.Entries))
	}

	got := make([]string, 0, len(parsed.Entries))
	for _, entry := range parsed.Entries {
		got = append(got, entry.Origin+"|"+entry.Version+"|"+entry.Subdir+"|"+entry.Name)
	}
	want := []string{
		"https://example.com/repo-a|v1.0.0||root",
		"https://example.com/repo-a|v1.0.0|plugins/a|alpha",
		"https://example.com/repo-a|v1.0.0|plugins/z|zeta",
		"https://example.com/repo-a|v2.0.0|plugins/b|beta",
		"https://example.com/repo-b|v1.0.0||",
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("entry %d: expected %q, got %q", index, want[index], got[index])
		}
	}

	loaded, err := LoadLock(lockPath)
	if err != nil {
		t.Fatalf("LoadLock: %v", err)
	}
	if len(loaded) != 3 {
		t.Fatalf("expected 3 resolved lock keys, got %d", len(loaded))
	}
	if loaded[LockKey{Origin: "https://example.com/repo-a", Version: "v1.0.0"}] != "aaaa1111" {
		t.Fatalf("unexpected rev for repo-a v1.0.0")
	}
	if loaded[LockKey{Origin: "https://example.com/repo-a", Version: "v2.0.0"}] != "bbbb2222" {
		t.Fatalf("unexpected rev for repo-a v2.0.0")
	}
	if loaded[LockKey{Origin: "https://example.com/repo-b", Version: "v1.0.0"}] != "cccc3333" {
		t.Fatalf("unexpected rev for repo-b v1.0.0")
	}
}
