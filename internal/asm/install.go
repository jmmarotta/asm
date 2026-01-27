package asm

import (
	"fmt"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/linker"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Install() (InstallReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return InstallReport{}, err
	}

	return installSkills(state)
}

func installSkills(state manifest.State) (InstallReport, error) {
	debug.Logf("install skills count=%d", len(state.Config.Skills))
	if len(state.Config.Skills) == 0 {
		prune, err := linker.Prune(linker.Target{Name: "skills", Path: state.Paths.SkillsDir}, nil)
		if err != nil {
			return InstallReport{}, err
		}
		return InstallReport{Pruned: prune.Removed, Warnings: prune.Warnings, NoSkills: true}, nil
	}

	sources, warnings, sumChanged, err := resolveInstallSources(state)
	if err != nil {
		return InstallReport{}, fmt.Errorf("resolve sources: %w", err)
	}

	result, err := linker.SyncAndPrune(linker.Target{Name: "skills", Path: state.Paths.SkillsDir}, sources)
	if err != nil {
		return InstallReport{}, err
	}

	warnings = append(warnings, result.Warnings...)

	if sumChanged {
		if err := manifest.SaveSum(state.SumPath, state.Sum); err != nil {
			return InstallReport{}, err
		}
	}

	return InstallReport{Linked: result.Linked, Pruned: result.Removed, Warnings: warnings}, nil
}

func resolveInstallSources(state manifest.State) ([]linker.Source, []linker.Warning, bool, error) {
	originVersions := state.Config.GitOriginVersions()
	originPaths := make(map[string]string)
	warnings := []linker.Warning{}
	sumChanged := false
	if len(originVersions) > 0 {
		result, err := gitstore.ResolveOrigins(state.Paths.StoreDir, originVersions, state.Config.Replace, state.Sum, true)
		if err != nil {
			return nil, nil, false, err
		}
		sumChanged = result.SumChanged
		originPaths = result.Paths
		for _, warning := range result.Warnings {
			warnings = append(warnings, linker.Warning{Message: warning})
		}
	}

	sources, err := linker.SourcesFromConfig(state.Config, originPaths)
	if err != nil {
		return nil, nil, false, err
	}
	return sources, warnings, sumChanged, nil
}
