package manifest

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type Config struct {
	Skills  []Skill           `json:"skills"`
	Replace map[string]string `json:"replace,omitempty"`
}

type Skill struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Origin  string `json:"origin"`
	Subdir  string `json:"subdir,omitempty"`
	Version string `json:"version,omitempty"`
}

type SumKey struct {
	Origin  string
	Version string
}

func (config *Config) Validate() error {
	versions := make(map[string]string)
	names := make(map[string]int)
	identities := make(map[skillIdentity]int)
	for index, skill := range config.Skills {
		if err := skill.Validate(index); err != nil {
			return err
		}
		if prior, ok := names[skill.Name]; ok {
			return fmt.Errorf("skills[%d]: duplicate name %q (already at skills[%d])", index, skill.Name, prior)
		}
		names[skill.Name] = index

		normalizedSubdir, err := normalizeSubdir(skill.Subdir)
		if err != nil {
			return fmt.Errorf("skills[%d]: invalid subdir %q: %w", index, skill.Subdir, err)
		}
		identity := skillIdentity{origin: skill.Origin, subdir: normalizedSubdir}
		if prior, ok := identities[identity]; ok {
			return fmt.Errorf("skills[%d]: origin %q subdir %q already used by skills[%d]", index, skill.Origin, normalizedSubdir, prior)
		}
		identities[identity] = index
		if skill.Type != "git" {
			continue
		}
		if existing, ok := versions[skill.Origin]; ok && existing != skill.Version {
			return fmt.Errorf("skills[%d]: origin %q uses multiple versions", index, skill.Origin)
		}
		versions[skill.Origin] = skill.Version
	}

	return nil
}

func (config *Config) UpsertSkill(skill Skill) {
	for index, existing := range config.Skills {
		if existing.Name == skill.Name {
			config.Skills[index] = skill
			return
		}
	}
	config.Skills = append(config.Skills, skill)
}

func (config *Config) RemoveSkill(name string) (Skill, bool) {
	for index, skill := range config.Skills {
		if skill.Name == name {
			config.Skills = append(config.Skills[:index], config.Skills[index+1:]...)
			return skill, true
		}
	}
	return Skill{}, false
}

func FindSkill(skills []Skill, name string) (Skill, bool) {
	for _, skill := range skills {
		if skill.Name == name {
			return skill, true
		}
	}
	return Skill{}, false
}

func (skill Skill) Validate(index int) error {
	if skill.Name == "" {
		return fmt.Errorf("skills[%d]: missing name", index)
	}
	if skill.Type == "" {
		return fmt.Errorf("skills[%d]: missing type", index)
	}
	if skill.Origin == "" {
		return fmt.Errorf("skills[%d]: missing origin", index)
	}
	if skill.Type != "git" && skill.Type != "path" {
		return fmt.Errorf("skills[%d]: invalid type %q", index, skill.Type)
	}
	if skill.Type == "git" && skill.Version == "" {
		return fmt.Errorf("skills[%d]: missing version", index)
	}
	if skill.Type == "git" && skill.Version != "" {
		if !semver.IsValid(skill.Version) && !module.IsPseudoVersion(skill.Version) {
			return fmt.Errorf("skills[%d]: invalid version %q", index, skill.Version)
		}
	}
	if skill.Type == "path" && skill.Version != "" {
		return fmt.Errorf("skills[%d]: path skills cannot set version", index)
	}
	return nil
}

type skillIdentity struct {
	origin string
	subdir string
}

func normalizeSubdir(subdir string) (string, error) {
	if subdir == "" {
		return "", nil
	}
	if filepath.IsAbs(subdir) {
		return "", fmt.Errorf("path must be relative")
	}
	cleaned := filepath.Clean(subdir)
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path cannot escape repo")
	}
	return filepath.ToSlash(cleaned), nil
}
