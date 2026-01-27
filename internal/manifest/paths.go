package manifest

import "path/filepath"

type Paths struct {
	Root      string
	StoreDir  string
	CacheDir  string
	SkillsDir string
}

func RepoPaths(repoRoot string) Paths {
	base := filepath.Join(repoRoot, ".asm")
	return Paths{
		Root:      repoRoot,
		StoreDir:  filepath.Join(base, "store"),
		CacheDir:  filepath.Join(base, "cache"),
		SkillsDir: filepath.Join(repoRoot, "skills"),
	}
}
