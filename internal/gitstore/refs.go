package gitstore

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

type RefIndex struct {
	All      map[string]struct{}
	Branches map[string]struct{}
	Tags     map[string]struct{}
}

func ListRemoteRefs(origin string) (RefIndex, error) {
	debug.Logf("list remote refs origin=%s", debug.SanitizeOrigin(origin))
	access, err := ResolveRemoteAccess(origin)
	if err != nil {
		return RefIndex{}, err
	}
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{access.URL},
	})

	refs, err := remote.List(&git.ListOptions{Auth: access.Auth})
	if err != nil {
		return RefIndex{}, fmt.Errorf("list remote refs for %s: %w", debug.SanitizeOrigin(origin), err)
	}

	index := RefIndex{
		All:      make(map[string]struct{}),
		Branches: make(map[string]struct{}),
		Tags:     make(map[string]struct{}),
	}

	for _, ref := range refs {
		switch {
		case ref.Name().IsBranch():
			name := ref.Name().Short()
			index.All[name] = struct{}{}
			index.Branches[name] = struct{}{}
		case ref.Name().IsTag():
			name := ref.Name().Short()
			index.All[name] = struct{}{}
			index.Tags[name] = struct{}{}
		}
	}

	return index, nil
}

func RemoteHeadHash(origin string) (string, error) {
	debug.Logf("list remote head origin=%s", debug.SanitizeOrigin(origin))
	access, err := ResolveRemoteAccess(origin)
	if err != nil {
		return "", err
	}
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{access.URL},
	})

	refs, err := remote.List(&git.ListOptions{Auth: access.Auth})
	if err != nil {
		return "", fmt.Errorf("list remote refs for %s: %w", debug.SanitizeOrigin(origin), err)
	}

	var headRef *plumbing.Reference
	for _, ref := range refs {
		if ref.Name() == plumbing.HEAD {
			headRef = ref
			break
		}
	}
	if headRef == nil {
		return "", fmt.Errorf("remote head not found for %s", debug.SanitizeOrigin(origin))
	}

	if headRef.Type() == plumbing.HashReference && !headRef.Hash().IsZero() {
		return headRef.Hash().String(), nil
	}
	if headRef.Type() == plumbing.SymbolicReference {
		target := headRef.Target()
		for _, ref := range refs {
			if ref.Name() == target && !ref.Hash().IsZero() {
				return ref.Hash().String(), nil
			}
		}
	}

	if !headRef.Hash().IsZero() {
		return headRef.Hash().String(), nil
	}

	return "", fmt.Errorf("remote head not found for %s", debug.SanitizeOrigin(origin))
}
