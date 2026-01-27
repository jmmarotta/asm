package linker

import "github.com/jmmarotta/agent_skills_manager/internal/manifest"

func SourcesFromSkillPaths(paths []manifest.SkillPath) []Source {
	sources := make([]Source, 0, len(paths))
	for _, path := range paths {
		sources = append(sources, Source{Name: path.Name, Path: path.Path})
	}
	return sources
}

func SourcesFromConfig(config manifest.Config, originPaths map[string]string) ([]Source, error) {
	skillPaths, err := config.ResolveSkillPaths(originPaths)
	if err != nil {
		return nil, err
	}
	return SourcesFromSkillPaths(skillPaths), nil
}
