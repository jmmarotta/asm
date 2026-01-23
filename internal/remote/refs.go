package remote

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

type RefIndex struct {
	All      map[string]struct{}
	Branches map[string]struct{}
	Tags     map[string]struct{}
}

func ListRemoteRefs(origin string) (RefIndex, error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{origin},
	})

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return RefIndex{}, err
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
