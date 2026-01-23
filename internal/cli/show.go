package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

const showScopeFlag = "scope"

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a configured source",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}

	cmd.Flags().String(showScopeFlag, "", "Scope to show (local|global)")

	return cmd
}

func runShow(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	scopeFlag, err := cmd.Flags().GetString(showScopeFlag)
	if err != nil {
		return err
	}
	scopeFlag = strings.TrimSpace(scopeFlag)
	if scopeFlag == "" {
		scopeFlag = string(scope.ScopeEffective)
	}

	requested, err := scope.ParseScope(scopeFlag)
	if err != nil {
		return err
	}

	if requested == scope.ScopeLocal && !configs.InRepo {
		return fmt.Errorf("local scope requires a git repo")
	}
	if requested == scope.ScopeAll {
		return fmt.Errorf("scope must be local or global")
	}

	name := args[0]
	var source config.Source
	var sourceScope scope.Scope
	var found bool

	switch requested {
	case scope.ScopeLocal:
		source, found = config.FindSource(configs.Local.Sources, name)
		sourceScope = scope.ScopeLocal
	case scope.ScopeGlobal:
		source, found = config.FindSource(configs.Global.Sources, name)
		sourceScope = scope.ScopeGlobal
	case scope.ScopeEffective:
		if configs.InRepo {
			source, found = config.FindSource(configs.Local.Sources, name)
			if found {
				sourceScope = scope.ScopeLocal
				break
			}
		}
		source, found = config.FindSource(configs.Global.Sources, name)
		sourceScope = scope.ScopeGlobal
	default:
		return fmt.Errorf("unsupported scope: %s", requested)
	}

	if !found {
		return fmt.Errorf("source %q not found", name)
	}

	scoped := config.ScopedSource{Source: source, Scope: sourceScope}
	payload, err := json.MarshalIndent(scoped, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(payload))
	return nil
}
