package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	output, err := exec.Command("git", "-C", repoRoot, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func isGitDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
