package cli

import (
	"fmt"
	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/paths"
	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

type configSet struct {
	Local            config.Config
	Global           config.Config
	LocalConfigPath  string
	GlobalConfigPath string
	InRepo           bool
	RepoRoot         string
}

type scopedConfig struct {
	Config     config.Config
	ConfigPath string
	Paths      paths.Paths
	Scope      scope.Scope
	RepoRoot   string
}

func loadConfigSet() (configSet, error) {
	repoRoot, inRepo, err := scope.FindRepoRoot("")
	if err != nil {
		return configSet{}, err
	}

	globalPaths, err := paths.GlobalPaths()
	if err != nil {
		return configSet{}, err
	}

	globalConfig, globalPath, err := config.Load(globalPaths.ConfigDir)
	if err != nil {
		return configSet{}, err
	}

	localConfig := config.Config{}
	localPath := ""
	if inRepo {
		localPaths := paths.LocalPaths(repoRoot)
		localConfig, localPath, err = config.Load(localPaths.ConfigDir)
		if err != nil {
			return configSet{}, err
		}
	}

	return configSet{
		Local:            localConfig,
		Global:           globalConfig,
		LocalConfigPath:  localPath,
		GlobalConfigPath: globalPath,
		InRepo:           inRepo,
		RepoRoot:         repoRoot,
	}, nil
}

func selectMutatingConfig(configs configSet, local bool, global bool) (scopedConfig, error) {
	if local && global {
		return scopedConfig{}, fmt.Errorf("use only one of --local or --global")
	}

	if local {
		if !configs.InRepo {
			return scopedConfig{}, fmt.Errorf("local scope requires a git repo")
		}
		localPaths := paths.LocalPaths(configs.RepoRoot)
		return scopedConfig{
			Config:     configs.Local,
			ConfigPath: configs.LocalConfigPath,
			Paths:      localPaths,
			Scope:      scope.ScopeLocal,
			RepoRoot:   configs.RepoRoot,
		}, nil
	}

	globalPaths, err := paths.GlobalPaths()
	if err != nil {
		return scopedConfig{}, err
	}

	return scopedConfig{
		Config:     configs.Global,
		ConfigPath: configs.GlobalConfigPath,
		Paths:      globalPaths,
		Scope:      scope.ScopeGlobal,
		RepoRoot:   configs.RepoRoot,
	}, nil
}
