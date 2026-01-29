package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Input struct {
	Origin    string
	Ref       string
	Subdir    string
	LocalPath string
	RepoRoot  string
	IsLocal   bool
}

func ParseInput(input string, pathFlag string) (Input, error) {
	normalized, _, err := NormalizeFileOrigin(input)
	if err != nil {
		return Input{}, err
	}
	input = normalized
	if err := ValidateOriginScheme(input); err != nil {
		return Input{}, err
	}
	if IsRemoteOrigin(input) {
		return parseRemoteInput(input, pathFlag)
	}

	return parseLocalInput(input, pathFlag)
}

func parseLocalInput(input string, pathFlag string) (Input, error) {
	info, err := os.Stat(input)
	if err != nil {
		return Input{}, fmt.Errorf("local path not found: %s", input)
	}
	if !info.IsDir() {
		return Input{}, fmt.Errorf("local path is not a directory: %s", input)
	}

	absolute, err := filepath.Abs(input)
	if err != nil {
		return Input{}, err
	}

	repoRoot, inRepo, err := findRepoRoot(absolute)
	if err != nil {
		return Input{}, err
	}

	origin := absolute
	subdir := ""
	if inRepo {
		origin = repoRoot
		if relative, err := filepath.Rel(repoRoot, absolute); err == nil && relative != "." {
			subdir = relative
		}
	}

	if pathFlag != "" {
		if subdir == "" {
			subdir = pathFlag
		} else {
			subdir = filepath.Join(subdir, pathFlag)
		}
	}

	subdir, err = cleanSubdir(subdir)
	if err != nil {
		return Input{}, err
	}

	return Input{
		Origin:    origin,
		Ref:       "",
		Subdir:    subdir,
		LocalPath: absolute,
		RepoRoot:  repoRoot,
		IsLocal:   true,
	}, nil
}

func parseRemoteInput(input string, pathFlag string) (Input, error) {
	origin, ref := ParseOriginRef(input)
	origin = NormalizeOrigin(origin)

	subdir, err := cleanSubdir(pathFlag)
	if err != nil {
		return Input{}, err
	}

	return Input{
		Origin:  origin,
		Ref:     ref,
		Subdir:  subdir,
		IsLocal: false,
	}, nil
}

func cleanSubdir(subdir string) (string, error) {
	if subdir == "" {
		return "", nil
	}

	if filepath.IsAbs(subdir) {
		return "", fmt.Errorf("path must be relative: %s", subdir)
	}

	cleaned := filepath.Clean(subdir)
	if cleaned == "." {
		return "", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path cannot escape repo: %s", subdir)
	}

	return filepath.ToSlash(cleaned), nil
}
