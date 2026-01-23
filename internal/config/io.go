package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/jsonc"
)

const (
	jsoncFilename = "config.jsonc"
	jsonFilename  = "config.json"
)

func Load(configDir string) (Config, string, error) {
	path, exists, err := resolveConfigPath(configDir)
	if err != nil {
		return Config{}, path, err
	}

	if !exists {
		return Config{}, path, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, path, err
	}

	cleaned := jsonc.ToJSON(data)
	var parsed Config
	if err := json.Unmarshal(cleaned, &parsed); err != nil {
		return Config{}, path, err
	}

	expanded, err := expandConfigOrigins(parsed)
	if err != nil {
		return Config{}, path, err
	}

	if err := expanded.Validate(); err != nil {
		return Config{}, path, err
	}

	return expanded, path, nil
}

func Save(path string, config Config) error {
	normalized, err := normalizeConfigOrigins(config)
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

func resolveConfigPath(configDir string) (string, bool, error) {
	jsoncPath := filepath.Join(configDir, jsoncFilename)
	if exists, err := fileExists(jsoncPath); err != nil {
		return jsoncPath, false, err
	} else if exists {
		return jsoncPath, true, nil
	}

	jsonPath := filepath.Join(configDir, jsonFilename)
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

func expandConfigOrigins(config Config) (Config, error) {
	expanded := Config{
		Sources: make([]Source, len(config.Sources)),
		Targets: make([]Target, len(config.Targets)),
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	for index, source := range config.Sources {
		if source.Type == "path" {
			source.Origin = expandHomePath(source.Origin, home)
		}
		expanded.Sources[index] = source
	}

	for index, target := range config.Targets {
		target.Path = expandHomePath(target.Path, home)
		expanded.Targets[index] = target
	}

	return expanded, nil
}

func normalizeConfigOrigins(config Config) (Config, error) {
	normalized := Config{
		Sources: make([]Source, len(config.Sources)),
		Targets: make([]Target, len(config.Targets)),
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	for index, source := range config.Sources {
		if source.Type == "path" {
			source.Origin = collapseHomePath(source.Origin, home)
		}
		normalized.Sources[index] = source
	}

	for index, target := range config.Targets {
		target.Path = collapseHomePath(target.Path, home)
		normalized.Targets[index] = target
	}

	return normalized, nil
}

func expandHomePath(value string, home string) string {
	if value == "$HOME" {
		return home
	}
	if strings.HasPrefix(value, "$HOME/") {
		return filepath.Join(home, strings.TrimPrefix(value, "$HOME/"))
	}
	return value
}

func collapseHomePath(value string, home string) string {
	value = filepath.Clean(value)
	home = filepath.Clean(home)
	if value == home {
		return "$HOME"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(value, prefix) {
		return filepath.Join("$HOME", strings.TrimPrefix(value, prefix))
	}
	return value
}
