package source

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var scpPattern = regexp.MustCompile(`^[^@]+@[^:]+:`)
var githubTreePattern = regexp.MustCompile(`^https?://github.com/[^/]+/[^/]+/tree/`)
var allowedRemoteSchemes = map[string]bool{
	"git":   true,
	"http":  true,
	"https": true,
	"ssh":   true,
}

func schemeForOrigin(origin string) (string, bool) {
	index := strings.Index(origin, "://")
	if index <= 0 {
		return "", false
	}
	return strings.ToLower(origin[:index]), true
}

func NormalizeFileOrigin(origin string) (string, bool, error) {
	scheme, ok := schemeForOrigin(origin)
	if !ok || scheme != "file" {
		return origin, false, nil
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return "", false, fmt.Errorf("invalid file origin %q: %w", origin, err)
	}
	if parsed.Host != "" && parsed.Host != "localhost" {
		return "", false, fmt.Errorf("unsupported file origin host %q", parsed.Host)
	}
	pathValue, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", false, fmt.Errorf("invalid file origin %q: %w", origin, err)
	}
	if pathValue == "" {
		return "", false, fmt.Errorf("invalid file origin %q", origin)
	}

	return filepath.Clean(pathValue), true, nil
}

func ValidateOriginScheme(origin string) error {
	scheme, ok := schemeForOrigin(origin)
	if !ok {
		return nil
	}
	if scheme == "file" {
		return nil
	}
	if allowedRemoteSchemes[scheme] {
		return nil
	}
	return fmt.Errorf("unsupported origin scheme %q", scheme)
}

func IsRemoteOrigin(origin string) bool {
	if scheme, ok := schemeForOrigin(origin); ok {
		return allowedRemoteSchemes[scheme]
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
