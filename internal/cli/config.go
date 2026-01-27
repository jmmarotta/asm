package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/paths"
)

type manifestState struct {
	Root         string
	ManifestPath string
	SumPath      string
	Paths        paths.Paths
	Config       config.Config
	Sum          map[config.SumKey]string
}

func loadManifest() (manifestState, error) {
	path, err := config.FindManifestPath("")
	if err != nil {
		return manifestState{}, err
	}
	debug.Logf("manifest path=%s", path)

	return loadManifestAt(path)
}

func loadOrInitManifest() (manifestState, bool, error) {
	path, err := config.FindManifestPath("")
	if err == nil {
		debug.Logf("manifest path=%s", path)
		state, err := loadManifestAt(path)
		return state, false, err
	}
	if !errors.Is(err, config.ErrManifestNotFound) {
		return manifestState{}, false, err
	}

	root, err := os.Getwd()
	if err != nil {
		return manifestState{}, false, err
	}

	manifestPath := config.DefaultManifestPath(root)
	debug.Logf("manifest init path=%s", manifestPath)
	return manifestState{
		Root:         root,
		ManifestPath: manifestPath,
		SumPath:      config.SumPath(root),
		Paths:        paths.RepoPaths(root),
		Config: config.Config{
			Replace: map[string]string{},
		},
		Sum: map[config.SumKey]string{},
	}, true, nil
}

func loadManifestAt(path string) (manifestState, error) {
	configValue, err := config.Load(path)
	if err != nil {
		return manifestState{}, err
	}
	if configValue.Replace == nil {
		configValue.Replace = map[string]string{}
	}

	root := filepath.Dir(path)
	sumPath := config.SumPath(root)
	entries, err := config.LoadSum(sumPath)
	if err != nil {
		return manifestState{}, err
	}

	return manifestState{
		Root:         root,
		ManifestPath: path,
		SumPath:      sumPath,
		Paths:        paths.RepoPaths(root),
		Config:       configValue,
		Sum:          entries,
	}, nil
}

func saveManifest(state manifestState) error {
	if state.Config.Replace == nil {
		state.Config.Replace = map[string]string{}
	}
	if err := config.Save(state.ManifestPath, state.Config); err != nil {
		return err
	}
	if len(state.Sum) == 0 {
		if err := os.Remove(state.SumPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return config.SaveSum(state.SumPath, state.Sum)
}
