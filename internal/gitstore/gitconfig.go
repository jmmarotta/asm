package gitstore

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type urlRewrite struct {
	Base      string
	InsteadOf string
}

func applyInsteadOf(origin string) (string, bool, error) {
	rules, err := loadURLRewrites()
	if err != nil {
		return origin, false, err
	}

	var match urlRewrite
	found := false
	for _, rule := range rules {
		if rule.InsteadOf == "" {
			continue
		}
		if strings.HasPrefix(origin, rule.InsteadOf) {
			if !found || len(rule.InsteadOf) > len(match.InsteadOf) {
				match = rule
				found = true
			}
		}
	}

	if !found {
		return origin, false, nil
	}

	return match.Base + origin[len(match.InsteadOf):], true, nil
}

func loadURLRewrites() ([]urlRewrite, error) {
	files := gitConfigFiles()
	seen := make(map[string]struct{})
	rules := []urlRewrite{}
	for _, path := range files {
		parsed, err := parseGitConfigFile(path, seen)
		if err != nil {
			return nil, err
		}
		rules = append(rules, parsed...)
	}
	return rules, nil
}

func gitConfigFiles() []string {
	if custom := os.Getenv("GIT_CONFIG_GLOBAL"); custom != "" {
		return []string{custom}
	}

	paths := []string{}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		paths = append(paths, filepath.Join(home, ".gitconfig"))
	}

	xdg := os.Getenv("XDG_CONFIG_HOME")
	if xdg == "" && home != "" {
		xdg = filepath.Join(home, ".config")
	}
	if xdg != "" {
		paths = append(paths, filepath.Join(xdg, "git", "config"))
	}

	return paths
}

func parseGitConfigFile(path string, seen map[string]struct{}) ([]urlRewrite, error) {
	if path == "" {
		return nil, nil
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = path
	}
	if _, ok := seen[absolute]; ok {
		return nil, nil
	}
	seen[absolute] = struct{}{}

	file, err := os.Open(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var rules []urlRewrite
	currentSection := ""
	currentSubsection := ""
	baseDir := filepath.Dir(absolute)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if section, subsection, ok := parseSectionHeader(line); ok {
			currentSection = section
			currentSubsection = subsection
			continue
		}

		key, value, ok := parseKeyValue(line)
		if !ok {
			continue
		}

		switch currentSection {
		case "url":
			if key == "insteadof" {
				rules = append(rules, urlRewrite{Base: currentSubsection, InsteadOf: value})
			}
		case "include":
			if key == "path" && value != "" {
				includePath, err := resolveIncludePath(value, baseDir)
				if err != nil {
					return nil, err
				}
				included, err := parseGitConfigFile(includePath, seen)
				if err != nil {
					return nil, err
				}
				rules = append(rules, included...)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func parseSectionHeader(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
		return "", "", false
	}
	content := strings.TrimSpace(line[1 : len(line)-1])
	if content == "" {
		return "", "", false
	}

	section := content
	subsection := ""
	if idx := strings.IndexAny(content, " \t"); idx != -1 {
		section = strings.TrimSpace(content[:idx])
		rest := strings.TrimSpace(content[idx:])
		if rest != "" {
			if strings.HasPrefix(rest, "\"") {
				if unquoted, err := strconv.Unquote(rest); err == nil {
					subsection = unquoted
				} else {
					subsection = strings.Trim(rest, "\"")
				}
			} else {
				subsection = rest
			}
		}
	}

	return strings.ToLower(section), subsection, true
}

func parseKeyValue(line string) (string, string, bool) {
	if line == "" {
		return "", "", false
	}

	key := ""
	value := ""
	if idx := strings.Index(line, "="); idx != -1 {
		key = strings.TrimSpace(line[:idx])
		value = strings.TrimSpace(line[idx+1:])
	} else {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			return "", "", false
		}
		key = fields[0]
		if len(fields) > 1 {
			value = strings.Join(fields[1:], " ")
		}
	}

	if key == "" {
		return "", "", false
	}
	key = strings.ToLower(key)
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		} else {
			value = strings.Trim(value, "\"")
		}
	}

	return key, value, true
}

func resolveIncludePath(value string, baseDir string) (string, error) {
	pathValue := strings.TrimSpace(value)
	if pathValue == "" {
		return "", nil
	}

	if strings.HasPrefix(pathValue, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if pathValue == "~" {
			pathValue = home
		} else if strings.HasPrefix(pathValue, "~/") {
			pathValue = filepath.Join(home, pathValue[2:])
		}
	}

	if !filepath.IsAbs(pathValue) {
		pathValue = filepath.Join(baseDir, pathValue)
	}
	return pathValue, nil
}
