package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func SyncAndPrune(target Target, sources []Source) (Result, error) {
	result := Result{}
	if target.Path == "" {
		return result, fmt.Errorf("target path is empty")
	}

	if len(sources) > 0 {
		syncResult, err := Sync([]Target{target}, sources)
		if err != nil {
			return result, err
		}
		result.Linked = syncResult.Linked
		result.Warnings = append(result.Warnings, syncResult.Warnings...)
	}

	pruneResult, err := Prune(target, sources)
	if err != nil {
		return result, err
	}
	result.Removed += pruneResult.Removed
	result.Warnings = append(result.Warnings, pruneResult.Warnings...)
	return result, nil
}

func Prune(target Target, sources []Source) (Result, error) {
	result := Result{}
	if target.Path == "" {
		return result, fmt.Errorf("target path is empty")
	}
	info, err := os.Stat(target.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	if !info.IsDir() {
		return result, fmt.Errorf("target path is not a directory: %s", target.Path)
	}

	keep, keepDirs, err := buildKeepSets(sources)
	if err != nil {
		return result, err
	}

	if err := filepath.WalkDir(target.Path, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == target.Path {
			return nil
		}
		relative, err := filepath.Rel(target.Path, path)
		if err != nil {
			return err
		}
		relative = filepath.Clean(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			if _, ok := keep[relative]; !ok {
				if err := os.Remove(path); err != nil {
					return err
				}
				result.Removed++
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if _, ok := keep[relative]; ok {
			result.Warnings = append(result.Warnings, Warning{Target: path, Message: "destination exists and is not a symlink"})
			return nil
		}
		if _, ok := keepDirs[relative]; !ok {
			result.Warnings = append(result.Warnings, Warning{Target: path, Message: "unmanaged entry exists"})
		}
		return nil
	}); err != nil {
		return result, err
	}

	if err := pruneEmptyDirs(target.Path, keepDirs, &result); err != nil {
		return result, err
	}

	return result, nil
}

func buildKeepSets(sources []Source) (map[string]struct{}, map[string]struct{}, error) {
	keep := make(map[string]struct{}, len(sources))
	keepDirs := make(map[string]struct{})
	for _, source := range sources {
		safeName, err := safeNamePath(source.Name)
		if err != nil {
			return nil, nil, err
		}
		safeName = filepath.Clean(safeName)
		keep[safeName] = struct{}{}
		parts := strings.Split(safeName, string(filepath.Separator))
		if len(parts) > 1 {
			current := ""
			for _, part := range parts[:len(parts)-1] {
				if current == "" {
					current = part
				} else {
					current = filepath.Join(current, part)
				}
				keepDirs[current] = struct{}{}
			}
		}
	}

	return keep, keepDirs, nil
}

func pruneEmptyDirs(root string, keepDirs map[string]struct{}, result *Result) error {
	var dirs []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	}); err != nil {
		return err
	}

	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, dir := range dirs {
		relative, err := filepath.Rel(root, dir)
		if err != nil {
			return err
		}
		relative = filepath.Clean(relative)
		if _, ok := keepDirs[relative]; ok {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil {
				return err
			}
			result.Removed++
		}
	}

	return nil
}
