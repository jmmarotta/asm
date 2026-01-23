package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

const listScopeFlag = "scope"

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured sources",
		RunE:  runList,
	}

	cmd.Flags().String(listScopeFlag, "", "Scope to list (local|global|all)")

	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	scopeFlag, err := cmd.Flags().GetString(listScopeFlag)
	if err != nil {
		return err
	}
	scopeFlag = strings.TrimSpace(scopeFlag)
	if scopeFlag == "" {
		if configs.InRepo {
			scopeFlag = string(scope.ScopeAll)
		} else {
			scopeFlag = string(scope.ScopeGlobal)
		}
	}

	requested, err := scope.ParseScope(scopeFlag)
	if err != nil {
		return err
	}

	if requested == scope.ScopeLocal && !configs.InRepo {
		return fmt.Errorf("local scope requires a git repo")
	}
	if requested == scope.ScopeAll && !configs.InRepo {
		requested = scope.ScopeGlobal
	}
	if requested == scope.ScopeEffective {
		return fmt.Errorf("scope must be local, global, or all")
	}

	var sources []config.ScopedSource
	switch requested {
	case scope.ScopeGlobal:
		sources = config.ScopedSourcesFrom(configs.Global.Sources, scope.ScopeGlobal)
	case scope.ScopeLocal:
		sources = config.ScopedSourcesFrom(configs.Local.Sources, scope.ScopeLocal)
	case scope.ScopeAll:
		sources = config.MergeSources(configs.Local.Sources, configs.Global.Sources)
	default:
		return fmt.Errorf("unsupported scope: %s", requested)
	}

	if len(sources) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sources found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tTYPE\tORIGIN\tREF\tSUBDIR\tSCOPE")
	for _, source := range sources {
		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\t%s\n",
			source.Name,
			source.Type,
			source.Origin,
			source.Ref,
			source.Subdir,
			source.Scope,
		)
	}

	return writer.Flush()
}
