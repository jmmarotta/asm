package paths

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigDir string
	StoreDir  string
	CacheDir  string
}

func GlobalPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ConfigDir: filepath.Join(home, ".config", "asm"),
		StoreDir:  filepath.Join(home, ".local", "share", "asm", "store"),
		CacheDir:  filepath.Join(home, ".cache", "asm"),
	}, nil
}

func LocalPaths(repoRoot string) Paths {
	base := filepath.Join(repoRoot, ".asm")
	return Paths{
		ConfigDir: base,
		StoreDir:  filepath.Join(base, "store"),
		CacheDir:  filepath.Join(base, "cache"),
	}
}
