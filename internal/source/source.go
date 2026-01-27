package source

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var scpPattern = regexp.MustCompile(`^[^@]+@[^:]+:`)
var githubTreePattern = regexp.MustCompile(`^https?://github.com/[^/]+/[^/]+/tree/`)

func IsRemoteOrigin(origin string) bool {
	if strings.Contains(origin, "://") {
		return true
	}
	return scpPattern.MatchString(origin)
}

func ParseOriginRef(origin string) (string, string) {
	if !IsRemoteOrigin(origin) {
		return origin, ""
	}

	lastAt := strings.LastIndex(origin, "@")
	if lastAt == -1 {
		return origin, ""
	}

	lastSlash := strings.LastIndex(origin, "/")
	lastColon := strings.LastIndex(origin, ":")
	if lastAt <= maxIndex(lastSlash, lastColon) {
		return origin, ""
	}

	return origin[:lastAt], origin[lastAt+1:]
}

func NormalizeOrigin(origin string) string {
	normalized := strings.TrimSuffix(origin, ".git")
	normalized = strings.TrimSuffix(normalized, "/")

	if IsRemoteOrigin(origin) {
		return normalized
	}

	return filepath.Clean(normalized)
}

func AuthorForLocalPath(path string) string {
	repoRoot, inRepo, err := findRepoRoot(path)
	if err == nil && inRepo {
		return filepath.Base(repoRoot)
	}

	return filepath.Base(filepath.Dir(path))
}

func AuthorForRemoteOrigin(origin string) string {
	if strings.Contains(origin, "://") {
		parsed, err := url.Parse(origin)
		if err == nil {
			parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
			if len(parts) >= 1 && parts[0] != "" {
				return parts[0]
			}
		}
	}

	if scpPattern.MatchString(origin) {
		parts := strings.SplitN(origin, ":", 2)
		if len(parts) == 2 {
			path := strings.TrimPrefix(parts[1], "/")
			segments := strings.Split(path, "/")
			if len(segments) >= 1 {
				return segments[0]
			}
		}
	}

	return "unknown"
}

func IsGitHubTreeURL(raw string) bool {
	return githubTreePattern.MatchString(raw)
}

func maxIndex(values ...int) int {
	max := -1
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}
