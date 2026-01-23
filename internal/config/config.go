package config

import "fmt"

type Config struct {
	Sources []Source `json:"sources"`
	Targets []Target `json:"targets"`
}

type Source struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Origin string `json:"origin"`
	Subdir string `json:"subdir,omitempty"`
	Ref    string `json:"ref,omitempty"`
}

type Target struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (config *Config) Validate() error {
	for index, source := range config.Sources {
		if err := source.Validate(index); err != nil {
			return err
		}
	}

	for index, target := range config.Targets {
		if err := target.Validate(index); err != nil {
			return err
		}
	}

	return nil
}

func (config *Config) UpsertSource(source Source) {
	for index, existing := range config.Sources {
		if existing.Name == source.Name {
			config.Sources[index] = source
			return
		}
	}
	config.Sources = append(config.Sources, source)
}

func (config *Config) RemoveSource(name string) (Source, bool) {
	for index, source := range config.Sources {
		if source.Name == name {
			config.Sources = append(config.Sources[:index], config.Sources[index+1:]...)
			return source, true
		}
	}
	return Source{}, false
}

func (config *Config) UpsertTarget(target Target) {
	for index, existing := range config.Targets {
		if existing.Name == target.Name {
			config.Targets[index] = target
			return
		}
	}
	config.Targets = append(config.Targets, target)
}

func (config *Config) RemoveTarget(name string) (Target, bool) {
	for index, target := range config.Targets {
		if target.Name == name {
			config.Targets = append(config.Targets[:index], config.Targets[index+1:]...)
			return target, true
		}
	}
	return Target{}, false
}

func (source Source) Validate(index int) error {
	if source.Name == "" {
		return fmt.Errorf("source[%d]: missing name", index)
	}
	if source.Type == "" {
		return fmt.Errorf("source[%d]: missing type", index)
	}
	if source.Origin == "" {
		return fmt.Errorf("source[%d]: missing origin", index)
	}
	return nil
}

func (target Target) Validate(index int) error {
	if target.Name == "" {
		return fmt.Errorf("target[%d]: missing name", index)
	}
	if target.Path == "" {
		return fmt.Errorf("target[%d]: missing path", index)
	}
	return nil
}
