package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tidwall/jsonc"
)

const (
	jsoncFilename = "skills.jsonc"
	jsonFilename  = "skills.json"
	sumFilename   = "skills.sum"
)

var ErrManifestNotFound = errors.New("skills.jsonc not found")

func FindManifestPath(start string) (string, error) {
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}

	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	for {
		if path, exists, err := resolveManifestPath(current); err != nil {
			return "", err
		} else if exists {
			return path, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", ErrManifestNotFound
}

func DefaultManifestPath(root string) string {
	return filepath.Join(root, jsoncFilename)
}

func SumPath(root string) string {
	return filepath.Join(root, sumFilename)
}

func Load(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("manifest path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cleaned := jsonc.ToJSON(data)
	var parsed Config
	if err := json.Unmarshal(cleaned, &parsed); err != nil {
		return Config{}, err
	}

	root := filepath.Dir(path)
	expanded, err := expandConfigPaths(parsed, root)
	if err != nil {
		return Config{}, err
	}

	if err := expanded.Validate(); err != nil {
		return Config{}, err
	}

	return expanded, nil
}

func Save(path string, config Config) error {
	if path == "" {
		return fmt.Errorf("manifest path is required")
	}

	root := filepath.Dir(path)
	normalized, err := normalizeConfigPaths(config, root)
	if err != nil {
		return err
	}

	if err := normalized.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func LoadSum(path string) (map[SumKey]string, error) {
	entries := make(map[SumKey]string)
	if path == "" {
		return entries, fmt.Errorf("sum path is required")
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid skills.sum entry: %q", line)
		}
		key := SumKey{Origin: fields[0], Version: fields[1]}
		if existing, ok := entries[key]; ok && existing != fields[2] {
			return nil, fmt.Errorf("skills.sum has conflicting entries for %s %s", key.Origin, key.Version)
		}
		entries[key] = fields[2]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func SaveSum(path string, entries map[SumKey]string) error {
	if path == "" {
		return fmt.Errorf("sum path is required")
	}

	keys := make([]SumKey, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		left := keys[i]
		right := keys[j]
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		return left.Version < right.Version
	})

	builder := &strings.Builder{}
	for _, key := range keys {
		fmt.Fprintf(builder, "%s %s %s\n", key.Origin, key.Version, entries[key])
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func resolveManifestPath(root string) (string, bool, error) {
	jsoncPath := filepath.Join(root, jsoncFilename)
	if exists, err := fileExists(jsoncPath); err != nil {
		return jsoncPath, false, err
	} else if exists {
		return jsoncPath, true, nil
	}

	jsonPath := filepath.Join(root, jsonFilename)
	if exists, err := fileExists(jsonPath); err != nil {
		return jsoncPath, false, err
	} else if exists {
		return jsonPath, true, nil
	}

	return jsoncPath, false, nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func expandConfigPaths(config Config, root string) (Config, error) {
	expanded := Config{
		Skills:  make([]Skill, len(config.Skills)),
		Replace: make(map[string]string, len(config.Replace)),
	}

	for index, skill := range config.Skills {
		if skill.Type == "path" {
			skill.Origin = expandRelativePath(skill.Origin, root)
		}
		expanded.Skills[index] = skill
	}

	for origin, replace := range config.Replace {
		expanded.Replace[origin] = expandRelativePath(replace, root)
	}

	return expanded, nil
}

func normalizeConfigPaths(config Config, root string) (Config, error) {
	normalized := Config{
		Skills:  make([]Skill, len(config.Skills)),
		Replace: make(map[string]string, len(config.Replace)),
	}

	for index, skill := range config.Skills {
		if skill.Type == "path" {
			skill.Origin = collapseRelativePath(skill.Origin, root)
		}
		normalized.Skills[index] = skill
	}

	for origin, replace := range config.Replace {
		normalized.Replace[origin] = collapseRelativePath(replace, root)
	}

	return normalized, nil
}

func expandRelativePath(value string, root string) string {
	if value == "" {
		return ""
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return filepath.Clean(filepath.Join(root, filepath.FromSlash(value)))
}

func collapseRelativePath(value string, root string) string {
	if value == "" {
		return ""
	}
	value = filepath.Clean(value)
	root = filepath.Clean(root)
	relative, err := filepath.Rel(root, value)
	if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(relative)
	}
	return value
}
