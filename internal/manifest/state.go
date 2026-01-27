package manifest

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
)

type State struct {
	Root         string
	ManifestPath string
	SumPath      string
	Paths        Paths
	Config       Config
	Sum          map[SumKey]string
}

func LoadState() (State, error) {
	path, err := FindManifestPath("")
	if err != nil {
		return State{}, err
	}
	debug.Logf("manifest path=%s", path)

	return LoadStateAt(path)
}

func LoadOrInitState() (State, bool, error) {
	path, err := FindManifestPath("")
	if err == nil {
		debug.Logf("manifest path=%s", path)
		state, err := LoadStateAt(path)
		return state, false, err
	}
	if !errors.Is(err, ErrManifestNotFound) {
		return State{}, false, err
	}

	root, err := os.Getwd()
	if err != nil {
		return State{}, false, err
	}

	manifestPath := DefaultManifestPath(root)
	debug.Logf("manifest init path=%s", manifestPath)
	return State{
		Root:         root,
		ManifestPath: manifestPath,
		SumPath:      SumPath(root),
		Paths:        RepoPaths(root),
		Config: Config{
			Replace: map[string]string{},
		},
		Sum: map[SumKey]string{},
	}, true, nil
}

func LoadStateAt(path string) (State, error) {
	configValue, err := Load(path)
	if err != nil {
		return State{}, err
	}
	if configValue.Replace == nil {
		configValue.Replace = map[string]string{}
	}

	root := filepath.Dir(path)
	sumPath := SumPath(root)
	entries, err := LoadSum(sumPath)
	if err != nil {
		return State{}, err
	}

	return State{
		Root:         root,
		ManifestPath: path,
		SumPath:      sumPath,
		Paths:        RepoPaths(root),
		Config:       configValue,
		Sum:          entries,
	}, nil
}

func SaveState(state State) error {
	if state.Config.Replace == nil {
		state.Config.Replace = map[string]string{}
	}
	if err := Save(state.ManifestPath, state.Config); err != nil {
		return err
	}
	if len(state.Sum) == 0 {
		if err := os.Remove(state.SumPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return SaveSum(state.SumPath, state.Sum)
}
