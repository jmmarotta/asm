package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

func EnsureRepo(path string, origin string, ref string) error {
	if _, err := os.Stat(path); err == nil {
		return UpdateRepo(path, origin, ref)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	options := &git.CloneOptions{
		URL:          origin,
		Depth:        1,
		SingleBranch: true,
		Tags:         git.AllTags,
	}

	reference, err := resolveReference(origin, ref)
	if err != nil {
		return err
	}
	if reference != "" {
		options.ReferenceName = reference
	}

	repo, err := git.PlainClone(path, false, options)
	if err != nil {
		return err
	}

	return checkoutRef(repo, ref, reference)
}

func UpdateRepo(path string, origin string, ref string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		Tags:       git.AllTags,
		RefSpecs: []config.RefSpec{
			"+refs/heads/*:refs/remotes/origin/*",
			"+refs/tags/*:refs/tags/*",
		},
	}

	if err := repo.Fetch(fetchOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	reference, err := resolveReference(origin, ref)
	if err != nil {
		return err
	}
	if reference.IsBranch() {
		reference = plumbing.NewRemoteReferenceName("origin", reference.Short())
	}

	return checkoutRef(repo, ref, reference)
}

func checkoutRef(repo *git.Repository, ref string, reference plumbing.ReferenceName) error {
	if ref == "" {
		return nil
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	options := &git.CheckoutOptions{Force: true}
	if reference != "" {
		options.Branch = reference
		return worktree.Checkout(options)
	}

	if len(ref) == 40 && isHex(ref) {
		options.Hash = plumbing.NewHash(ref)
		return worktree.Checkout(options)
	}

	return fmt.Errorf("ref %q not found", ref)
}

func resolveReference(origin string, ref string) (plumbing.ReferenceName, error) {
	if ref == "" {
		return "", nil
	}

	refs, err := ListRemoteRefs(origin)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(ref, "refs/") {
		return plumbing.ReferenceName(ref), nil
	}
	if strings.HasPrefix(ref, "tags/") {
		return plumbing.NewTagReferenceName(strings.TrimPrefix(ref, "tags/")), nil
	}

	if _, ok := refs.Tags[ref]; ok {
		return plumbing.NewTagReferenceName(ref), nil
	}
	if _, ok := refs.Branches[ref]; ok {
		return plumbing.NewBranchReferenceName(ref), nil
	}

	return "", nil
}

func isHex(value string) bool {
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
