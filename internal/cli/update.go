package cli

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/remote"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
	"github.com/jmmarotta/agent_skills_manager/internal/version"
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update skill revisions",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runUpdate,
	}

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) error {
	state, err := loadManifest()
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}
	debug.Logf("update start args=%v", args)

	if len(state.Config.Skills) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No skills found.")
		return nil
	}

	origins, err := resolveUpdateOrigins(state.Config, args)
	if err != nil {
		return err
	}

	if state.Sum == nil {
		state.Sum = map[config.SumKey]string{}
	}

	for origin, versionValue := range origins {
		debug.Logf("update origin=%s version=%s", debug.SanitizeOrigin(origin), versionValue)
		path := store.RepoPath(state.Paths.StoreDir, origin)
		if err := remote.EnsureRepo(path, origin); err != nil {
			return err
		}
		repo, err := git.PlainOpen(path)
		if err != nil {
			return fmt.Errorf("open repo %s: %w", path, err)
		}
		rev, err := version.ResolveForVersion(repo, versionValue)
		if err != nil {
			return fmt.Errorf("resolve version %s: %w", versionValue, err)
		}
		state.Sum[config.SumKey{Origin: origin, Version: versionValue}] = rev
	}

	if err := saveManifest(state); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := installSkills(cmd.OutOrStdout(), cmd.ErrOrStderr(), state); err != nil {
		return fmt.Errorf("install skills: %w", err)
	}

	return nil
}

func resolveUpdateOrigins(configValue config.Config, args []string) (map[string]string, error) {
	origins := make(map[string]string)
	if len(args) == 0 {
		for _, skill := range configValue.Skills {
			if skill.Type != "git" {
				continue
			}
			origins[skill.Origin] = skill.Version
		}
		return origins, nil
	}

	skill, found := config.FindSkill(configValue.Skills, args[0])
	if !found {
		return nil, fmt.Errorf("skill %q not found", args[0])
	}
	if skill.Type != "git" {
		return nil, fmt.Errorf("skill %q is not a git source", args[0])
	}
	origins[skill.Origin] = skill.Version
	return origins, nil
}
