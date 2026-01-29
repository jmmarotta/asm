package asm

import "github.com/jmmarotta/agent_skills_manager/internal/linker"

type InstallReport struct {
	Linked   int
	Pruned   int
	Warnings []linker.Warning
	NoSkills bool
}

type SkillSummary struct {
	Name    string
	Origin  string
	Version string
	Subdir  string
}

type ListReport struct {
	Skills   []SkillSummary
	NoSkills bool
}

type ShowReport struct {
	Name    string `json:"name"`
	Origin  string `json:"origin"`
	Subdir  string `json:"subdir,omitempty"`
	Version string `json:"version,omitempty"`
	Replace string `json:"replace,omitempty"`
}

type InitReport struct {
	Root         string
	ManifestPath string
}

type UpdateReport struct {
	Install        InstallReport
	UpdatedOrigins []string
}

type RemoveReport struct {
	Install      InstallReport
	Removed      []SkillSummary
	PrunedStores []string
	Warnings     []string
	NoChanges    bool
}
