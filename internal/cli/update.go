package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/remote"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
)

const (
	updateLocalFlag  = "local"
	updateGlobalFlag = "global"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update configured sources",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runUpdate,
	}

	cmd.Flags().Bool(updateLocalFlag, false, "Update sources in local config")
	cmd.Flags().Bool(updateGlobalFlag, false, "Update sources in global config")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	localFlag, err := cmd.Flags().GetBool(updateLocalFlag)
	if err != nil {
		return err
	}
	globalFlag, err := cmd.Flags().GetBool(updateGlobalFlag)
	if err != nil {
		return err
	}

	scoped, err := selectMutatingConfig(configs, localFlag, globalFlag)
	if err != nil {
		return err
	}

	sources, err := resolveSourcesForUpdate(scoped.Config, args)
	if err != nil {
		return err
	}

	for _, source := range sources {
		if source.Type == "git" {
			repoPath := store.RepoPath(scoped.Paths.StoreDir, source.Origin, source.Ref)
			if err := remote.EnsureRepo(repoPath, source.Origin, source.Ref); err != nil {
				return err
			}
			continue
		}

		target := source.Origin
		if source.Subdir != "" {
			target = filepath.Join(target, source.Subdir)
		}
		if _, err := os.Stat(target); err != nil {
			return fmt.Errorf("local path missing: %s", target)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Sync not implemented yet.")
	return nil
}

func resolveSourcesForUpdate(cfg config.Config, args []string) ([]config.Source, error) {
	if len(args) == 0 {
		return cfg.Sources, nil
	}

	source, found := config.FindSource(cfg.Sources, args[0])
	if !found {
		return nil, fmt.Errorf("source %q not found", args[0])
	}

	return []config.Source{source}, nil
}
