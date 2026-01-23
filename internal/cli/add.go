package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/remote"
	"github.com/jmmarotta/agent_skills_manager/internal/source"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
)

const (
	addLocalFlag  = "local"
	addGlobalFlag = "global"
	addPathFlag   = "path"
)

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <path-or-url>",
		Short: "Add a skill source",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}

	cmd.Flags().Bool(addLocalFlag, false, "Add source to local config")
	cmd.Flags().Bool(addGlobalFlag, false, "Add source to global config")
	cmd.Flags().String(addPathFlag, "", "Subdirectory path to install")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	configs, err := loadConfigSet()
	if err != nil {
		return err
	}

	localFlag, err := cmd.Flags().GetBool(addLocalFlag)
	if err != nil {
		return err
	}
	globalFlag, err := cmd.Flags().GetBool(addGlobalFlag)
	if err != nil {
		return err
	}

	scoped, err := selectMutatingConfig(configs, localFlag, globalFlag)
	if err != nil {
		return err
	}

	pathFlag, err := cmd.Flags().GetString(addPathFlag)
	if err != nil {
		return err
	}
	pathFlag = strings.TrimSpace(pathFlag)
	inputSpec, err := parseAddInput(args[0], pathFlag)
	if err != nil {
		return err
	}

	repoPath := inputSpec.LocalPath
	if !inputSpec.IsLocal {
		repoPath = store.RepoPath(scoped.Paths.StoreDir, inputSpec.Origin, inputSpec.Ref)
		if err := remote.EnsureRepo(repoPath, inputSpec.Origin, inputSpec.Ref); err != nil {
			return err
		}
	}

	skills, err := source.DiscoverSkills(repoPath, inputSpec.Subdir)
	if err != nil {
		return err
	}

	existing := make(map[string]struct{})
	for _, source := range scoped.Config.Sources {
		existing[source.Name] = struct{}{}
	}

	author := resolveAuthor(inputSpec)
	for _, skill := range skills {
		name := skill.Name
		if _, ok := existing[name]; ok {
			name = fmt.Sprintf("%s/%s", author, name)
			if _, collision := existing[name]; collision {
				return fmt.Errorf("name collision for %q", skill.Name)
			}
		}

		scoped.Config.UpsertSource(config.Source{
			Name:   name,
			Type:   inputSpec.Type,
			Origin: inputSpec.Origin,
			Subdir: filepath.ToSlash(skill.Subdir),
			Ref:    inputSpec.Ref,
		})
		existing[name] = struct{}{}
	}

	if err := config.Save(scoped.ConfigPath, scoped.Config); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Sync not implemented yet.")
	return nil
}

func parseAddInput(input string, pathFlag string) (source.Input, error) {
	if source.IsGitHubTreeURL(input) {
		if pathFlag != "" {
			return source.Input{}, fmt.Errorf("omit --path when using a github tree url")
		}

		origin, ok, err := source.GitHubTreeOrigin(input)
		if err != nil {
			return source.Input{}, err
		}
		if ok {
			refs, err := remote.ListRemoteRefs(origin)
			if err != nil {
				return source.Input{}, err
			}

			tree, _, err := source.ParseGitHubTreeURL(input, refs.All)
			if err != nil {
				return source.Input{}, fmt.Errorf("unable to parse github tree url; use origin@ref --path instead: %w", err)
			}

			return source.Input{
				Type:    "git",
				Origin:  source.NormalizeOrigin(tree.Origin),
				Ref:     tree.Ref,
				Subdir:  tree.Subdir,
				IsLocal: false,
			}, nil
		}
	}

	return source.ParseInput(input, pathFlag)
}

func resolveAuthor(inputSpec source.Input) string {
	if inputSpec.IsLocal {
		return source.AuthorForLocalPath(inputSpec.Origin)
	}

	return source.AuthorForRemoteOrigin(inputSpec.Origin)
}
