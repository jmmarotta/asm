package asm

import (
	"fmt"
	"os"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Remove(names []string) (RemoveReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return RemoveReport{}, err
	}

	uniqueNames := uniqueRemoveNames(names)
	removed := make([]SkillSummary, 0, len(uniqueNames))
	warnings := []string{}
	originOrder := []string{}
	originSeen := map[string]bool{}

	for _, name := range uniqueNames {
		skill, ok := state.Config.RemoveSkill(name)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("skill %q not found", name))
			continue
		}
		removed = append(removed, SkillSummary{
			Name:    skill.Name,
			Type:    skill.Type,
			Origin:  skill.Origin,
			Version: skill.Version,
			Subdir:  skill.Subdir,
		})
		if skill.Type == "git" {
			if !originSeen[skill.Origin] {
				originSeen[skill.Origin] = true
				originOrder = append(originOrder, skill.Origin)
			}
		}
	}

	if len(removed) == 0 {
		return RemoveReport{Warnings: warnings, NoChanges: true}, nil
	}

	prunedStores := []string{}
	for _, origin := range originOrder {
		if !originInUse(state.Config, origin) {
			delete(state.Config.Replace, origin)
			deleteLockForOrigin(state.Lock, origin)
			if err := os.RemoveAll(gitstore.RepoPath(state.Paths.StoreDir, origin)); err != nil {
				return RemoveReport{}, err
			}
			prunedStores = append(prunedStores, origin)
		}
	}

	if err := manifest.SaveState(state); err != nil {
		return RemoveReport{}, err
	}

	report, err := installSkills(state)
	if err != nil {
		return RemoveReport{}, fmt.Errorf("install skills: %w", err)
	}

	return RemoveReport{
		Install:      report,
		Removed:      removed,
		PrunedStores: prunedStores,
		Warnings:     warnings,
	}, nil
}

func uniqueRemoveNames(names []string) []string {
	unique := make([]string, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		unique = append(unique, name)
	}
	return unique
}

func originInUse(configValue manifest.Config, origin string) bool {
	for _, skill := range configValue.Skills {
		if skill.Type == "git" && skill.Origin == origin {
			return true
		}
	}
	return false
}

func deleteLockForOrigin(lock map[manifest.LockKey]string, origin string) {
	for key := range lock {
		if key.Origin == origin {
			delete(lock, key)
		}
	}
}
