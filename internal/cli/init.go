package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/paths"
)

const initCwdFlag = "cwd"

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a skills manifest",
		RunE:  runInit,
	}

	cmd.Flags().String(initCwdFlag, "", "Initialize in a specific directory")

	return cmd
}

func runInit(cmd *cobra.Command, _ []string) error {
	root, err := resolveInitRoot(cmd)
	if err != nil {
		return err
	}

	manifestPath := ""
	if path, err := config.FindManifestPath(root); err == nil {
		manifestDir := filepath.Dir(path)
		if manifestDir != root {
			return fmt.Errorf("skills.jsonc already exists at %s", path)
		}
		manifestPath = path
	} else if !errors.Is(err, config.ErrManifestNotFound) {
		return err
	} else {
		manifestPath = config.DefaultManifestPath(root)
	}

	if _, err := os.Stat(manifestPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := config.Save(manifestPath, config.Config{Skills: []config.Skill{}}); err != nil {
			return err
		}
	}

	pathsValue := paths.RepoPaths(root)
	if err := os.MkdirAll(pathsValue.StoreDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(pathsValue.CacheDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(pathsValue.SkillsDir, 0o755); err != nil {
		return err
	}

	if err := ensureGitignore(root, []string{".asm/", "skills/"}); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Initialized skills.jsonc")
	return nil
}

func resolveInitRoot(cmd *cobra.Command) (string, error) {
	cwd, err := cmd.Flags().GetString(initCwdFlag)
	if err != nil {
		return "", err
	}
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	root, err := filepath.Abs(cwd)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("cwd does not exist: %s", root)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("cwd is not a directory: %s", root)
	}

	return root, nil
}

func ensureGitignore(root string, entries []string) error {
	path := filepath.Join(root, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := ""
	if err == nil {
		content = string(data)
	}

	lines := strings.Split(content, "\n")
	present := map[string]bool{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "!") {
			continue
		}
		trimmed = strings.TrimPrefix(trimmed, "/")
		trimmed = strings.TrimSuffix(trimmed, "/")
		present[trimmed] = true
	}

	additions := []string{}
	for _, entry := range entries {
		normalized := strings.TrimPrefix(entry, "/")
		normalized = strings.TrimSuffix(normalized, "/")
		if !present[normalized] {
			additions = append(additions, entry)
		}
	}
	if len(additions) == 0 {
		return nil
	}

	builder := &strings.Builder{}
	builder.WriteString(content)
	if content != "" && !strings.HasSuffix(content, "\n") {
		builder.WriteString("\n")
	}
	for _, entry := range additions {
		builder.WriteString(entry)
		builder.WriteString("\n")
	}

	return os.WriteFile(path, []byte(builder.String()), 0o644)
}
