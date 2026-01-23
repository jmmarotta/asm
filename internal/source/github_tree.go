package source

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type GitHubTreeSpec struct {
	Origin string
	Ref    string
	Subdir string
}

func ParseGitHubTreeURL(raw string, refs map[string]struct{}) (GitHubTreeSpec, bool, error) {
	origin, treeSegments, ok, err := parseGitHubTreeSegments(raw)
	if err != nil || !ok {
		return GitHubTreeSpec{}, ok, err
	}

	if len(treeSegments) == 0 {
		return GitHubTreeSpec{}, true, fmt.Errorf("github tree url missing ref")
	}

	ref, subdir, err := resolveTreeRef(treeSegments, refs)
	if err != nil {
		return GitHubTreeSpec{}, true, err
	}

	return GitHubTreeSpec{
		Origin: origin,
		Ref:    ref,
		Subdir: subdir,
	}, true, nil
}

func GitHubTreeOrigin(raw string) (string, bool, error) {
	origin, _, ok, err := parseGitHubTreeSegments(raw)
	if err != nil || !ok {
		return "", ok, err
	}

	return origin, true, nil
}

func parseGitHubTreeSegments(raw string) (string, []string, bool, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", nil, false, err
	}
	if parsed.Host != "github.com" {
		return "", nil, false, nil
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) < 4 || segments[2] != "tree" {
		return "", nil, false, nil
	}

	owner := segments[0]
	repo := strings.TrimSuffix(segments[1], ".git")
	origin := "https://github.com/" + owner + "/" + repo
	return origin, segments[3:], true, nil
}

func resolveTreeRef(segments []string, refs map[string]struct{}) (string, string, error) {
	if len(refs) == 0 {
		return "", "", fmt.Errorf("unable to resolve ref from github tree url")
	}

	for end := len(segments); end > 0; end-- {
		candidate := path.Join(segments[:end]...)
		if _, ok := refs[candidate]; ok {
			subdir := ""
			if end < len(segments) {
				subdir = path.Join(segments[end:]...)
			}
			return candidate, subdir, nil
		}
	}

	return "", "", fmt.Errorf("unable to resolve ref from github tree url")
}
