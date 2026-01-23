package source

import (
	"fmt"
	"os"
	"path/filepath"
)

type SkillDir struct {
	Name   string
	Subdir string
	Path   string
}

func DiscoverSkills(root string, subdir string) ([]SkillDir, error) {
	root = filepath.Clean(root)

	if subdir != "" {
		target := filepath.Join(root, subdir)
		return discoverFromTarget(root, target, subdir)
	}

	if ok, _ := isSkillDir(root); ok {
		return []SkillDir{{
			Name:   filepath.Base(root),
			Subdir: "",
			Path:   root,
		}}, nil
	}

	candidates := []string{root, filepath.Join(root, "skills"), filepath.Join(root, "plugins")}
	qualifying := make([]string, 0)
	for _, candidate := range candidates {
		if ok, err := isMultiSkillRoot(candidate); err != nil {
			return nil, err
		} else if ok {
			qualifying = append(qualifying, candidate)
		}
	}

	if len(qualifying) == 0 {
		return nil, fmt.Errorf("no skills found in %s", root)
	}
	if len(qualifying) > 1 {
		return nil, fmt.Errorf("multiple skill roots found; use --path to disambiguate")
	}

	return expandMultiSkillRoot(root, qualifying[0])
}

func discoverFromTarget(root string, target string, subdir string) ([]SkillDir, error) {
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", subdir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", subdir)
	}

	if ok, _ := isSkillDir(target); ok {
		return []SkillDir{{
			Name:   filepath.Base(target),
			Subdir: filepath.ToSlash(filepath.Clean(subdir)),
			Path:   target,
		}}, nil
	}

	if ok, err := isMultiSkillRoot(target); err != nil {
		return nil, err
	} else if ok {
		return expandMultiSkillRoot(root, target)
	}

	return nil, fmt.Errorf("no skills found at %s", subdir)
}

func expandMultiSkillRoot(root string, multiRoot string) ([]SkillDir, error) {
	entries, err := os.ReadDir(multiRoot)
	if err != nil {
		return nil, err
	}

	skills := make([]SkillDir, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(multiRoot, entry.Name())
		if ok, _ := isSkillDir(path); !ok {
			continue
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return nil, err
		}
		skills = append(skills, SkillDir{
			Name:   entry.Name(),
			Subdir: filepath.ToSlash(relative),
			Path:   path,
		})
	}

	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found in %s", multiRoot)
	}

	return skills, nil
}

func isSkillDir(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func isMultiSkillRoot(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	hasChild := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		hasChild = true
		ok, err := isSkillDir(filepath.Join(dir, entry.Name()))
		if err != nil || !ok {
			return false, err
		}
	}

	return hasChild, nil
}
