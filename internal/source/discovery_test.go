package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkillsSingleRoot(t *testing.T) {
	root := t.TempDir()
	if err := touch(filepath.Join(root, "SKILL.md")); err != nil {
		t.Fatalf("touch SKILL.md: %v", err)
	}

	skills, err := DiscoverSkills(root, "")
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Subdir != "" {
		t.Fatalf("expected empty subdir, got %q", skills[0].Subdir)
	}
}

func TestDiscoverSkillsMultiRoot(t *testing.T) {
	root := t.TempDir()
	if err := touch(filepath.Join(root, "plugins", "one", "SKILL.md")); err != nil {
		t.Fatalf("touch one: %v", err)
	}
	if err := touch(filepath.Join(root, "plugins", "two", "SKILL.md")); err != nil {
		t.Fatalf("touch two: %v", err)
	}

	skills, err := DiscoverSkills(root, "")
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestDiscoverSkillsAmbiguousRoots(t *testing.T) {
	root := t.TempDir()
	if err := touch(filepath.Join(root, "plugins", "one", "SKILL.md")); err != nil {
		t.Fatalf("touch plugins: %v", err)
	}
	if err := touch(filepath.Join(root, "skills", "two", "SKILL.md")); err != nil {
		t.Fatalf("touch skills: %v", err)
	}

	if _, err := DiscoverSkills(root, ""); err == nil {
		t.Fatalf("expected error for ambiguous roots")
	}
}

func TestDiscoverSkillsPathSpecific(t *testing.T) {
	root := t.TempDir()
	if err := touch(filepath.Join(root, "plugins", "one", "SKILL.md")); err != nil {
		t.Fatalf("touch one: %v", err)
	}
	if err := touch(filepath.Join(root, "plugins", "two", "SKILL.md")); err != nil {
		t.Fatalf("touch two: %v", err)
	}

	skills, err := DiscoverSkills(root, "plugins/one")
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "one" {
		t.Fatalf("expected name one, got %q", skills[0].Name)
	}
}

func TestDiscoverSkillsPathMultiRoot(t *testing.T) {
	root := t.TempDir()
	if err := touch(filepath.Join(root, "plugins", "one", "SKILL.md")); err != nil {
		t.Fatalf("touch one: %v", err)
	}
	if err := touch(filepath.Join(root, "plugins", "two", "SKILL.md")); err != nil {
		t.Fatalf("touch two: %v", err)
	}

	skills, err := DiscoverSkills(root, "plugins")
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func touch(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte("# skill"), 0o644)
}
