package asm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
	"github.com/jmmarotta/agent_skills_manager/internal/source"
)

func Update(selector string, pathFlag string) (UpdateReport, error) {
	state, err := manifest.LoadState()
	if err != nil {
		return UpdateReport{}, fmt.Errorf("load manifest: %w", err)
	}
	selector = strings.TrimSpace(selector)
	pathFlag = strings.TrimSpace(pathFlag)
	debug.Logf("update start selector=%q path=%q", selector, pathFlag)

	if len(state.Config.Skills) == 0 {
		return UpdateReport{Install: InstallReport{NoSkills: true}}, nil
	}

	origins, explicit, err := resolveUpdateOrigins(state.Config, selector, pathFlag)
	if err != nil {
		return UpdateReport{}, err
	}
	updatedOrigins := make([]string, 0, len(origins))
	for origin := range origins {
		updatedOrigins = append(updatedOrigins, origin)
	}
	sort.Strings(updatedOrigins)

	if state.Lock == nil {
		state.Lock = map[manifest.LockKey]string{}
	}

	for origin, versionValue := range origins {
		if !explicit && semver.IsValid(versionValue) && !module.IsPseudoVersion(versionValue) {
			continue
		}

		resolved, err := resolveLatestOrigin(state, origin)
		if err != nil {
			return UpdateReport{}, err
		}
		debug.Logf(
			"update origin=%s from=%s to=%s rev=%s",
			debug.SanitizeOrigin(origin),
			versionValue,
			resolved.Version,
			resolved.Rev,
		)

		updateOriginVersion(state.Config.Skills, origin, resolved.Version)
		deleteLockForOrigin(state.Lock, origin)
		state.Lock[manifest.LockKey{Origin: origin, Version: resolved.Version}] = resolved.Rev
	}

	if err := manifest.SaveState(state); err != nil {
		return UpdateReport{}, fmt.Errorf("save manifest: %w", err)
	}

	report, err := installSkills(state)
	if err != nil {
		return UpdateReport{}, fmt.Errorf("install skills: %w", err)
	}

	return UpdateReport{Install: report, UpdatedOrigins: updatedOrigins}, nil
}

func resolveUpdateOrigins(configValue manifest.Config, selector string, pathFlag string) (map[string]string, bool, error) {
	origins := make(map[string]string)
	if selector == "" {
		if pathFlag != "" {
			return nil, false, fmt.Errorf("--path requires an origin selector")
		}
		for _, skill := range configValue.Skills {
			if skill.Version == "" {
				continue
			}
			if !module.IsPseudoVersion(skill.Version) {
				continue
			}
			origins[skill.Origin] = skill.Version
		}
		return origins, false, nil
	}

	if pathFlag == "" {
		skill, found := manifest.FindSkill(configValue.Skills, selector)
		if found {
			if skill.Version == "" {
				return nil, true, fmt.Errorf("skill %q does not have a version", selector)
			}
			origins[skill.Origin] = skill.Version
			return origins, true, nil
		}
	}

	origin, err := normalizeUpdateOrigin(selector)
	if err != nil {
		return nil, true, err
	}

	if pathFlag != "" {
		normalizedPath, err := normalizeUpdateSubdir(pathFlag)
		if err != nil {
			return nil, true, err
		}

		skill, found := findSkillByIdentity(configValue.Skills, origin, normalizedPath)
		if !found {
			return nil, true, fmt.Errorf("skill not found for origin %q and path %q", selector, pathFlag)
		}
		if skill.Version == "" {
			return nil, true, fmt.Errorf("skill %q does not have a version", skill.Name)
		}
		origins[skill.Origin] = skill.Version
		return origins, true, nil
	}

	skill, found := findSkillByOrigin(configValue.Skills, origin)
	if !found {
		return nil, true, fmt.Errorf("origin %q not found", selector)
	}
	origins[skill.Origin] = skill.Version
	return origins, true, nil
}

func normalizeUpdateOrigin(value string) (string, error) {
	if source.IsGitHubTreeURL(value) {
		origin, ok, err := source.GitHubTreeOrigin(value)
		if err != nil {
			return "", fmt.Errorf("parse github tree origin: %w", err)
		}
		if ok {
			value = origin
		}
	}

	origin, _, err := source.NormalizeFileOrigin(value)
	if err != nil {
		return "", err
	}
	if err := source.ValidateOriginScheme(origin); err != nil {
		return "", err
	}

	if source.IsRemoteOrigin(origin) {
		origin, _ = source.ParseOriginRef(origin)
		return source.NormalizeOrigin(origin), nil
	}

	abs, err := filepath.Abs(origin)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func normalizeUpdateSubdir(value string) (string, error) {
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("path must be relative: %s", value)
	}

	cleaned := filepath.Clean(value)
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path cannot escape repo: %s", value)
	}

	return filepath.ToSlash(cleaned), nil
}

func findSkillByIdentity(skills []manifest.Skill, origin string, subdir string) (manifest.Skill, bool) {
	for _, skill := range skills {
		if skill.Origin == origin && skill.Subdir == subdir {
			return skill, true
		}
	}
	return manifest.Skill{}, false
}

func findSkillByOrigin(skills []manifest.Skill, origin string) (manifest.Skill, bool) {
	for _, skill := range skills {
		if skill.Origin != origin {
			continue
		}
		if skill.Version == "" {
			return manifest.Skill{}, false
		}
		return skill, true
	}
	return manifest.Skill{}, false
}

func updateOriginVersion(skills []manifest.Skill, origin string, version string) {
	for index, skill := range skills {
		if skill.Origin != origin || skill.Version == "" {
			continue
		}
		skill.Version = version
		skills[index] = skill
	}
}

func resolveLatestOrigin(state manifest.State, origin string) (gitstore.Resolved, error) {
	replacePath := ""
	if state.Config.Replace != nil {
		replacePath = state.Config.Replace[origin]
	}
	if replacePath != "" {
		if info, err := os.Stat(replacePath); err == nil && info.IsDir() {
			resolved, err := gitstore.ResolveForRefAt(replacePath, "")
			if err == nil {
				return resolved, nil
			}
			debug.Logf("update replace fallback origin=%s err=%v", debug.SanitizeOrigin(origin), err)
		}
	}

	path := gitstore.RepoPath(state.Paths.StoreDir, origin)
	if err := gitstore.EnsureRepo(path, origin); err != nil {
		return gitstore.Resolved{}, err
	}

	resolved, err := resolveRemoteRef(path, origin, "")
	if err != nil {
		return gitstore.Resolved{}, fmt.Errorf("resolve latest for %s: %w", debug.SanitizeOrigin(origin), err)
	}

	return resolved, nil
}
