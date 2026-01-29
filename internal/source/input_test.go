package source

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseInputLocalInfersRepoRoot(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	skillDir := filepath.Join(repo, "plugins", "foo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}

	input, err := ParseInput(skillDir, "")
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}

	if input.Origin != repo {
		t.Fatalf("expected origin %q, got %q", repo, input.Origin)
	}
	if input.Subdir != "plugins/foo" {
		t.Fatalf("expected subdir plugins/foo, got %q", input.Subdir)
	}
	if input.Ref != "" {
		t.Fatalf("expected empty ref, got %q", input.Ref)
	}
	if input.RepoRoot != repo {
		t.Fatalf("expected repo root %q, got %q", repo, input.RepoRoot)
	}
}

func TestParseInputLocalPathFlagJoins(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	input, err := ParseInput(repo, "skills/foo")
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if input.Subdir != "skills/foo" {
		t.Fatalf("expected subdir skills/foo, got %q", input.Subdir)
	}
}

func TestParseInputRemote(t *testing.T) {
	input, err := ParseInput("https://github.com/org/repo.git@main", "plugins/foo")
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if input.Origin != "https://github.com/org/repo" {
		t.Fatalf("expected normalized origin, got %q", input.Origin)
	}
	if input.Ref != "main" {
		t.Fatalf("expected ref main, got %q", input.Ref)
	}
	if input.Subdir != "plugins/foo" {
		t.Fatalf("expected subdir plugins/foo, got %q", input.Subdir)
	}
}

func TestParseInputRejectsAbsoluteSubdir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	if _, err := ParseInput(root, "/abs"); err == nil {
		t.Fatalf("expected error for absolute subdir")
	}
}

func TestParseInputRejectsEscapingSubdir(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	if _, err := ParseInput(root, "../escape"); err == nil {
		t.Fatalf("expected error for escaping subdir")
	}
}

func TestParseInputFileURIConvertsToPath(t *testing.T) {
	root := t.TempDir()
	fileURI := "file://" + filepath.ToSlash(root)

	input, err := ParseInput(fileURI, "")
	if err != nil {
		t.Fatalf("ParseInput: %v", err)
	}
	if input.Origin != root {
		t.Fatalf("expected origin %q, got %q", root, input.Origin)
	}
	if !input.IsLocal {
		t.Fatalf("expected local input")
	}
}

func TestParseInputRejectsUnknownScheme(t *testing.T) {
	if _, err := ParseInput("s3://bucket/repo", ""); err == nil {
		t.Fatalf("expected error for unsupported scheme")
	}
}
