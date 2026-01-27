package version

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

const shortHashLength = 12

type Resolved struct {
	Version string
	Rev     string
}

func ResolveForRef(repo *git.Repository, ref string) (Resolved, error) {
	debug.Logf("resolve ref=%q", ref)
	if ref == "" {
		commit, err := headCommit(repo)
		if err != nil {
			return Resolved{}, err
		}
		return resolveFromCommit(repo, commit)
	}

	if module.IsPseudoVersion(ref) {
		rev, err := ResolveForVersion(repo, ref)
		if err != nil {
			return Resolved{}, err
		}
		return Resolved{Version: ref, Rev: rev}, nil
	}

	if semver.IsValid(ref) {
		commit, err := tagCommit(repo, ref)
		if err != nil {
			if errors.Is(err, plumbing.ErrReferenceNotFound) {
				return Resolved{}, fmt.Errorf("tag %q not found", ref)
			}
			return Resolved{}, fmt.Errorf("resolve tag %q: %w", ref, err)
		}
		return Resolved{Version: ref, Rev: commit.Hash.String()}, nil
	}

	commit, err := commitForRef(repo, ref)
	if err != nil {
		return Resolved{}, err
	}

	return resolveFromCommit(repo, commit)
}

func ResolveForVersion(repo *git.Repository, version string) (string, error) {
	debug.Logf("resolve version=%q", version)
	if module.IsPseudoVersion(version) {
		rev := pseudoVersionRev(version)
		if rev == "" {
			return "", fmt.Errorf("version %q is not a valid pseudo-version", version)
		}
		commit, err := commitForPrefix(repo, rev)
		if err != nil {
			return "", err
		}
		return commit.Hash.String(), nil
	}
	if semver.IsValid(version) {
		commit, err := tagCommit(repo, version)
		if err != nil {
			if errors.Is(err, plumbing.ErrReferenceNotFound) {
				return "", fmt.Errorf("tag %q not found", version)
			}
			return "", fmt.Errorf("resolve tag %q: %w", version, err)
		}
		return commit.Hash.String(), nil
	}
	return "", fmt.Errorf("version %q is not valid", version)
}

func resolveFromCommit(repo *git.Repository, commit *object.Commit) (Resolved, error) {
	if commit == nil {
		return Resolved{}, fmt.Errorf("missing commit")
	}

	debug.Logf("resolve commit=%s", commit.Hash)

	tagsByCommit, err := semverTagsByCommit(repo)
	if err != nil {
		return Resolved{}, fmt.Errorf("list tags: %w", err)
	}
	if tags, ok := tagsByCommit[commit.Hash]; ok {
		return Resolved{Version: maxSemver(tags), Rev: commit.Hash.String()}, nil
	}

	base, err := baseVersion(repo, commit, tagsByCommit)
	if err != nil {
		return Resolved{}, fmt.Errorf("base version: %w", err)
	}
	major := "v0"
	if base != "" {
		major = semver.Major(base)
	}
	rev := commit.Hash.String()
	short := rev
	if len(short) > shortHashLength {
		short = short[:shortHashLength]
	}

	version := module.PseudoVersion(major, base, commit.Committer.When.UTC(), short)
	return Resolved{Version: version, Rev: rev}, nil
}

func headCommit(repo *git.Repository) (*object.Commit, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("read head: %w", err)
	}
	debug.Logf("head hash=%s", ref.Hash())
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("load head commit %s: %w", ref.Hash(), err)
	}
	return commit, nil
}

func commitForRef(repo *git.Repository, ref string) (*object.Commit, error) {
	if strings.HasPrefix(ref, "refs/") {
		reference, err := repo.Reference(plumbing.ReferenceName(ref), true)
		if err != nil {
			return nil, fmt.Errorf("read reference %q: %w", ref, err)
		}
		commit, err := repo.CommitObject(reference.Hash())
		if err != nil {
			return nil, fmt.Errorf("load commit for %s: %w", reference.Name(), err)
		}
		return commit, nil
	}

	if len(ref) >= shortHashLength && isHex(ref) {
		return commitForPrefix(repo, ref)
	}

	branch := plumbing.NewBranchReferenceName(ref)
	if reference, err := repo.Reference(branch, true); err == nil {
		commit, err := repo.CommitObject(reference.Hash())
		if err != nil {
			return nil, fmt.Errorf("load commit for branch %s: %w", branch, err)
		}
		return commit, nil
	}

	remoteBranch := plumbing.NewRemoteReferenceName("origin", ref)
	if reference, err := repo.Reference(remoteBranch, true); err == nil {
		commit, err := repo.CommitObject(reference.Hash())
		if err != nil {
			return nil, fmt.Errorf("load commit for remote branch %s: %w", remoteBranch, err)
		}
		return commit, nil
	}

	return nil, fmt.Errorf("ref %q not found", ref)
}

