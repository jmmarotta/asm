package config

import "testing"

func TestConfigValidateMissingFields(t *testing.T) {
	config := Config{
		Skills: []Skill{{Name: "", Type: "git", Origin: "x", Version: "v1.0.0"}},
	}
	if err := config.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestConfigUpsertAndRemove(t *testing.T) {
	config := Config{}
	config.UpsertSkill(Skill{Name: "one", Type: "git", Origin: "a", Version: "v1.0.0"})
	config.UpsertSkill(Skill{Name: "one", Type: "git", Origin: "b", Version: "v1.0.0"})
	if len(config.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(config.Skills))
	}
	if config.Skills[0].Origin != "b" {
		t.Fatalf("expected origin b, got %q", config.Skills[0].Origin)
	}

	removed, ok := config.RemoveSkill("one")
	if !ok {
		t.Fatalf("expected remove to succeed")
	}
	if removed.Name != "one" {
		t.Fatalf("expected removed name one, got %q", removed.Name)
	}
}

func TestConfigRejectsMultipleVersionsPerOrigin(t *testing.T) {
	config := Config{
		Skills: []Skill{
			{Name: "one", Type: "git", Origin: "example", Version: "v1.0.0"},
			{Name: "two", Type: "git", Origin: "example", Version: "v1.1.0"},
		},
	}
	if err := config.Validate(); err == nil {
		t.Fatalf("expected version conflict error")
	}
}
