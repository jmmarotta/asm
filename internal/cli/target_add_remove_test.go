package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestTargetAddGlobalStoresAbsolutePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	targetPath := filepath.Join(home, "targets", "main")

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"target", "add", "main", "~/targets/main"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("target add: %v", err)
	}

	loaded, _, err := config.Load(filepath.Join(home, ".config", "asm"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(loaded.Targets))
	}
	if loaded.Targets[0].Path != targetPath {
		t.Fatalf("expected path %q, got %q", targetPath, loaded.Targets[0].Path)
	}
}

func TestTargetAddLocalStoresAbsolutePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"target", "add", "--local", "local", "./skills"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("target add local: %v", err)
	}

	expected, err := filepath.Abs("skills")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	loaded, _, err := config.Load(filepath.Join(repo, ".asm"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(loaded.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(loaded.Targets))
	}
	if loaded.Targets[0].Path != expected {
		t.Fatalf("expected path %q, got %q", expected, loaded.Targets[0].Path)
	}
}

func TestTargetRemoveCleansSymlink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	source := filepath.Join(home, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	targetPath := filepath.Join(home, "targets")
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.Symlink(source, filepath.Join(targetPath, "foo")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	saveConfig(t, filepath.Join(home, ".config", "asm"), config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: source}},
		Targets: []config.Target{{Name: "skills", Path: targetPath}},
	})

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"target", "remove", "skills"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("target remove: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(targetPath, "foo")); !os.IsNotExist(err) {
		t.Fatalf("expected symlink removed")
	}
}

func TestTargetRemoveSkipsCleanupWhenTargetStillEffective(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	globalTarget := filepath.Join(home, "global-target")
	if err := os.MkdirAll(globalTarget, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.Symlink(filepath.Join(home, "source"), filepath.Join(globalTarget, "foo")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	saveConfig(t, filepath.Join(home, ".config", "asm"), config.Config{
		Targets: []config.Target{{Name: "skills", Path: globalTarget}},
	})
	saveConfig(t, filepath.Join(repo, ".asm"), config.Config{
		Targets: []config.Target{{Name: "skills", Path: filepath.Join(repo, "local-target")}},
	})

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"target", "remove", "skills"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("target remove: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(globalTarget, "foo")); err != nil {
		t.Fatalf("expected symlink to remain")
	}
}

func TestTargetRemoveWarnsOnRealDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	targetPath := filepath.Join(home, "targets")
	if err := os.MkdirAll(filepath.Join(targetPath, "foo"), 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}

	saveConfig(t, filepath.Join(home, ".config", "asm"), config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: filepath.Join(home, "source")}},
		Targets: []config.Target{{Name: "skills", Path: targetPath}},
	})

	cmd, buffer := newTestCommand()
	cmd.SetArgs([]string{"target", "remove", "skills"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("target remove: %v", err)
	}
	if !strings.Contains(buffer.String(), "warning") {
		t.Fatalf("expected warning output")
	}
}