func tagCommit(repo *git.Repository, tag string) (*object.Commit, error) {
	ref, err := repo.Reference(plumbing.NewTagReferenceName(tag), true)
	if err != nil {
		return nil, fmt.Errorf("read tag %q: %w", tag, err)
	}

	hash := ref.Hash()
	if tagObject, err := repo.TagObject(hash); err == nil {
		hash = tagObject.Target
	}

	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("load tag commit %s: %w", tag, err)
	}
	return commit, nil
}

func commitForPrefix(repo *git.Repository, prefix string) (*object.Commit, error) {
	if len(prefix) == 40 && isHex(prefix) {
		commit, err := repo.CommitObject(plumbing.NewHash(prefix))
		if err != nil {
			return nil, fmt.Errorf("load commit %s: %w", prefix, err)
		}
		return commit, nil
	}

	iter, err := repo.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("list commits: %w", err)
	}
	defer iter.Close()

	var match *object.Commit
	// go-git iterators return io.EOF when exhausted.
	for {
		commit, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("iterate commits: %w", err)
		}
		if strings.HasPrefix(commit.Hash.String(), prefix) {
			if match != nil {
				return nil, fmt.Errorf("commit prefix %q is ambiguous", prefix)
			}
			match = commit
		}
	}

	if match == nil {
		return nil, fmt.Errorf("commit %q not found", prefix)
	}

	return match, nil
}

func semverTagsByCommit(repo *git.Repository) (map[plumbing.Hash][]string, error) {
	iter, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	tags := make(map[plumbing.Hash][]string)
	// go-git iterators return io.EOF when exhausted.
	for {
		ref, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("iterate tags: %w", err)
		}

		name := ref.Name().Short()
		if !semver.IsValid(name) {
			continue
		}
		canonical := semver.Canonical(name)
		if canonical == "" {
			continue
		}

		hash := ref.Hash()
		if tagObject, err := repo.TagObject(hash); err == nil {
			hash = tagObject.Target
		}
		tags[hash] = append(tags[hash], canonical)
	}

	return tags, nil
}

func baseVersion(repo *git.Repository, commit *object.Commit, tags map[plumbing.Hash][]string) (string, error) {
	if commit == nil {
		return "", fmt.Errorf("missing commit")
	}

	type entry struct {
		hash  plumbing.Hash
		depth int
	}

	queue := []entry{{hash: commit.Hash, depth: 0}}
	visited := map[plumbing.Hash]struct{}{commit.Hash: {}}
	bestDepth := -1
	candidates := []string{}

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]

		if bestDepth != -1 && next.depth > bestDepth {
			break
		}
		if found, ok := tags[next.hash]; ok {
			if bestDepth == -1 || next.depth < bestDepth {
				bestDepth = next.depth
				candidates = append([]string{}, found...)
			} else if next.depth == bestDepth {
				candidates = append(candidates, found...)
			}
		}

		if bestDepth != -1 && next.depth >= bestDepth {
			continue
		}

		obj, err := repo.CommitObject(next.hash)
		if err != nil {
			return "", fmt.Errorf("load commit %s: %w", next.hash, err)
		}
		for _, parent := range obj.ParentHashes {
			if _, ok := visited[parent]; ok {
				continue
			}
			visited[parent] = struct{}{}
			queue = append(queue, entry{hash: parent, depth: next.depth + 1})
		}
	}

	return maxSemver(candidates), nil
}

func maxSemver(values []string) string {
	if len(values) == 0 {
		return ""
	}

	valid := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			valid = append(valid, value)
		}
	}
	if len(valid) == 0 {
		return ""
	}

	sort.Slice(valid, func(i, j int) bool {
		return semver.Compare(valid[i], valid[j]) > 0
	})

	return valid[0]
}

func pseudoVersionRev(version string) string {
	base := strings.SplitN(version, "+", 2)[0]
	idx := strings.LastIndex(base, "-")
	if idx == -1 || idx == len(base)-1 {
		return ""
	}
	return base[idx+1:]
}

func isHex(value string) bool {
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
