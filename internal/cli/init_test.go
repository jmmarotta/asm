package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func TestInitCreatesManifestAndDirs(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, "skills.jsonc")); err != nil {
		t.Fatalf("skills.jsonc: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".asm", "store")); err != nil {
		t.Fatalf("store dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, ".asm", "cache")); err != nil {
		t.Fatalf("cache dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "skills")); err != nil {
		t.Fatalf("skills dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "skills-lock.json")); !os.IsNotExist(err) {
		t.Fatalf("expected no skills-lock.json")
	}

	content := readGitignore(t, repo)
	if countGitignoreLine(content, ".asm/") != 1 {
		t.Fatalf("expected .asm/ in gitignore")
	}
	if countGitignoreLine(content, "skills/") != 1 {
		t.Fatalf("expected skills/ in gitignore")
	}
}

func TestInitIdempotentGitignore(t *testing.T) {
	repo := t.TempDir()
	setWorkingDir(t, repo)

	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("# existing\n"), 0o644); err != nil {
		t.Fatalf("write gitignore: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}
	cmd, _, _ = newTestCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init 2: %v", err)
	}

	content := readGitignore(t, repo)
	if countGitignoreLine(content, ".asm/") != 1 {
		t.Fatalf("expected single .asm/ entry")
	}
	if countGitignoreLine(content, "skills/") != 1 {
		t.Fatalf("expected single skills/ entry")
	}
}

func TestInitRejectsParentManifest(t *testing.T) {
	root := t.TempDir()
	if err := manifest.Save(filepath.Join(root, "skills.jsonc"), manifest.Config{Skills: []manifest.Skill{}}); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	child := filepath.Join(root, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("mkdir child: %v", err)
	}
	setWorkingDir(t, child)

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"init"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected init to fail")
	}
}

func TestInitWithCwdFlagCreatesInDir(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"init", "--cwd", repo})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, "skills.jsonc")); err != nil {
		t.Fatalf("skills.jsonc: %v", err)
	}
	content := readGitignore(t, repo)
	if countGitignoreLine(content, ".asm/") != 1 {
		t.Fatalf("expected .asm/ in gitignore")
	}
	if countGitignoreLine(content, "skills/") != 1 {
		t.Fatalf("expected skills/ in gitignore")
	}
}

func TestInitWithCwdFlagMissingDirErrors(t *testing.T) {
	root := t.TempDir()
	missing := filepath.Join(root, "missing")

	cmd, _, _ := newTestCommand()
	cmd.SetArgs([]string{"init", "--cwd", missing})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected init to fail")
	}
}

func readGitignore(t *testing.T, root string) string {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		t.Fatalf("read gitignore: %v", err)
	}
	return string(data)
}

func countGitignoreLine(content string, line string) int {
	count := 0
	for _, entry := range strings.Split(content, "\n") {
		if strings.TrimSpace(entry) == line {
			count++
		}
	}
	return count
}
