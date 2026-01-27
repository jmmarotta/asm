package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
)

func newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	state, err := loadManifest()
	if err != nil {
		return err
	}

	removed, ok := state.Config.RemoveSkill(args[0])
	if !ok {
		return fmt.Errorf("skill %q not found", args[0])
	}

	if removed.Type == "git" {
		if !originInUse(state.Config, removed.Origin) {
			delete(state.Config.Replace, removed.Origin)
			deleteSumForOrigin(state.Sum, removed.Origin)
			if err := os.RemoveAll(store.RepoPath(state.Paths.StoreDir, removed.Origin)); err != nil {
				return err
			}
		}
	}

	if err := saveManifest(state); err != nil {
		return err
	}

	return installSkills(cmd.OutOrStdout(), cmd.ErrOrStderr(), state)
}

func originInUse(configValue config.Config, origin string) bool {
	for _, skill := range configValue.Skills {
		if skill.Type == "git" && skill.Origin == origin {
			return true
		}
	}
	return false
}

func deleteSumForOrigin(sum map[config.SumKey]string, origin string) {
	for key := range sum {
		if key.Origin == origin {
			delete(sum, key)
		}
	}
}
