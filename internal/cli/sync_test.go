package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestSyncScopeFiltersSources(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	globalSource := filepath.Join(home, "global-source")
	if err := os.MkdirAll(globalSource, 0o755); err != nil {
		t.Fatalf("mkdir global source: %v", err)
	}
	localSource := filepath.Join(repo, "local-source")
	if err := os.MkdirAll(localSource, 0o755); err != nil {
		t.Fatalf("mkdir local source: %v", err)
	}

	targetPath := filepath.Join(home, "targets")
	saveConfig(t, filepath.Join(home, ".config", "asm"), config.Config{
		Sources: []config.Source{{Name: "global", Type: "path", Origin: globalSource}},
		Targets: []config.Target{{Name: "main", Path: targetPath}},
	})
	saveConfig(t, filepath.Join(repo, ".asm"), config.Config{
		Sources: []config.Source{{Name: "local", Type: "path", Origin: localSource}},
	})

	cmd, _ := newTestCommand()
	cmd.SetArgs([]string{"sync", "--scope", "global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	info, err := os.Lstat(filepath.Join(targetPath, "global"))
	if err != nil {
		t.Fatalf("expected global symlink")
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink for global source")
	}
	if _, err := os.Lstat(filepath.Join(targetPath, "local")); !os.IsNotExist(err) {
		t.Fatalf("expected local not to be synced")
	}
}
