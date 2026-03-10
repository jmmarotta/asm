package asm

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

type ListSort string

const (
	ListSortName   ListSort = "name"
	ListSortOrigin ListSort = "origin"
)

type ListOptions struct {
	Sort ListSort
}

func ParseListSort(value string) (ListSort, error) {
	switch ListSort(strings.ToLower(strings.TrimSpace(value))) {
	case "", ListSortName:
		return ListSortName, nil
	case ListSortOrigin:
		return ListSortOrigin, nil
	default:
		return "", fmt.Errorf("invalid value for --sort %q: expected name or origin", value)
	}
}

func List(options ListOptions) (ListReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return ListReport{}, err
	}

	if len(state.Config.Skills) == 0 {
		return ListReport{NoSkills: true}, nil
	}

	orderedSkills := append([]manifest.Skill(nil), state.Config.Skills...)
	if options.Sort == ListSortOrigin {
		sortSkillsByOrigin(orderedSkills)
	} else {
		manifest.SortSkills(orderedSkills)
	}
	report := ListReport{Skills: make([]SkillSummary, 0, len(orderedSkills))}
	for _, skill := range orderedSkills {
		report.Skills = append(report.Skills, SkillSummary{
			Name:    skill.Name,
			Origin:  skill.Origin,
			Version: skill.Version,
			Subdir:  skill.Subdir,
		})
	}

	return report, nil
}

func sortSkillsByOrigin(skills []manifest.Skill) {
	sort.Slice(skills, func(i, j int) bool {
		left := skills[i]
		right := skills[j]
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.Subdir != right.Subdir {
			return left.Subdir < right.Subdir
		}
		return left.Version < right.Version
	})
}
