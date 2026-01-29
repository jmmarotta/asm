package asm

import (
	"fmt"
	"sort"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Update(name string) (UpdateReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return UpdateReport{}, fmt.Errorf("load manifest: %w", err)
	}
	debug.Logf("update start name=%s", name)

	if len(state.Config.Skills) == 0 {
		return UpdateReport{Install: InstallReport{NoSkills: true}}, nil
	}

	origins, err := resolveUpdateOrigins(state.Config, name)
	if err != nil {
		return UpdateReport{}, err
	}
	updatedOrigins := make([]string, 0, len(origins))
	for origin := range origins {
		updatedOrigins = append(updatedOrigins, origin)
	}
	sort.Strings(updatedOrigins)

	if state.Lock == nil {
		state.Lock = map[manifest.LockKey]string{}
	}

	for origin, versionValue := range origins {
		debug.Logf("update origin=%s version=%s", debug.SanitizeOrigin(origin), versionValue)
		replacePath := ""
		if state.Config.Replace != nil {
			replacePath = state.Config.Replace[origin]
		}
		_, err := gitstore.ResolveOriginRevision(state.Paths.StoreDir, origin, versionValue, replacePath, state.Lock, false)
		if err != nil {
			return UpdateReport{}, err
		}
	}

	if err := manifest.SaveState(state); err != nil {
		return UpdateReport{}, fmt.Errorf("save manifest: %w", err)
	}

	report, err := installSkills(state)
	if err != nil {
		return UpdateReport{}, fmt.Errorf("install skills: %w", err)
	}

	return UpdateReport{Install: report, UpdatedOrigins: updatedOrigins}, nil
}

func resolveUpdateOrigins(configValue manifest.Config, name string) (map[string]string, error) {
	origins := make(map[string]string)
	if name == "" {
		for _, skill := range configValue.Skills {
			if skill.Version == "" {
				continue
			}
			origins[skill.Origin] = skill.Version
		}
		return origins, nil
	}

	skill, found := manifest.FindSkill(configValue.Skills, name)
	if !found {
		return nil, fmt.Errorf("skill %q not found", name)
	}
	if skill.Version == "" {
		return nil, fmt.Errorf("skill %q does not have a version", name)
	}
	origins[skill.Origin] = skill.Version
	return origins, nil
}
