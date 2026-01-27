package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func newTestCommand() (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	viper.Reset()
	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd, stdout, stderr
}

func setWorkingDir(t *testing.T, dir string) {
	current, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(current)
	})
}

func touchSkill(t *testing.T, dir string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# skill"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}

func initRepo(t *testing.T, root string) {
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
}

func saveConfig(t *testing.T, dir string, cfg config.Config) {
	path := filepath.Join(dir, "skills.jsonc")
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
}

func assertSymlink(t *testing.T, dest string, target string) {
	t.Helper()

	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("lstat %s: %v", dest, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s", dest)
	}

	link, err := os.Readlink(dest)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}

	resolved := link
	if !filepath.IsAbs(link) {
		resolved = filepath.Join(filepath.Dir(dest), link)
	}

	left, err := filepath.Abs(resolved)
	if err != nil {
		t.Fatalf("abs link: %v", err)
	}
	right, err := filepath.Abs(target)
	if err != nil {
		t.Fatalf("abs target: %v", err)
	}
	if filepath.Clean(left) != filepath.Clean(right) {
		t.Fatalf("expected link %s to point to %s", left, right)
	}
}
