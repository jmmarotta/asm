package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"

	"github.com/jmmarotta/agent_skills_manager/internal/config"
	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/remote"
	"github.com/jmmarotta/agent_skills_manager/internal/store"
	"github.com/jmmarotta/agent_skills_manager/internal/syncer"
	"github.com/jmmarotta/agent_skills_manager/internal/version"
)

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skills to ./skills",
		RunE:  runInstall,
	}

	return cmd
}

func runInstall(cmd *cobra.Command, _ []string) error {
	state, err := loadManifest()
	if err != nil {
		return err
	}

	return installSkills(cmd.OutOrStdout(), cmd.ErrOrStderr(), state)
}

func installSkills(out io.Writer, errOut io.Writer, state manifestState) error {
	debug.Logf("install skills count=%d", len(state.Config.Skills))
	if len(state.Config.Skills) == 0 {
		prune, err := pruneSkillsDir(state.Paths.SkillsDir, nil)
		if err != nil {
			return err
		}
		printWarnings(prune.Warnings, errOut)
		fmt.Fprintln(out, "No skills found.")
		return nil
	}

	sources, warnings, sumChanged, err := resolveInstallSources(state)
	if err != nil {
		return fmt.Errorf("resolve sources: %w", err)
	}

	result := syncer.Result{}
	if len(sources) > 0 {
		result, err = syncer.Sync([]syncer.Target{{Name: "skills", Path: state.Paths.SkillsDir}}, sources)
		if err != nil {
			return err
		}
	}

	prune, err := pruneSkillsDir(state.Paths.SkillsDir, sources)
	if err != nil {
		return err
	}

	warnings = append(warnings, result.Warnings...)
	warnings = append(warnings, prune.Warnings...)
	printWarnings(warnings, errOut)

	if sumChanged {
		if err := config.SaveSum(state.SumPath, state.Sum); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Installed: %d, Pruned: %d, Warnings: %d\n", result.Linked, prune.Removed, len(warnings))
	return nil
}

func resolveInstallSources(state manifestState) ([]syncer.Source, []syncer.Warning, bool, error) {
	originVersions := make(map[string]string)
	for _, skill := range state.Config.Skills {
		if skill.Type != "git" {
			continue
		}
		originVersions[skill.Origin] = skill.Version
	}

	originPaths := make(map[string]string)
	warnings := []syncer.Warning{}
	sumChanged := false
	for origin, versionValue := range originVersions {
		debug.Logf("resolve origin origin=%s version=%s", debug.SanitizeOrigin(origin), versionValue)
		path, rev, usingReplace, changed, warning, err := resolveOriginRevision(state, origin, versionValue)
		if err != nil {
			return nil, nil, false, fmt.Errorf("resolve origin %s: %w", debug.SanitizeOrigin(origin), err)
		}
		if warning != "" {
			warnings = append(warnings, syncer.Warning{Message: warning})
		}
		sumChanged = sumChanged || changed
		originPaths[origin] = path

		if !usingReplace {
			if err := checkoutRevision(path, rev); err != nil {
				return nil, nil, false, err
			}
		} else {
			if err := warnOnHeadMismatch(path, rev, &warnings); err != nil {
				return nil, nil, false, err
			}
		}
	}

	sources := make([]syncer.Source, 0, len(state.Config.Skills))
	for _, skill := range state.Config.Skills {
		base := skill.Origin
		if skill.Type == "git" {
			base = originPaths[skill.Origin]
		}
		path := base
		if skill.Subdir != "" {
			path = filepath.Join(base, filepath.FromSlash(skill.Subdir))
		}
		sources = append(sources, syncer.Source{Name: skill.Name, Path: path})
	}

	return sources, warnings, sumChanged, nil
}

func resolveOriginRevision(state manifestState, origin string, versionValue string) (string, string, bool, bool, string, error) {
	replacePath := state.Config.Replace[origin]
	if replacePath != "" {
		if info, err := os.Stat(replacePath); err == nil && info.IsDir() {
			rev, changed, err := resolveRevisionFromRepo(replacePath, origin, versionValue, state.Sum, true)
			if err == nil {
				return replacePath, rev, true, changed, "", nil
			}
			warning := fmt.Sprintf("replace path for %s not usable (%v); falling back to remote", origin, err)
			path, rev, changed, err := resolveOriginRevisionFromStore(state, origin, versionValue)
			return path, rev, false, changed, warning, err
		}
		warning := fmt.Sprintf("replace path missing for %s (%s); falling back to remote", origin, replacePath)
		path, rev, changed, err := resolveOriginRevisionFromStore(state, origin, versionValue)
		return path, rev, false, changed, warning, err
	}

	path, rev, changed, err := resolveOriginRevisionFromStore(state, origin, versionValue)
	return path, rev, false, changed, "", err
}

func resolveOriginRevisionFromStore(state manifestState, origin string, versionValue string) (string, string, bool, error) {
	path := store.RepoPath(state.Paths.StoreDir, origin)
	if err := remote.EnsureRepo(path, origin); err != nil {
		return "", "", false, err
	}

	rev, changed, err := resolveRevisionFromRepo(path, origin, versionValue, state.Sum, true)
	if err != nil {
		return "", "", false, err
	}

	return path, rev, changed, nil
}

func resolveRevisionFromRepo(repoPath string, origin string, versionValue string, sum map[config.SumKey]string, strict bool) (string, bool, error) {
	debug.Logf(
		"resolve revision repo=%s origin=%s version=%s strict=%t",
		repoPath,
		debug.SanitizeOrigin(origin),
		versionValue,
		strict,
	)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", false, fmt.Errorf("open repo %s: %w", repoPath, err)
	}

	key := config.SumKey{Origin: origin, Version: versionValue}
	if semver.IsValid(versionValue) {
		rev, err := version.ResolveForVersion(repo, versionValue)
		if err != nil {
			return "", false, err
		}
		if existing, ok := sum[key]; ok && existing != rev {
			if strict {
				return "", false, fmt.Errorf("version %s moved for %s; run asm update", versionValue, origin)
			}
		}
		if sum[key] != rev {
			sum[key] = rev
			return rev, true, nil
		}
		return rev, false, nil
	}
	if !module.IsPseudoVersion(versionValue) {
		return "", false, fmt.Errorf("invalid version %q", versionValue)
	}

	if existing, ok := sum[key]; ok {
		revPrefix := pseudoVersionRev(versionValue)
		if revPrefix == "" || !strings.HasPrefix(existing, revPrefix) {
			return "", false, fmt.Errorf("skills.sum entry for %s %s does not match version", origin, versionValue)
		}
		if _, err := repo.CommitObject(plumbing.NewHash(existing)); err == nil {
			return existing, false, nil
		}
	}

	rev, err := version.ResolveForVersion(repo, versionValue)
	if err != nil {
		return "", false, err
	}
	if sum[key] != rev {
		sum[key] = rev
		return rev, true, nil
	}
	return rev, false, nil
}

