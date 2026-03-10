package asm

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Remove(names []string, origins []string) (RemoveReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return RemoveReport{}, err
	}

	targets, warnings, err := resolveRemoveSkills(state.Config, names, origins)
	if err != nil {
		return RemoveReport{}, err
	}

	removed := make([]SkillSummary, 0, len(targets))
	originOrder := []string{}
	originSeen := map[string]bool{}

	for _, target := range targets {
		skill, ok := state.Config.RemoveSkill(target.Name)
		if !ok {
			return RemoveReport{}, fmt.Errorf("remove skill %q: not found", target.Name)
		}
		removed = append(removed, SkillSummary{
			Name:    skill.Name,
			Origin:  skill.Origin,
			Version: skill.Version,
			Subdir:  skill.Subdir,
		})
		if skill.Version != "" {
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

	report, err := installSkills(state, nil)
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

func resolveRemoveSkills(configValue manifest.Config, names []string, origins []string) ([]manifest.Skill, []string, error) {
	uniqueNames := uniqueRemoveValues(names)
	uniqueOrigins := uniqueRemoveValues(origins)
	selected := make(map[string]bool)
	removed := make([]manifest.Skill, 0, len(uniqueNames)+len(uniqueOrigins))
	warnings := []string{}

	for _, name := range uniqueNames {
		skill, ok := manifest.FindSkill(configValue.Skills, name)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("skill %q not found", name))
			continue
		}
		selected[skill.Name] = true
		removed = append(removed, skill)
	}

	for _, selector := range uniqueOrigins {
		normalized, err := normalizeOriginSelector(selector)
		if err != nil {
			return nil, nil, err
		}

		matches := findSkillsByOrigin(configValue.Skills, normalized)
		if len(matches) == 0 {
			warnings = append(warnings, fmt.Sprintf("origin %q not found", selector))
			continue
		}

		for _, skill := range matches {
			if selected[skill.Name] {
				continue
			}
			selected[skill.Name] = true
			removed = append(removed, skill)
		}
	}

	return removed, warnings, nil
}

func uniqueRemoveValues(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func findSkillsByOrigin(skills []manifest.Skill, origin string) []manifest.Skill {
	matched := []manifest.Skill{}
	for _, skill := range skills {
		if sameOrigin(skill.Origin, origin) {
			matched = append(matched, skill)
		}
	}
	return matched
}

func originInUse(configValue manifest.Config, origin string) bool {
	for _, skill := range configValue.Skills {
		if skill.Version != "" && skill.Origin == origin {
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
