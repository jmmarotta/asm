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

	sources, warnings, lockChanged, err := resolveInstallSources(state)
	if err != nil {
		return InstallReport{}, fmt.Errorf("resolve sources: %w", err)
	}

	result, err := linker.SyncAndPrune(linker.Target{Name: "skills", Path: state.Paths.SkillsDir}, sources)
	if err != nil {
		return InstallReport{}, err
	}

	warnings = append(warnings, result.Warnings...)

	if lockChanged {
		if err := manifest.SaveLockWithSkills(state.LockPath, state.Lock, state.Config.Skills); err != nil {
			return InstallReport{}, err
		}
	}

	return InstallReport{Linked: result.Linked, Pruned: result.Removed, Warnings: warnings}, nil
}

func resolveInstallSources(state manifest.State) ([]linker.Source, []linker.Warning, bool, error) {
	originVersions := state.Config.GitOriginVersions()
	originPaths := make(map[string]string)
	warnings := []linker.Warning{}
	lockChanged := false
	if len(originVersions) > 0 {
		result, err := gitstore.ResolveOrigins(state.Paths.StoreDir, originVersions, state.Config.Replace, state.Lock, true)
		if err != nil {
			return nil, nil, false, err
		}
		lockChanged = result.LockChanged
		originPaths = result.Paths
		for _, warning := range result.Warnings {
			warnings = append(warnings, linker.Warning{Message: warning})
		}
	}

	sources, err := linker.SourcesFromConfig(state.Config, originPaths)
	if err != nil {
		return nil, nil, false, err
	}
	return sources, warnings, lockChanged, nil
}
