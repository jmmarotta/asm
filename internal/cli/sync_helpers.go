package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
	"github.com/jmmarotta/agent_skills_manager/internal/syncer"
)

func resolveSyncSources(configs configSet, scopeFlag string) ([]config.ScopedSource, error) {
	scopeFlag = strings.TrimSpace(scopeFlag)
	requested := scope.ScopeEffective
	if scopeFlag != "" {
		parsed, err := scope.ParseScope(scopeFlag)
		if err != nil {
			return nil, err
		}
		requested = parsed
	}

	if requested == scope.ScopeLocal && !configs.InRepo {
		return nil, fmt.Errorf("local scope requires a git repo")
	}
	if requested == scope.ScopeAll {
		return nil, fmt.Errorf("scope must be local or global")
	}
	if requested == scope.ScopeEffective {
		if configs.InRepo {
			return config.MergeSources(configs.Local.Sources, configs.Global.Sources), nil
		}
		requested = scope.ScopeGlobal
	}

	switch requested {
	case scope.ScopeGlobal:
		return config.ScopedSourcesFrom(configs.Global.Sources, scope.ScopeGlobal), nil
	case scope.ScopeLocal:
		return config.ScopedSourcesFrom(configs.Local.Sources, scope.ScopeLocal), nil
	default:
		return nil, fmt.Errorf("unsupported scope: %s", requested)
	}
}

func resolveSyncTargets(configs configSet) ([]config.ScopedTarget, error) {
	if configs.InRepo {
		return config.MergeTargets(configs.Local.Targets, configs.Global.Targets), nil
	}

	return config.ScopedTargetsFrom(configs.Global.Targets, scope.ScopeGlobal), nil
}

func buildSyncSources(sources []config.ScopedSource, globalStore string, localStore string) ([]syncer.Source, error) {
	links := make([]syncer.Source, 0, len(sources))
	for _, scopedSource := range sources {
		path, err := resolveSourcePath(scopedSource, globalStore, localStore)
		if err != nil {
			return nil, err
		}
		links = append(links, syncer.Source{
			Name: scopedSource.Name,
			Path: path,
		})
	}
	return links, nil
}

func buildSyncTargets(targets []config.ScopedTarget) []syncer.Target {
	links := make([]syncer.Target, 0, len(targets))
	for _, target := range targets {
		links = append(links, syncer.Target{
			Name: target.Name,
			Path: target.Path,
		})
	}
	return links
}

func resolveSourcePath(source config.ScopedSource, globalStore string, localStore string) (string, error) {
	if source.Type == "git" {
		storeDir := globalStore
		if source.Scope == scope.ScopeLocal {
			storeDir = localStore
		}
		if storeDir == "" {
			return "", fmt.Errorf("missing store directory for %s scope", source.Scope)
		}
		repoRoot := store.RepoPath(storeDir, source.Origin, source.Ref)
		if source.Subdir == "" {
			return repoRoot, nil
		}
		return filepath.Join(repoRoot, filepath.FromSlash(source.Subdir)), nil
	}

	if source.Type == "path" {
		base := source.Origin
		if source.Subdir == "" {
			return base, nil
		}
		return filepath.Join(base, filepath.FromSlash(source.Subdir)), nil
	}

	return "", fmt.Errorf("unsupported source type: %s", source.Type)
}

func printWarnings(warnings []syncer.Warning, writer io.Writer) {
	for _, warning := range warnings {
		if warning.Target != "" {
			fmt.Fprintf(writer, "warning: %s (%s)\n", warning.Message, warning.Target)
			continue
		}
		fmt.Fprintf(writer, "warning: %s\n", warning.Message)
	}
}
