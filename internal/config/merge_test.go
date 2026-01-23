package config

import (
	"testing"

	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

func TestMergeSourcesLocalWins(t *testing.T) {
	local := []Source{{Name: "foo", Type: "path", Origin: "local"}}
	global := []Source{{Name: "foo", Type: "git", Origin: "global"}}

	merged := MergeSources(local, global)
	if len(merged) != 1 {
		t.Fatalf("expected 1 source, got %d", len(merged))
	}
	if merged[0].Origin != "local" || merged[0].Scope != scope.ScopeLocal {
		t.Fatalf("expected local source to win")
	}
}

func TestMergeTargetsLocalWins(t *testing.T) {
	local := []Target{{Name: "foo", Path: "local"}}
	global := []Target{{Name: "foo", Path: "global"}}

	merged := MergeTargets(local, global)
	if len(merged) != 1 {
		t.Fatalf("expected 1 target, got %d", len(merged))
	}
	if merged[0].Path != "local" || merged[0].Scope != scope.ScopeLocal {
		t.Fatalf("expected local target to win")
	}
}
