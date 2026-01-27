package gitstore

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

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

func HeadHash(repoRoot string) (string, error) {
	repo, err := openRepo(repoRoot)
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("read head: %w", err)
	}

	return head.Hash().String(), nil
}

func ResolveForRefAt(repoPath string, ref string) (Resolved, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return Resolved{}, err
	}

	return ResolveForRef(repo, ref)
}

func ResolveForVersionAt(repoPath string, version string) (string, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return "", err
	}

	return ResolveForVersion(repo, version)
}

func CommitExists(repoPath string, hash string) (bool, error) {
	repo, err := openRepo(repoPath)
	if err != nil {
		return false, err
	}

	if _, err := repo.CommitObject(plumbing.NewHash(hash)); err != nil {
		return false, nil
	}

	return true, nil
}

func CheckoutRevision(repoPath string, rev string) error {
	debug.Logf("checkout repo=%s rev=%s", repoPath, rev)
	repo, err := openRepo(repoPath)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("open worktree: %w", err)
	}

	if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Hash: plumbing.NewHash(rev)}); err != nil {
		return fmt.Errorf("checkout %s: %w", rev, err)
	}

	return nil
}

func openRepo(repoPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo %s: %w", repoPath, err)
	}

	return repo, nil
}