func checkoutRevision(repoPath string, rev string) error {
	debug.Logf("checkout repo=%s rev=%s", repoPath, rev)
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("open repo %s: %w", repoPath, err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("open worktree: %w", err)
	}
	if err := worktree.Checkout(&git.CheckoutOptions{Force: true, Hash: plumbing.NewHash(rev)}); err != nil {
		return fmt.Errorf("checkout %s: %w", rev, err)
	}
	return nil
}

func warnOnHeadMismatch(repoPath string, rev string, warnings *[]syncer.Warning) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("open repo %s: %w", repoPath, err)
	}
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("read head: %w", err)
	}
	if head.Hash().String() != rev {
		*warnings = append(*warnings, syncer.Warning{
			Message: fmt.Sprintf("replace repo %s is at %s, expected %s", repoPath, head.Hash().String(), rev),
		})
	}
	return nil
}

type pruneResult struct {
	Removed  int
	Warnings []syncer.Warning
}

func pruneSkillsDir(root string, sources []syncer.Source) (pruneResult, error) {
	result := pruneResult{}
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return result, err
	}
	if !info.IsDir() {
		return result, fmt.Errorf("skills path is not a directory: %s", root)
	}

	keep, keepDirs, err := buildKeepSets(sources)
	if err != nil {
		return result, err
	}

	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		relative, err := filepath.Rel(root, path)
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
			result.Warnings = append(result.Warnings, syncer.Warning{Target: path, Message: "destination exists and is not a symlink"})
			return nil
		}
		if _, ok := keepDirs[relative]; !ok {
			result.Warnings = append(result.Warnings, syncer.Warning{Target: path, Message: "unmanaged entry exists"})
		}
		return nil
	}); err != nil {
		return result, err
	}

	if err := pruneEmptyDirs(root, keepDirs, &result); err != nil {
		return result, err
	}

	return result, nil
}

func buildKeepSets(sources []syncer.Source) (map[string]struct{}, map[string]struct{}, error) {
	keep := make(map[string]struct{}, len(sources))
	keepDirs := make(map[string]struct{})
	for _, source := range sources {
		safeName, err := syncer.SafeNamePath(source.Name)
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

func pruneEmptyDirs(root string, keepDirs map[string]struct{}, result *pruneResult) error {
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

func printWarnings(warnings []syncer.Warning, writer io.Writer) {
	for _, warning := range warnings {
		if warning.Target != "" {
			fmt.Fprintf(writer, "warning: %s (%s)\n", warning.Message, warning.Target)
			continue
		}
		fmt.Fprintf(writer, "warning: %s\n", warning.Message)
	}
}

func pseudoVersionRev(versionValue string) string {
	base := strings.SplitN(versionValue, "+", 2)[0]
	idx := strings.LastIndex(base, "-")
	if idx == -1 || idx == len(base)-1 {
		return ""
	}
	return base[idx+1:]
}
