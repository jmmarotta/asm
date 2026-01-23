package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/paths"
	"github.com/jmmarotta/agent_skills_manager/internal/syncer"
)

const syncScopeFlag = "scope"

func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync skills to targets",
		RunE:  runSync,
	}

	cmd.Flags().String(syncScopeFlag, "", "Scope to sync (local|global)")
	return cmd
}

func runSync(cmd *cobra.Command, _ []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	scopeFlag, err := cmd.Flags().GetString(syncScopeFlag)
	if err != nil {
		return err
	}
	scopeFlag = strings.TrimSpace(scopeFlag)

	sources, err := resolveSyncSources(configs, scopeFlag)
	if err != nil {
		return err
	}
	targets, err := resolveSyncTargets(configs)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No targets found.")
		return nil
	}
	if len(sources) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No sources found.")
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

	result, err := syncer.Sync(buildSyncTargets(targets), sourceLinks)
	if err != nil {
		return err
	}

	printWarnings(result.Warnings, cmd.ErrOrStderr())
	fmt.Fprintf(cmd.OutOrStdout(), "Synced: %d, Warnings: %d\n", result.Linked, len(result.Warnings))
	return nil
}
