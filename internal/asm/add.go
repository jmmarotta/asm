package asm

import (
	"fmt"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
	"github.com/jmmarotta/agent_skills_manager/internal/source"
)

type addResolution struct {
	Type        string
	Origin      string
	RepoPath    string
	Version     string
	Rev         string
	ReplacePath string
}

func Add(input string, pathFlag string) (InstallReport, error) {
	state, _, err := manifest.LoadOrInitState()
	if err != nil {
		return InstallReport{}, fmt.Errorf("load manifest: %w", err)
	}

	pathFlag = strings.TrimSpace(pathFlag)
	debug.Logf("add start input=%q path=%q", input, pathFlag)

	inputSpec, err := parseAddInput(input, pathFlag)
	if err != nil {
		return InstallReport{}, fmt.Errorf("parse add input: %w", err)
	}

	resolution, err := resolveAddInput(state, inputSpec)
	if err != nil {
		return InstallReport{}, fmt.Errorf("resolve add input: %w", err)
	}

	skills, err := source.DiscoverSkills(resolution.RepoPath, inputSpec.Subdir)
	if err != nil {
		return InstallReport{}, fmt.Errorf("discover skills: %w", err)
	}

	if state.Config.Replace == nil {
		state.Config.Replace = map[string]string{}
	}
	if resolution.ReplacePath != "" {
		state.Config.Replace[resolution.Origin] = resolution.ReplacePath
	}

	author := resolveAuthor(inputSpec, resolution)
	discovered := make([]manifest.DiscoveredSkill, 0, len(skills))
	for _, skill := range skills {
		discovered = append(discovered, manifest.DiscoveredSkill{
			Name:   skill.Name,
			Subdir: skill.Subdir,
		})
	}
	if err := state.Config.UpsertDiscoveredSkills(discovered, manifest.UpsertOptions{
		Origin:  resolution.Origin,
		Type:    resolution.Type,
		Version: resolution.Version,
		Author:  author,
	}); err != nil {
		return InstallReport{}, err
	}

	if resolution.Type == "git" {
		if state.Sum == nil {
			state.Sum = map[manifest.SumKey]string{}
		}
		state.Sum[manifest.SumKey{Origin: resolution.Origin, Version: resolution.Version}] = resolution.Rev
	}

	if err := manifest.SaveState(state); err != nil {
		return InstallReport{}, fmt.Errorf("save manifest: %w", err)
	}

	report, err := installSkills(state)
	if err != nil {
		return InstallReport{}, fmt.Errorf("install skills: %w", err)
	}

	return report, nil
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
			refs, err := gitstore.ListRemoteRefs(origin)
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

func resolveAddInput(state manifest.State, inputSpec source.Input) (addResolution, error) {
	debug.Logf(
		"resolve add input origin=%s local=%t ref=%q subdir=%q",
		debug.SanitizeOrigin(inputSpec.Origin),
		inputSpec.IsLocal,
		inputSpec.Ref,
		inputSpec.Subdir,
	)
	if inputSpec.IsLocal {
		if inputSpec.RepoRoot != "" {
			originURL, ok, err := gitstore.OriginURL(inputSpec.RepoRoot)
			if err != nil {
				return addResolution{}, err
			}
			if ok {
				origin := source.NormalizeOrigin(originURL)
				resolved, err := gitstore.ResolveForRefAt(inputSpec.RepoRoot, inputSpec.Ref)
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

	repoPath := gitstore.RepoPath(state.Paths.StoreDir, inputSpec.Origin)
	if err := gitstore.EnsureRepo(repoPath, inputSpec.Origin); err != nil {
		return addResolution{}, err
	}
	resolved, err := gitstore.ResolveForRefAt(repoPath, inputSpec.Ref)
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
