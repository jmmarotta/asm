package source

import (
	"os"
	"path/filepath"
)

func findRepoRoot(start string) (string, bool, error) {
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

func isGitDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
