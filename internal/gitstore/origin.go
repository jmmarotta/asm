package gitstore

import (
	"fmt"
	"os"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

type OriginResolution struct {
	Path         string
	Rev          string
	UsingReplace bool
	LockChanged  bool
	Warning      string
}

type OriginPathsResult struct {
	Paths       map[string]string
	Warnings    []string
	LockChanged bool
}

func ResolveOriginRevision(storeDir string, origin string, version string, replacePath string, lock map[manifest.LockKey]string, strict bool) (OriginResolution, error) {
	if replacePath != "" {
		if info, err := os.Stat(replacePath); err == nil && info.IsDir() {
			rev, changed, err := ResolveRevision(replacePath, origin, version, lock, strict)
			if err == nil {
				return OriginResolution{
					Path:         replacePath,
					Rev:          rev,
					UsingReplace: true,
					LockChanged:  changed,
				}, nil
			}
			warning := fmt.Sprintf("replace path for %s not usable (%v); falling back to remote", origin, err)
			return resolveOriginFromStore(storeDir, origin, version, lock, strict, warning)
		}
		warning := fmt.Sprintf("replace path missing for %s (%s); falling back to remote", origin, replacePath)
		return resolveOriginFromStore(storeDir, origin, version, lock, strict, warning)
	}

	return resolveOriginFromStore(storeDir, origin, version, lock, strict, "")
}

func ResolveOrigins(storeDir string, origins map[string]string, replace map[string]string, lock map[manifest.LockKey]string, strict bool) (OriginPathsResult, error) {
	result := OriginPathsResult{Paths: map[string]string{}}
	for origin, version := range origins {
		debug.Logf("resolve origin origin=%s version=%s", debug.SanitizeOrigin(origin), version)
		replacePath := ""
		if replace != nil {
			replacePath = replace[origin]
		}
		resolution, err := ResolveOriginRevision(storeDir, origin, version, replacePath, lock, strict)
		if err != nil {
			return result, fmt.Errorf("resolve origin %s: %w", debug.SanitizeOrigin(origin), err)
		}
		if resolution.Warning != "" {
			result.Warnings = append(result.Warnings, resolution.Warning)
		}
		if resolution.LockChanged {
			result.LockChanged = true
		}
		result.Paths[origin] = resolution.Path
		applyWarning, err := ApplyOriginResolution(resolution)
		if err != nil {
			return result, err
		}
		if applyWarning != "" {
			result.Warnings = append(result.Warnings, applyWarning)
		}
	}
	return result, nil
}

func resolveOriginFromStore(storeDir string, origin string, version string, lock map[manifest.LockKey]string, strict bool, warning string) (OriginResolution, error) {
	path := RepoPath(storeDir, origin)
	if err := EnsureRepo(path, origin); err != nil {
		return OriginResolution{}, err
	}

	rev, changed, err := ResolveRevision(path, origin, version, lock, strict)
	if err != nil {
		return OriginResolution{}, err
	}

	return OriginResolution{
		Path:        path,
		Rev:         rev,
		LockChanged: changed,
		Warning:     warning,
	}, nil
}

func ApplyOriginResolution(resolution OriginResolution) (string, error) {
	if resolution.Path == "" {
		return "", fmt.Errorf("resolved path is empty")
	}
	if resolution.UsingReplace {
		head, err := HeadHash(resolution.Path)
		if err != nil {
			return "", err
		}
		if head != resolution.Rev {
			return fmt.Sprintf("replace repo %s is at %s, expected %s", resolution.Path, head, resolution.Rev), nil
		}
		return "", nil
	}

	return "", CheckoutRevision(resolution.Path, resolution.Rev)
}
