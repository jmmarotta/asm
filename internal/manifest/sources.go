package manifest

import (
	"fmt"
	"path/filepath"
)

type SkillPath struct {
	Name string
	Path string
}

type MissingOriginPathError struct {
	Origin string
	Skill  string
}

func (err MissingOriginPathError) Error() string {
	if err.Skill == "" {
		return fmt.Sprintf("missing origin path for %s", err.Origin)
	}
	return fmt.Sprintf("missing origin path for %s (skill %s)", err.Origin, err.Skill)
}

func (config Config) ResolveSkillPaths(originPaths map[string]string) ([]SkillPath, error) {
	paths := make([]SkillPath, 0, len(config.Skills))
	for _, skill := range config.Skills {
		base := skill.Origin
		if skill.Version != "" {
			resolved, ok := originPaths[skill.Origin]
			if !ok || resolved == "" {
				return nil, MissingOriginPathError{Origin: skill.Origin, Skill: skill.Name}
			}
			base = resolved
		}
		path := base
		if skill.Subdir != "" {
			path = filepath.Join(base, filepath.FromSlash(skill.Subdir))
		}
		paths = append(paths, SkillPath{Name: skill.Name, Path: path})
	}
	return paths, nil
}
