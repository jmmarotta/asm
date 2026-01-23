package config

import "testing"

func TestConfigValidateMissingFields(t *testing.T) {
	config := Config{
		Sources: []Source{{Name: "", Type: "git", Origin: "x"}},
	}
	if err := config.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestConfigUpsertAndRemove(t *testing.T) {
	config := Config{}
	config.UpsertSource(Source{Name: "one", Type: "git", Origin: "a"})
	config.UpsertSource(Source{Name: "one", Type: "git", Origin: "b"})
	if len(config.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(config.Sources))
	}
	if config.Sources[0].Origin != "b" {
		t.Fatalf("expected origin b, got %q", config.Sources[0].Origin)
	}

	removed, ok := config.RemoveSource("one")
	if !ok {
		t.Fatalf("expected remove to succeed")
	}
	if removed.Name != "one" {
		t.Fatalf("expected removed name one, got %q", removed.Name)
	}
}

func TestTargetUpsertAndRemove(t *testing.T) {
	config := Config{}
	config.UpsertTarget(Target{Name: "main", Path: "/tmp"})
	config.UpsertTarget(Target{Name: "main", Path: "/home"})
	if len(config.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(config.Targets))
	}
	if config.Targets[0].Path != "/home" {
		t.Fatalf("expected path /home, got %q", config.Targets[0].Path)
	}

	removed, ok := config.RemoveTarget("main")
	if !ok {
		t.Fatalf("expected remove to succeed")
	}
	if removed.Name != "main" {
		t.Fatalf("expected removed name main, got %q", removed.Name)
	}
}
