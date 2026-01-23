package cli

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

func TestListAndShowScopePrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repo := filepath.Join(home, "repo")
	initRepo(t, repo)
	setWorkingDir(t, repo)

	globalConfigDir := filepath.Join(home, ".config", "asm")
	saveConfig(t, globalConfigDir, config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: "/global"}},
	})

	localConfigDir := filepath.Join(repo, ".asm")
	saveConfig(t, localConfigDir, config.Config{
		Sources: []config.Source{{Name: "foo", Type: "path", Origin: "/local"}},
	})

	cmd, buffer := newTestCommand()
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
	output := buffer.String()
	if !strings.Contains(output, "local") {
		t.Fatalf("expected local scope in list output")
	}

	cmd, buffer = newTestCommand()
	cmd.SetArgs([]string{"show", "foo"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show: %v", err)
	}
	var scoped config.ScopedSource
	if err := json.Unmarshal(buffer.Bytes(), &scoped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if scoped.Scope != scope.ScopeLocal {
		t.Fatalf("expected local scope, got %s", scoped.Scope)
	}

	cmd, buffer = newTestCommand()
	cmd.SetArgs([]string{"show", "foo", "--scope", "global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("show global: %v", err)
	}
	if err := json.Unmarshal(buffer.Bytes(), &scoped); err != nil {
		t.Fatalf("unmarshal global: %v", err)
	}
	if scoped.Scope != scope.ScopeGlobal {
		t.Fatalf("expected global scope, got %s", scoped.Scope)
	}
}
