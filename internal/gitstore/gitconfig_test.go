package gitstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyInsteadOfWithInclude(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	writeFile(t, home, ".gitconfig", "[include]\n\tpath = includes/gitconfig\n")
	writeFile(t, home, "includes/gitconfig", "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/\n[url \"git@github.com:org/\"]\n\tinsteadOf = https://github.com/org/\n")

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	t.Setenv("GIT_CONFIG_GLOBAL", "")

	got, changed, err := applyInsteadOf("https://github.com/org/repo")
	if err != nil {
		t.Fatalf("applyInsteadOf: %v", err)
	}
	if !changed {
		t.Fatalf("expected rewrite to apply")
	}
	if got != "git@github.com:org/repo" {
		t.Fatalf("expected rewritten url, got %q", got)
	}

	got, changed, err = applyInsteadOf("https://github.com/other/repo")
	if err != nil {
		t.Fatalf("applyInsteadOf: %v", err)
	}
	if !changed {
		t.Fatalf("expected rewrite to apply")
	}
	if got != "git@github.com:other/repo" {
		t.Fatalf("expected rewritten url, got %q", got)
	}
}

func TestApplyInsteadOfHonorsGitConfigGlobal(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	writeFile(t, home, ".gitconfig", "[url \"git@github.com:\"]\n\tinsteadOf = https://github.com/\n")
	custom := filepath.Join(root, "custom")
	writeFile(t, root, "custom", "[url \"ssh://git@github.com/\"]\n\tinsteadOf = https://github.com/\n")

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
	t.Setenv("GIT_CONFIG_GLOBAL", custom)

	got, changed, err := applyInsteadOf("https://github.com/org/repo")
	if err != nil {
		t.Fatalf("applyInsteadOf: %v", err)
	}
	if !changed {
		t.Fatalf("expected rewrite to apply")
	}
	if got != "ssh://git@github.com/org/repo" {
		t.Fatalf("expected rewritten url, got %q", got)
	}
}
