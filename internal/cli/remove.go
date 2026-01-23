package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
)

const (
	removeLocalFlag  = "local"
	removeGlobalFlag = "global"
)

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a configured source",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}

	cmd.Flags().Bool(removeLocalFlag, false, "Remove source from local config")
	cmd.Flags().Bool(removeGlobalFlag, false, "Remove source from global config")

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	localFlag, err := cmd.Flags().GetBool(removeLocalFlag)
	if err != nil {
		return err
	}
	globalFlag, err := cmd.Flags().GetBool(removeGlobalFlag)
	if err != nil {
		return err
	}

	scoped, err := selectMutatingConfig(configs, localFlag, globalFlag)
	if err != nil {
		return err
	}

	removed, ok := scoped.Config.RemoveSource(args[0])
	if !ok {
		return fmt.Errorf("source %q not found", args[0])
	}

	if err := config.Save(scoped.ConfigPath, scoped.Config); err != nil {
		return err
	}

	if removed.Type == "git" {
		repoPath := store.RepoPath(scoped.Paths.StoreDir, removed.Origin, removed.Ref)
		if !repoKeyExists(scoped.Config, removed.Origin, removed.Ref) {
			if err := os.RemoveAll(repoPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func repoKeyExists(cfg config.Config, origin string, ref string) bool {
	key := store.RepoKey(origin, ref)
	for _, source := range cfg.Sources {
		if source.Type != "git" {
			continue
		}
		if store.RepoKey(source.Origin, source.Ref) == key {
			return true
		}
	}
	return false
}
