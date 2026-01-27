package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
)

func TestLsAndShow(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "foo")
	if err := config.Save(filepath.Join(repo, "skills.jsonc"), config.Config{
		Skills: []config.Skill{{
			Name:   "foo",
			Type:   "path",
			Origin: skillDir,
		}},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"ls"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ls: %v", err)
	}
	if !strings.Contains(stdout.String(), "foo") {
		t.Fatalf("expected ls output to contain skill name")
	}

	cmd, stdout, _ = newTestCommand()
	cmd.SetArgs([]string{"show", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	var output struct {
		config.Skill
		Replace string `json:"replace"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if output.Name != "foo" {
		t.Fatalf("expected foo, got %s", output.Name)
	}
}

func TestDebugFlagWritesToStderr(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "foo")
	if err := config.Save(filepath.Join(repo, "skills.jsonc"), config.Config{
		Skills: []config.Skill{{
			Name:   "foo",
			Type:   "path",
			Origin: skillDir,
		}},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"--debug", "show", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show debug: %v", err)
	}
	if !strings.Contains(stderr.String(), "debug:") {
		t.Fatalf("expected debug logs on stderr")
	}

	var output struct {
		config.Skill
		Replace string `json:"replace"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if output.Name != "foo" {
		t.Fatalf("expected foo, got %s", output.Name)
	}
}
