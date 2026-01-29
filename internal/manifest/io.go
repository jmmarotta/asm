package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tidwall/jsonc"

	"github.com/jmmarotta/agent_skills_manager/internal/source"
)

const (
	jsoncFilename = "skills.jsonc"
	jsonFilename  = "skills.json"
	lockFilename  = "skills-lock.json"

	lockSchemaVersion = 1
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

func LockPath(root string) string {
	return filepath.Join(root, lockFilename)
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

type lockFile struct {
	Schema  int         `json:"schema"`
	Entries []LockEntry `json:"entries"`
}

type LockEntry struct {
	Origin  string `json:"origin"`
	Version string `json:"version"`
	Rev     string `json:"rev"`
}

func LoadLock(path string) (map[LockKey]string, error) {
	entries := make(map[LockKey]string)
	if path == "" {
		return entries, fmt.Errorf("lock path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}

	var parsed lockFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Schema != 0 && parsed.Schema != lockSchemaVersion {
		return nil, fmt.Errorf("unsupported lock schema %d", parsed.Schema)
	}

	for _, entry := range parsed.Entries {
		if entry.Origin == "" || entry.Version == "" || entry.Rev == "" {
			return nil, fmt.Errorf("invalid skills-lock.json entry: origin=%q version=%q rev=%q", entry.Origin, entry.Version, entry.Rev)
		}
		key := LockKey{Origin: entry.Origin, Version: entry.Version}
		if existing, ok := entries[key]; ok && existing != entry.Rev {
			return nil, fmt.Errorf("skills-lock.json has conflicting entries for %s %s", key.Origin, key.Version)
		}
		entries[key] = entry.Rev
	}

	return entries, nil
}

func SaveLock(path string, entries map[LockKey]string) error {
	if path == "" {
		return fmt.Errorf("lock path is required")
	}

	lockEntries := make([]LockEntry, 0, len(entries))
	for key, rev := range entries {
		lockEntries = append(lockEntries, LockEntry{Origin: key.Origin, Version: key.Version, Rev: rev})
	}

	sort.Slice(lockEntries, func(i, j int) bool {
		left := lockEntries[i]
		right := lockEntries[j]
		if left.Origin != right.Origin {
			return left.Origin < right.Origin
		}
		if left.Version != right.Version {
			return left.Version < right.Version
		}
		return left.Rev < right.Rev
	})

	payload, err := json.MarshalIndent(lockFile{Schema: lockSchemaVersion, Entries: lockEntries}, "", "  ")
	if err != nil {
		return err
	}
	if len(payload) == 0 || payload[len(payload)-1] != '\n' {
		payload = append(payload, '\n')
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
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
		originValue, _, err := source.NormalizeFileOrigin(skill.Origin)
		if err != nil {
			return Config{}, fmt.Errorf("skills[%d]: %w", index, err)
		}
		skill.Origin = originValue
		if err := source.ValidateOriginScheme(skill.Origin); err != nil {
			return Config{}, fmt.Errorf("skills[%d]: %w", index, err)
		}
		if !source.IsRemoteOrigin(skill.Origin) {
			skill.Origin = expandRelativePath(skill.Origin, root)
		}
		expanded.Skills[index] = skill
	}

	for origin, replace := range config.Replace {
		replaceValue, _, err := source.NormalizeFileOrigin(replace)
		if err != nil {
			return Config{}, fmt.Errorf("replace[%q]: %w", origin, err)
		}
		if err := source.ValidateOriginScheme(replaceValue); err != nil {
			return Config{}, fmt.Errorf("replace[%q]: %w", origin, err)
		}
		if source.IsRemoteOrigin(replaceValue) {
			return Config{}, fmt.Errorf("replace[%q]: replace path must be a local path", origin)
		}
		expanded.Replace[origin] = expandRelativePath(replaceValue, root)
	}

	return expanded, nil
}

func normalizeConfigPaths(config Config, root string) (Config, error) {
	normalized := Config{
		Skills:  make([]Skill, len(config.Skills)),
		Replace: make(map[string]string, len(config.Replace)),
	}

	for index, skill := range config.Skills {
		originValue, _, err := source.NormalizeFileOrigin(skill.Origin)
		if err != nil {
			return Config{}, fmt.Errorf("skills[%d]: %w", index, err)
		}
		skill.Origin = originValue
		if err := source.ValidateOriginScheme(skill.Origin); err != nil {
			return Config{}, fmt.Errorf("skills[%d]: %w", index, err)
		}
		if !source.IsRemoteOrigin(skill.Origin) {
			skill.Origin = collapseRelativePath(skill.Origin, root)
		}
		normalized.Skills[index] = skill
	}

	for origin, replace := range config.Replace {
		replaceValue, _, err := source.NormalizeFileOrigin(replace)
		if err != nil {
			return Config{}, fmt.Errorf("replace[%q]: %w", origin, err)
		}
		if err := source.ValidateOriginScheme(replaceValue); err != nil {
			return Config{}, fmt.Errorf("replace[%q]: %w", origin, err)
		}
		if source.IsRemoteOrigin(replaceValue) {
			return Config{}, fmt.Errorf("replace[%q]: replace path must be a local path", origin)
		}
		normalized.Replace[origin] = collapseRelativePath(replaceValue, root)
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
