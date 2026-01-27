package remote

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{origin},
	})

	refs, err := remote.List(&git.ListOptions{})
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
