package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestIndexWritesMarkdown(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "foo")
	writeSkillDoc(t, skillDir, "# Foo Skill\n\nFoo does bar.\n")
	skillDir2 := filepath.Join(t.TempDir(), "bar")
	writeSkillDoc(t, skillDir2, "---\ntitle: Bar Skill\ndescription: Handles bars.\n---\n\n# Ignored\n")

	saveConfig(t, repo, manifest.Config{
		Skills: []manifest.Skill{
			{Name: "foo", Origin: skillDir},
			{Name: "acme/bar", Origin: skillDir2},
		},
	})

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"index"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("index: %v", err)
	}

	outputPath := filepath.Join(repo, "skills-index.md")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "| foo | Foo Skill | Foo does bar. | skills/foo/ |") {
		t.Fatalf("expected foo row, got:\n%s", content)
	}
	if !strings.Contains(content, "| acme/bar | Bar Skill | Handles bars. | skills/acme/bar/ |") {
		t.Fatalf("expected bar row, got:\n%s", content)
	}
	if !strings.Contains(stdout.String(), "Wrote ") {
		t.Fatalf("expected output to mention Wrote")
	}
}

func TestIndexWarnsOnMissingSkillDoc(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	skillDir := filepath.Join(t.TempDir(), "missing")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	saveConfig(t, repo, manifest.Config{
		Skills: []manifest.Skill{
			{Name: "missing", Origin: skillDir},
		},
	})

	cmd, stdout, stderr := newTestCommand()
	cmd.SetArgs([]string{"index"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("index: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("expected warning, got: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Wrote ") {
		t.Fatalf("expected output to mention Wrote")
	}

	outputPath := filepath.Join(repo, "skills-index.md")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "| missing | missing |  | skills/missing/ |") {
		t.Fatalf("expected missing row, got:\n%s", content)
	}
}

func writeSkillDoc(t *testing.T, dir string, content string) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}
