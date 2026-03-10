package asm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/gitstore"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

const defaultUpdateResolveParallelism = 4

type latestOriginResult struct {
	Resolved   gitstore.Resolved
	Prefetched gitstore.PrefetchedOrigin
}

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

	resolvedOrigins, err := resolveLatestOrigins(state, updatedOrigins, origins, explicit)
	if err != nil {
		return UpdateReport{}, err
	}

	prefetched := map[string]gitstore.PrefetchedOrigin{}
	for _, origin := range updatedOrigins {
		versionValue := origins[origin]
		if !explicit && semver.IsValid(versionValue) && !module.IsPseudoVersion(versionValue) {
			continue
		}

		resolved := resolvedOrigins[origin]
		debug.Logf(
			"update origin=%s from=%s to=%s rev=%s",
			debug.SanitizeOrigin(origin),
			versionValue,
			resolved.Resolved.Version,
			resolved.Resolved.Rev,
		)

		updateOriginVersion(state.Config.Skills, origin, resolved.Resolved.Version)
		deleteLockForOrigin(state.Lock, origin)
		state.Lock[manifest.LockKey{Origin: origin, Version: resolved.Resolved.Version}] = resolved.Resolved.Rev
		prefetched[origin] = resolved.Prefetched
	}

	if err := manifest.SaveState(state); err != nil {
		return UpdateReport{}, fmt.Errorf("save manifest: %w", err)
	}

	report, err := installSkills(state, prefetched)
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

	origin, err := normalizeOriginSelector(selector)
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
		if sameOrigin(skill.Origin, origin) && skill.Subdir == subdir {
			return skill, true
		}
	}
	return manifest.Skill{}, false
}

func findSkillByOrigin(skills []manifest.Skill, origin string) (manifest.Skill, bool) {
	for _, skill := range skills {
		if !sameOrigin(skill.Origin, origin) {
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

func resolveLatestOrigins(state manifest.State, orderedOrigins []string, versions map[string]string, explicit bool) (map[string]latestOriginResult, error) {
	results := map[string]latestOriginResult{}
	pending := make([]string, 0, len(orderedOrigins))
	for _, origin := range orderedOrigins {
		version := versions[origin]
		if !explicit && semver.IsValid(version) && !module.IsPseudoVersion(version) {
			continue
		}
		pending = append(pending, origin)
	}
	if len(pending) == 0 {
		return results, nil
	}

	type item struct {
		origin string
		result latestOriginResult
		err    error
	}

	maxParallel := defaultUpdateResolveParallelism
	if maxParallel > len(pending) {
		maxParallel = len(pending)
	}
	jobs := make(chan string)
	items := make(chan item, len(pending))

	var wg sync.WaitGroup
	for range maxParallel {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for origin := range jobs {
				resolved, prefetched, err := resolveLatestOriginPrefetched(state, origin)
				items <- item{origin: origin, result: latestOriginResult{Resolved: resolved, Prefetched: prefetched}, err: err}
			}
		}()
	}

	for _, origin := range pending {
		jobs <- origin
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(items)
	}()

	collected := map[string]item{}
	for item := range items {
		collected[item.origin] = item
	}
	for _, origin := range pending {
		item := collected[origin]
		if item.err != nil {
			return nil, item.err
		}
		results[origin] = item.result
	}

	return results, nil
}

func resolveLatestOriginPrefetched(state manifest.State, origin string) (gitstore.Resolved, gitstore.PrefetchedOrigin, error) {
	replacePath := ""
	if state.Config.Replace != nil {
		replacePath = state.Config.Replace[origin]
	}
	if replacePath != "" {
		if info, err := os.Stat(replacePath); err == nil && info.IsDir() {
			resolved, err := gitstore.ResolveForRefAt(replacePath, "")
			if err == nil {
				return resolved, gitstore.PrefetchedOrigin{Path: replacePath, Rev: resolved.Rev, UsingReplace: true}, nil
			}
			debug.Logf("update replace fallback origin=%s err=%v", debug.SanitizeOrigin(origin), err)
		}
	}

	path := gitstore.RepoPath(state.Paths.StoreDir, origin)
	if err := gitstore.EnsureRepo(path, origin); err != nil {
		return gitstore.Resolved{}, gitstore.PrefetchedOrigin{}, err
	}

	resolved, err := resolveRemoteRef(path, origin, "")
	if err != nil {
		return gitstore.Resolved{}, gitstore.PrefetchedOrigin{}, fmt.Errorf("resolve latest for %s: %w", debug.SanitizeOrigin(origin), err)
	}

	return resolved, gitstore.PrefetchedOrigin{Path: path, Rev: resolved.Rev}, nil
}
