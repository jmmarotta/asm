package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitutil"
	"github.com/jmmarotta/agent_skills_manager/internal/remote"
	"github.com/jmmarotta/agent_skills_manager/internal/source"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
	"github.com/jmmarotta/agent_skills_manager/internal/version"
)

const addPathFlag = "path"

type addResolution struct {
	Type        string
	Origin      string
	RepoPath    string
	Version     string
	Rev         string
	ReplacePath string
}

type skillIdentity struct {
	Origin string
	Subdir string
}

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <path-or-url>",
		Short: "Add a skill",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}

	cmd.Flags().String(addPathFlag, "", "Subdirectory path to install")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	state, _, err := loadOrInitManifest()
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	pathFlag, err := cmd.Flags().GetString(addPathFlag)
	if err != nil {
		return err
	}
	pathFlag = strings.TrimSpace(pathFlag)
	debug.Logf("add start input=%q path=%q", args[0], pathFlag)

	inputSpec, err := parseAddInput(args[0], pathFlag)
	if err != nil {
		return fmt.Errorf("parse add input: %w", err)
	}

	resolution, err := resolveAddInput(state, inputSpec)
	if err != nil {
		return fmt.Errorf("resolve add input: %w", err)
	}

	skills, err := source.DiscoverSkills(resolution.RepoPath, inputSpec.Subdir)
	if err != nil {
		return fmt.Errorf("discover skills: %w", err)
	}

	if state.Config.Replace == nil {
		state.Config.Replace = map[string]string{}
	}
	if resolution.ReplacePath != "" {
		state.Config.Replace[resolution.Origin] = resolution.ReplacePath
	}

	existingByName := make(map[string]skillIdentity)
	existingByIdentity := make(map[skillIdentity]string)
	for _, skill := range state.Config.Skills {
		identity := skillIdentity{Origin: skill.Origin, Subdir: normalizeSubdir(skill.Subdir)}
		if prior, ok := existingByIdentity[identity]; ok && prior != skill.Name {
			return fmt.Errorf("duplicate skills for origin %q subdir %q: %q and %q", skill.Origin, identity.Subdir, prior, skill.Name)
		}
		existingByIdentity[identity] = skill.Name
		existingByName[skill.Name] = identity
	}

	if resolution.Type == "git" {
		for index, skill := range state.Config.Skills {
			if skill.Type == "git" && skill.Origin == resolution.Origin {
				skill.Version = resolution.Version
				state.Config.Skills[index] = skill
			}
		}
	}

	author := resolveAuthor(inputSpec, resolution)
	for _, skill := range skills {
		subdir := normalizeSubdir(skill.Subdir)
		identity := skillIdentity{Origin: resolution.Origin, Subdir: subdir}
		name := skill.Name
		if existingName, ok := existingByIdentity[identity]; ok {
			name = existingName
		} else if existingIdentity, ok := existingByName[name]; ok && existingIdentity != identity {
			name = fmt.Sprintf("%s/%s", author, name)
			if collisionIdentity, collision := existingByName[name]; collision && collisionIdentity != identity {
				return fmt.Errorf("name collision for %q", skill.Name)
			}
		}

		entry := config.Skill{
			Name:   name,
			Type:   resolution.Type,
			Origin: resolution.Origin,
			Subdir: subdir,
		}
		if resolution.Type == "git" {
			entry.Version = resolution.Version
		}
		state.Config.UpsertSkill(entry)
		existingByIdentity[identity] = name
		existingByName[name] = identity
	}

	if resolution.Type == "git" {
		if state.Sum == nil {
			state.Sum = map[config.SumKey]string{}
		}
		state.Sum[config.SumKey{Origin: resolution.Origin, Version: resolution.Version}] = resolution.Rev
	}

	if err := saveManifest(state); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	if err := installSkills(cmd.OutOrStdout(), cmd.ErrOrStderr(), state); err != nil {
		return fmt.Errorf("install skills: %w", err)
	}

	return nil
}

func normalizeSubdir(subdir string) string {
	if subdir == "" {
		return ""
	}
	cleaned := filepath.ToSlash(filepath.Clean(subdir))
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func parseAddInput(input string, pathFlag string) (source.Input, error) {
	debug.Logf("parse add input raw=%q path=%q", input, pathFlag)
	if source.IsGitHubTreeURL(input) {
		if pathFlag != "" {
			return source.Input{}, fmt.Errorf("omit --path when using a github tree url")
		}

		origin, ok, err := source.GitHubTreeOrigin(input)
		if err != nil {
			return source.Input{}, fmt.Errorf("parse github tree origin: %w", err)
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

func resolveAddInput(state manifestState, inputSpec source.Input) (addResolution, error) {
	debug.Logf(
		"resolve add input origin=%s local=%t ref=%q subdir=%q",
		debug.SanitizeOrigin(inputSpec.Origin),
		inputSpec.IsLocal,
		inputSpec.Ref,
		inputSpec.Subdir,
	)
	if inputSpec.IsLocal {
		if inputSpec.RepoRoot != "" {
			originURL, ok, err := gitutil.OriginURL(inputSpec.RepoRoot)
			if err != nil {
				return addResolution{}, err
			}
			if ok {
				origin := source.NormalizeOrigin(originURL)
				repo, err := git.PlainOpen(inputSpec.RepoRoot)
				if err != nil {
					return addResolution{}, fmt.Errorf("open repo %s: %w", inputSpec.RepoRoot, err)
				}
				resolved, err := version.ResolveForRef(repo, inputSpec.Ref)
				if err != nil {
					return addResolution{}, fmt.Errorf("resolve ref %q: %w", inputSpec.Ref, err)
				}
				return addResolution{
					Type:        "git",
					Origin:      origin,
					RepoPath:    inputSpec.RepoRoot,
					Version:     resolved.Version,
					Rev:         resolved.Rev,
					ReplacePath: inputSpec.RepoRoot,
				}, nil
			}
		}

		return addResolution{
			Type:     "path",
			Origin:   inputSpec.Origin,
			RepoPath: inputSpec.Origin,
		}, nil
	}

	repoPath := store.RepoPath(state.Paths.StoreDir, inputSpec.Origin)
	if err := remote.EnsureRepo(repoPath, inputSpec.Origin); err != nil {
		return addResolution{}, err
	}
	remoteRepo, err := git.PlainOpen(repoPath)
	if err != nil {
		return addResolution{}, fmt.Errorf("open repo %s: %w", repoPath, err)
	}
	resolved, err := version.ResolveForRef(remoteRepo, inputSpec.Ref)
	if err != nil {
		return addResolution{}, fmt.Errorf("resolve ref %q: %w", inputSpec.Ref, err)
	}

	return addResolution{
		Type:     "git",
		Origin:   inputSpec.Origin,
		RepoPath: repoPath,
		Version:  resolved.Version,
		Rev:      resolved.Rev,
	}, nil
}

func resolveAuthor(inputSpec source.Input, resolution addResolution) string {
	if resolution.Type == "git" {
		return source.AuthorForRemoteOrigin(resolution.Origin)
	}
	if inputSpec.IsLocal {
		return source.AuthorForLocalPath(inputSpec.Origin)
	}
	return source.AuthorForRemoteOrigin(resolution.Origin)
}
