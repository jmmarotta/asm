package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestLsAndShow(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "foo")
	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:   "foo",
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
		manifest.Skill
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
	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:   "foo",
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
		manifest.Skill
		Replace string `json:"replace"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if output.Name != "foo" {
		t.Fatalf("expected foo, got %s", output.Name)
	}
}

func TestShowIncludesReplaceAndVersion(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	replacePath := filepath.Join(t.TempDir(), "repo")
	origin := "https://example.com/repo"
	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:    "foo",
			Origin:  origin,
			Version: "v1.0.0",
		}},
		Replace: map[string]string{origin: replacePath},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"show", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if output["replace"] != replacePath {
		t.Fatalf("expected replace %q, got %v", replacePath, output["replace"])
	}
	if output["version"] != "v1.0.0" {
		t.Fatalf("expected version v1.0.0, got %v", output["version"])
	}
}

func TestShowOmitsEmptyFields(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "foo")
	if err := manifest.Save(filepath.Join(repo, "skills.jsonc"), manifest.Config{
		Skills: []manifest.Skill{{
			Name:   "foo",
			Origin: skillDir,
		}},
	}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"show", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := output["version"]; ok {
		t.Fatalf("expected version omitted")
	}
	if _, ok := output["replace"]; ok {
		t.Fatalf("expected replace omitted")
	}
	if _, ok := output["subdir"]; ok {
		t.Fatalf("expected subdir omitted")
	}
	if _, ok := output["type"]; ok {
		t.Fatalf("expected type omitted")
	}
}
