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

func newTestCommand() (*cobra.Command, *bytes.Buffer) {
	viper.Reset()
	cmd := newRootCommand()
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	return cmd, buffer
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
	path := filepath.Join(dir, "config.jsonc")
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
}
