package gitutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

func FindRepoRoot(start string) (string, bool, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", false, err
		}
		start = cwd
	}

	current, err := filepath.Abs(start)
	if err != nil {
		return "", false, err
	}

	for {
		if isGitDir(current) {
			return current, true, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

func HeadSHA(repoRoot string) (string, error) {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return "", fmt.Errorf("open repo %s: %w", repoRoot, err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	return head.Hash().String(), nil
}

func OriginURL(repoRoot string) (string, bool, error) {
	debug.Logf("read origin url repo=%s", repoRoot)
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return "", false, fmt.Errorf("open repo %s: %w", repoRoot, err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		if err == git.ErrRemoteNotFound {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read origin remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", false, nil
	}

	return urls[0], true, nil
}

func isGitDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
