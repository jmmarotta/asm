package asm

import (
	"fmt"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func Show(name string) (ShowReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return ShowReport{}, err
	}

	skill, found := manifest.FindSkill(state.Config.Skills, name)
	if !found {
		return ShowReport{}, fmt.Errorf("skill %q not found", name)
	}

	return ShowReport{
		Name:    skill.Name,
		Type:    skill.Type,
		Origin:  skill.Origin,
		Subdir:  skill.Subdir,
		Version: skill.Version,
		Replace: state.Config.Replace[skill.Origin],
	}, nil
}
