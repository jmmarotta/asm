package asm

import "github.com/jmmarotta/agent_skills_manager/internal/manifest"

func List() (ListReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return ListReport{}, err
	}

	if len(state.Config.Skills) == 0 {
		return ListReport{NoSkills: true}, nil
	}

	report := ListReport{Skills: make([]SkillSummary, 0, len(state.Config.Skills))}
	for _, skill := range state.Config.Skills {
		report.Skills = append(report.Skills, SkillSummary{
			Name:    skill.Name,
			Origin:  skill.Origin,
			Version: skill.Version,
			Subdir:  skill.Subdir,
		})
	}

	return report, nil
}
