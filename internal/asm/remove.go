package asm

import (
	"fmt"
	"os"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Remove(name string) (RemoveReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return RemoveReport{}, err
	}

	removed, ok := state.Config.RemoveSkill(name)
	if !ok {
		return RemoveReport{}, fmt.Errorf("skill %q not found", name)
	}
	removedSummary := SkillSummary{
		Name:    removed.Name,
		Type:    removed.Type,
		Origin:  removed.Origin,
		Version: removed.Version,
		Subdir:  removed.Subdir,
	}
	prunedStore := false

	if removed.Type == "git" {
		if !originInUse(state.Config, removed.Origin) {
			delete(state.Config.Replace, removed.Origin)
			deleteSumForOrigin(state.Sum, removed.Origin)
			if err := os.RemoveAll(gitstore.RepoPath(state.Paths.StoreDir, removed.Origin)); err != nil {
				return RemoveReport{}, err
			}
			prunedStore = true
		}
	}

	if err := manifest.SaveState(state); err != nil {
		return RemoveReport{}, err
	}

	report, err := installSkills(state)
	if err != nil {
		return RemoveReport{}, fmt.Errorf("install skills: %w", err)
	}

	return RemoveReport{Install: report, Removed: removedSummary, PrunedStore: prunedStore}, nil
}

func originInUse(configValue manifest.Config, origin string) bool {
	for _, skill := range configValue.Skills {
		if skill.Type == "git" && skill.Origin == origin {
			return true
		}
	}
	return false
}

func deleteSumForOrigin(sum map[manifest.SumKey]string, origin string) {
	for key := range sum {
		if key.Origin == origin {
			delete(sum, key)
		}
	}
}
