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

func ResolveRevision(repoPath string, origin string, version string, sum map[manifest.SumKey]string, strict bool) (string, bool, error) {
	debug.Logf("resolve revision repo=%s origin=%s version=%s strict=%t", repoPath, debug.SanitizeOrigin(origin), version, strict)
	if sum == nil {
		sum = map[manifest.SumKey]string{}
	}

	repo, err := openRepo(repoPath)
	if err != nil {
		return "", false, err
	}

	key := manifest.SumKey{Origin: origin, Version: version}
	if semver.IsValid(version) {
		rev, err := ResolveForVersion(repo, version)
		if err != nil {
			return "", false, err
		}
		if existing, ok := sum[key]; ok && existing != rev {
			if strict {
				return "", false, fmt.Errorf("version %s moved for %s; run asm update", version, origin)
			}
		}
		if sum[key] != rev {
			sum[key] = rev
			return rev, true, nil
		}
		return rev, false, nil
	}

	if !module.IsPseudoVersion(version) {
		return "", false, fmt.Errorf("invalid version %q", version)
	}

	if existing, ok := sum[key]; ok {
		revPrefix := pseudoVersionRev(version)
		if revPrefix == "" || !strings.HasPrefix(existing, revPrefix) {
			return "", false, fmt.Errorf("skills.sum entry for %s %s does not match version", origin, version)
		}
		if _, err := repo.CommitObject(plumbing.NewHash(existing)); err == nil {
			return existing, false, nil
		}
	}

	rev, err := ResolveForVersion(repo, version)
	if err != nil {
		return "", false, err
	}
	if sum[key] != rev {
		sum[key] = rev
		return rev, true, nil
	}
	return rev, false, nil
}
