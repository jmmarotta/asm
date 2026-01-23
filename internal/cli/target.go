package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/paths"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
	"github.com/jmmarotta/agent_skills_manager/internal/syncer"
)

const (
	targetScopeFlag  = "scope"
	targetLocalFlag  = "local"
	targetGlobalFlag = "global"
)

func newTargetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target",
		Short: "Manage sync targets",
	}

	cmd.AddCommand(newTargetListCommand())
	cmd.AddCommand(newTargetAddCommand())
	cmd.AddCommand(newTargetRemoveCommand())
	return cmd
}

func newTargetListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured targets",
		RunE:  runTargetList,
	}

	cmd.Flags().String(targetScopeFlag, "", "Scope to list (local|global|all)")

	return cmd
}

func runTargetList(cmd *cobra.Command, _ []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	scopeFlag, err := cmd.Flags().GetString(targetScopeFlag)
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

	var targets []config.ScopedTarget
	switch requested {
	case scope.ScopeGlobal:
		targets = config.ScopedTargetsFrom(configs.Global.Targets, scope.ScopeGlobal)
	case scope.ScopeLocal:
		targets = config.ScopedTargetsFrom(configs.Local.Targets, scope.ScopeLocal)
	case scope.ScopeAll:
		targets = config.MergeTargets(configs.Local.Targets, configs.Global.Targets)
	default:
		return fmt.Errorf("unsupported scope: %s", requested)
	}

	if len(targets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No targets found.")
		return nil
	}

	writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tPATH\tSCOPE")
	for _, target := range targets {
		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\n",
			target.Name,
			target.Path,
			target.Scope,
		)
	}

	return writer.Flush()
}

func newTargetAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <path>",
		Short: "Add a sync target",
		Args:  cobra.ExactArgs(2),
		RunE:  runTargetAdd,
	}

	cmd.Flags().Bool(targetLocalFlag, false, "Add target to local config")
	cmd.Flags().Bool(targetGlobalFlag, false, "Add target to global config")
	return cmd
}

func runTargetAdd(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	localFlag, err := cmd.Flags().GetBool(targetLocalFlag)
	if err != nil {
		return err
	}
	globalFlag, err := cmd.Flags().GetBool(targetGlobalFlag)
	if err != nil {
		return err
	}

	scoped, err := selectMutatingConfig(configs, localFlag, globalFlag)
	if err != nil {
		return err
	}

	path, err := resolveTargetPath(args[1])
	if err != nil {
		return err
	}

	scoped.Config.UpsertTarget(config.Target{
		Name: args[0],
		Path: path,
	})

	return config.Save(scoped.ConfigPath, scoped.Config)
}

func newTargetRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a sync target",
		Args:  cobra.ExactArgs(1),
		RunE:  runTargetRemove,
	}

	cmd.Flags().Bool(targetLocalFlag, false, "Remove target from local config")
	cmd.Flags().Bool(targetGlobalFlag, false, "Remove target from global config")
	return cmd
}

func runTargetRemove(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	localFlag, err := cmd.Flags().GetBool(targetLocalFlag)
	if err != nil {
		return err
	}
	globalFlag, err := cmd.Flags().GetBool(targetGlobalFlag)
	if err != nil {
		return err
	}

	scoped, err := selectMutatingConfig(configs, localFlag, globalFlag)
	if err != nil {
		return err
	}

	removed, ok := scoped.Config.RemoveTarget(args[0])
	if !ok {
		return fmt.Errorf("target %q not found", args[0])
	}

	if err := config.Save(scoped.ConfigPath, scoped.Config); err != nil {
		return err
	}

	if scoped.Scope == scope.ScopeLocal {
		configs.Local = scoped.Config
	} else {
		configs.Global = scoped.Config
	}

	effectiveTargets, err := resolveSyncTargets(configs)
	if err != nil {
		return err
	}
	if targetExists(effectiveTargets, removed.Name) {
		return nil
	}

	sources, err := resolveSyncSources(configs, "")
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return nil
	}

	globalPaths, err := paths.GlobalPaths()
	if err != nil {
		return err
	}
	localPaths := paths.Paths{}
	if configs.InRepo {
		localPaths = paths.LocalPaths(configs.RepoRoot)
	}

	sourceLinks, err := buildSyncSources(sources, globalPaths.StoreDir, localPaths.StoreDir)
	if err != nil {
		return err
	}

	result, err := syncer.Cleanup(syncer.Target{Name: removed.Name, Path: removed.Path}, sourceLinks)
	if err != nil {
		return err
	}

	printWarnings(result.Warnings, cmd.ErrOrStderr())
	return nil
}

func resolveTargetPath(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("path is required")
	}

	expanded, err := expandHomeInput(value)
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(expanded) {
		return filepath.Abs(expanded)
	}

	return filepath.Clean(expanded), nil
}

func expandHomeInput(value string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if value == "~" {
		return home, nil
	}
	if strings.HasPrefix(value, "~/") {
		return filepath.Join(home, strings.TrimPrefix(value, "~/")), nil
	}
	if value == "$HOME" {
		return home, nil
	}
	if strings.HasPrefix(value, "$HOME/") {
		return filepath.Join(home, strings.TrimPrefix(value, "$HOME/")), nil
	}

	return value, nil
}

func targetExists(targets []config.ScopedTarget, name string) bool {
	for _, target := range targets {
		if target.Name == name {
			return true
		}
	}
	return false
}
