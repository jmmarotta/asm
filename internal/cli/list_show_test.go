package cli

import (
	"encoding/json"
	"os"
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

func TestLsOutputsSkillsAlphabetically(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	alpha := filepath.Join(t.TempDir(), "alpha")
	mid := filepath.Join(t.TempDir(), "mid")
	zeta := filepath.Join(t.TempDir(), "zeta")

	payload := struct {
		Skills []manifest.Skill `json:"skills"`
	}{
		Skills: []manifest.Skill{
			{Name: "zeta", Origin: zeta},
			{Name: "alpha", Origin: alpha},
			{Name: "mid", Origin: mid},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "skills.jsonc"), data, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"ls"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ls: %v", err)
	}

	output := stdout.String()
	alphaIndex := strings.Index(output, "alpha")
	midIndex := strings.Index(output, "mid")
	zetaIndex := strings.Index(output, "zeta")
	if alphaIndex == -1 || midIndex == -1 || zetaIndex == -1 {
		t.Fatalf("expected all skill names in output, got %q", output)
	}
	if !(alphaIndex < midIndex && midIndex < zetaIndex) {
		t.Fatalf("expected alphabetical order alpha -> mid -> zeta, got %q", output)
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
