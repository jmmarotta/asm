package scope

import (
	"fmt"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/gitutil"
)

type Scope string

const (
	ScopeGlobal    Scope = "global"
	ScopeLocal     Scope = "local"
	ScopeAll       Scope = "all"
	ScopeEffective Scope = "effective"
)

func ParseScope(value string) (Scope, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case string(ScopeGlobal):
		return ScopeGlobal, nil
	case string(ScopeLocal):
		return ScopeLocal, nil
	case string(ScopeAll):
		return ScopeAll, nil
	case string(ScopeEffective):
		return ScopeEffective, nil
	case "":
		return "", fmt.Errorf("scope is required")
	default:
		return "", fmt.Errorf("invalid scope: %s", value)
	}
}

func FindRepoRoot(start string) (string, bool, error) {
	return gitutil.FindRepoRoot(start)
}

func (scope Scope) String() string {
	return string(scope)
}
