package syncer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Source struct {
	Name string
	Path string
}

type Target struct {
	Name string
	Path string
}

type Warning struct {
	Target  string
	Message string
}

type Result struct {
	Linked   int
	Removed  int
	Warnings []Warning
}

func Sync(targets []Target, sources []Source) (Result, error) {
	result := Result{}
	if len(targets) == 0 || len(sources) == 0 {
		return result, nil
	}

	sourcePaths := make(map[string]string, len(sources))
	safeNames := make(map[string]string, len(sources))
	for _, source := range sources {
		safeName, err := safeNamePath(source.Name)
		if err != nil {
			return result, err
		}

		path, err := filepath.Abs(source.Path)
		if err != nil {
			return result, err
		}

		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				result.Warnings = append(result.Warnings, Warning{
					Message: fmt.Sprintf("source %s missing at %s", source.Name, path),
				})
				continue
			}
			return result, err
		}

		sourcePaths[source.Name] = path
		safeNames[source.Name] = safeName
	}

	for _, target := range targets {
		if target.Path == "" {
			return result, fmt.Errorf("target path is empty")
		}
		if err := ensureDir(target.Path); err != nil {
			return result, err
		}

		for name, sourcePath := range sourcePaths {
			safeName := safeNames[name]
			dest := filepath.Join(target.Path, safeName)
			updated, warning, err := ensureSymlink(dest, sourcePath)
			if err != nil {
				return result, err
			}
			if warning != "" {
				result.Warnings = append(result.Warnings, Warning{
					Target:  dest,
					Message: warning,
				})
				continue
			}
			if updated {
				result.Linked++
			}
		}
	}

	return result, nil
}

func Cleanup(target Target, sources []Source) (Result, error) {
	result := Result{}
	if len(sources) == 0 {
		return result, nil
	}
	if target.Path == "" {
		return result, nil
	}
	if _, err := os.Stat(target.Path); err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}

	for _, source := range sources {
		safeName, err := safeNamePath(source.Name)
		if err != nil {
			return result, err
		}
		dest := filepath.Join(target.Path, safeName)
		removed, warning, err := removeSymlink(dest)
		if err != nil {
			return result, err
		}
		if warning != "" {
			result.Warnings = append(result.Warnings, Warning{Target: dest, Message: warning})
			continue
		}
		if removed {
			result.Removed++
		}
	}

	return result, nil
}

func safeNamePath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("source name is empty")
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return "", fmt.Errorf("source name must be relative: %s", name)
	}

	parts := strings.Split(name, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("invalid source name: %s", name)
		}
	}

	return filepath.Join(parts...), nil
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("target path is not a directory: %s", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(path, 0o755)
}

func ensureSymlink(dest string, source string) (bool, string, error) {
	info, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return false, "", err
			}
			if err := os.Symlink(source, dest); err != nil {
				return false, "", err
			}
			return true, "", nil
		}
		return false, "", err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		same, err := linkMatches(dest, source)
		if err != nil {
			return false, "", err
		}
		if same {
			return false, "", nil
		}
		if err := os.Remove(dest); err != nil {
			return false, "", err
		}
		if err := os.Symlink(source, dest); err != nil {
			return false, "", err
		}
		return true, "", nil
	}

	return false, fmt.Sprintf("destination exists and is not a symlink: %s", dest), nil
}

func removeSymlink(dest string) (bool, string, error) {
	info, err := os.Lstat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return false, "", nil
		}
		return false, "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Sprintf("destination exists and is not a symlink: %s", dest), nil
	}
	if err := os.Remove(dest); err != nil {
		return false, "", err
	}
	return true, "", nil
}

func linkMatches(dest string, source string) (bool, error) {
	link, err := os.Readlink(dest)
	if err != nil {
		return false, err
	}

	resolved := link
	if !filepath.IsAbs(link) {
		resolved = filepath.Join(filepath.Dir(dest), link)
	}

	left, err := filepath.Abs(resolved)
	if err != nil {
		return false, err
	}
	right, err := filepath.Abs(source)
	if err != nil {
		return false, err
	}

	return filepath.Clean(left) == filepath.Clean(right), nil
}
