package remote

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

func EnsureRepo(path string, origin string) error {
	if _, err := os.Stat(path); err == nil {
		debug.Logf("update repo path=%s origin=%s", path, debug.SanitizeOrigin(origin))
		return UpdateRepo(path, origin)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	options := &git.CloneOptions{
		URL:   origin,
		Tags:  git.AllTags,
		Depth: 0,
	}

	debug.Logf("clone repo path=%s origin=%s", path, debug.SanitizeOrigin(origin))
	if _, err := git.PlainClone(path, false, options); err != nil {
		return fmt.Errorf("clone %s: %w", debug.SanitizeOrigin(origin), err)
	}

	return nil
}

func UpdateRepo(path string, origin string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("open repo %s: %w", path, err)
	}

	debug.Logf("fetch repo path=%s origin=%s", path, debug.SanitizeOrigin(origin))
	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		Tags:       git.AllTags,
		RefSpecs: []config.RefSpec{
			"+refs/heads/*:refs/remotes/origin/*",
			"+refs/tags/*:refs/tags/*",
		},
	}

	if err := repo.Fetch(fetchOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("fetch %s: %w", debug.SanitizeOrigin(origin), err)
	}

	return nil
}
