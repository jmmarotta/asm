package gitstore

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

func ResolveRevision(repoPath string, origin string, version string, lock map[manifest.LockKey]string, strict bool) (string, bool, error) {
	debug.Logf("resolve revision repo=%s origin=%s version=%s strict=%t", repoPath, debug.SanitizeOrigin(origin), version, strict)
	if lock == nil {
		lock = map[manifest.LockKey]string{}
	}

	repo, err := openRepo(repoPath)
	if err != nil {
		return "", false, err
	}

	key := manifest.LockKey{Origin: origin, Version: version}
	if semver.IsValid(version) {
		rev, err := ResolveForVersion(repo, version)
		if err != nil {
			return "", false, err
		}
		if existing, ok := lock[key]; ok && existing != rev {
			if strict {
				return "", false, fmt.Errorf("version %s moved for %s; run asm update", version, origin)
			}
		}
		if lock[key] != rev {
			lock[key] = rev
			return rev, true, nil
		}
		return rev, false, nil
	}

	if !module.IsPseudoVersion(version) {
		return "", false, fmt.Errorf("invalid version %q", version)
	}

	if existing, ok := lock[key]; ok {
		revPrefix := pseudoVersionRev(version)
		if revPrefix == "" || !strings.HasPrefix(existing, revPrefix) {
			return "", false, fmt.Errorf("skills-lock.json entry for %s %s does not match version", origin, version)
		}
		if _, err := repo.CommitObject(plumbing.NewHash(existing)); err == nil {
			return existing, false, nil
		}
	}

	rev, err := ResolveForVersion(repo, version)
	if err != nil {
		return "", false, err
	}
	if lock[key] != rev {
		lock[key] = rev
		return rev, true, nil
	}
	return rev, false, nil
}
