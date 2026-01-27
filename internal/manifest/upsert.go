package manifest

import "fmt"

type DiscoveredSkill struct {
	Name   string
	Subdir string
}

type UpsertOptions struct {
	Origin  string
	Type    string
	Version string
	Author  string
}

func (config *Config) UpsertDiscoveredSkills(skills []DiscoveredSkill, opts UpsertOptions) error {
	if opts.Origin == "" {
		return fmt.Errorf("origin is required")
	}
	if opts.Type == "" {
		return fmt.Errorf("type is required")
	}
	if opts.Author == "" {
		opts.Author = "unknown"
	}

	existingByName := make(map[string]skillIdentity)
	existingByIdentity := make(map[skillIdentity]string)
	for index, skill := range config.Skills {
		normalizedSubdir, err := normalizeSubdir(skill.Subdir)
		if err != nil {
			return fmt.Errorf("skills[%d]: invalid subdir %q: %w", index, skill.Subdir, err)
		}
		identity := skillIdentity{origin: skill.Origin, subdir: normalizedSubdir}
		if prior, ok := existingByIdentity[identity]; ok && prior != skill.Name {
			return fmt.Errorf("duplicate skills for origin %q subdir %q: %q and %q", skill.Origin, normalizedSubdir, prior, skill.Name)
		}
		existingByIdentity[identity] = skill.Name
		existingByName[skill.Name] = identity
	}

	if opts.Type == "git" {
		for index, skill := range config.Skills {
			if skill.Type == "git" && skill.Origin == opts.Origin {
				skill.Version = opts.Version
				config.Skills[index] = skill
			}
		}
	}

	for _, skill := range skills {
		normalizedSubdir, err := normalizeSubdir(skill.Subdir)
		if err != nil {
			return fmt.Errorf("invalid subdir %q: %w", skill.Subdir, err)
		}
		identity := skillIdentity{origin: opts.Origin, subdir: normalizedSubdir}
		name := skill.Name
		if existingName, ok := existingByIdentity[identity]; ok {
			name = existingName
		} else if existingIdentity, ok := existingByName[name]; ok && existingIdentity != identity {
			name = fmt.Sprintf("%s/%s", opts.Author, name)
			if collisionIdentity, collision := existingByName[name]; collision && collisionIdentity != identity {
				return fmt.Errorf("name collision for %q", skill.Name)
			}
		}

		entry := Skill{
			Name:   name,
			Type:   opts.Type,
			Origin: opts.Origin,
			Subdir: normalizedSubdir,
		}
		if opts.Type == "git" {
			entry.Version = opts.Version
		}
		config.UpsertSkill(entry)
		existingByIdentity[identity] = name
		existingByName[name] = identity
	}

	return nil
}
