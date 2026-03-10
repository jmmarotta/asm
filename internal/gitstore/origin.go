package gitstore

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/mod/module"

	"github.com/jmmarotta/agent_skills_manager/internal/debug"
	"github.com/jmmarotta/agent_skills_manager/internal/manifest"
)

const defaultOriginResolveParallelism = 4

type PrefetchedOrigin struct {
	Path         string
	Rev          string
	UsingReplace bool
}

type OriginRequest struct {
	Origin      string
	Version     string
	ReplacePath string
	Prefetched  *PrefetchedOrigin
}

type ResolveBatchOptions struct {
	Strict      bool
	MaxParallel int
}

type OriginResolution struct {
	Path         string
	Rev          string
	UsingReplace bool
	LockChanged  bool
	Warning      string
}

type OriginPathsResult struct {
	Paths       map[string]string
	Resolutions map[string]OriginResolution
	Warnings    []string
	LockChanged bool
}

func ResolveOriginRevision(storeDir string, origin string, version string, replacePath string, lock map[manifest.LockKey]string, strict bool) (OriginResolution, error) {
	if replacePath != "" {
		if info, err := os.Stat(replacePath); err == nil && info.IsDir() {
			rev, changed, err := ResolveRevision(replacePath, origin, version, lock, strict)
			if err == nil {
				return OriginResolution{
					Path:         replacePath,
					Rev:          rev,
					UsingReplace: true,
					LockChanged:  changed,
				}, nil
			}
			warning := fmt.Sprintf("replace path for %s not usable (%v); falling back to remote", origin, err)
			return resolveOriginFromStore(storeDir, origin, version, lock, strict, warning)
		}
		warning := fmt.Sprintf("replace path missing for %s (%s); falling back to remote", origin, replacePath)
		return resolveOriginFromStore(storeDir, origin, version, lock, strict, warning)
	}

	return resolveOriginFromStore(storeDir, origin, version, lock, strict, "")
}

func ResolveOrigins(storeDir string, origins map[string]string, replace map[string]string, lock map[manifest.LockKey]string, strict bool) (OriginPathsResult, error) {
	requests := make([]OriginRequest, 0, len(origins))
	originList := make([]string, 0, len(origins))
	for origin := range origins {
		originList = append(originList, origin)
	}
	sort.Strings(originList)
	for _, origin := range originList {
		replacePath := ""
		if replace != nil {
			replacePath = replace[origin]
		}
		requests = append(requests, OriginRequest{
			Origin:      origin,
			Version:     origins[origin],
			ReplacePath: replacePath,
		})
	}

	return ResolveOriginBatch(storeDir, requests, lock, ResolveBatchOptions{
		Strict:      strict,
		MaxParallel: defaultOriginResolveParallelism,
	})
}

func ResolveOriginBatch(storeDir string, requests []OriginRequest, lock map[manifest.LockKey]string, opts ResolveBatchOptions) (OriginPathsResult, error) {
	result := OriginPathsResult{
		Paths:       map[string]string{},
		Resolutions: map[string]OriginResolution{},
	}
	if len(requests) == 0 {
		return result, nil
	}

	type batchItem struct {
		index      int
		request    OriginRequest
		resolution OriginResolution
		err        error
	}

	maxParallel := opts.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 1
	}
	if maxParallel > len(requests) {
		maxParallel = len(requests)
	}

	jobs := make(chan int)
	results := make(chan batchItem, len(requests))

	var wg sync.WaitGroup
	for range maxParallel {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				request := requests[index]
				resolution, err := resolveOriginRequest(storeDir, request, lock, opts.Strict)
				results <- batchItem{index: index, request: request, resolution: resolution, err: err}
			}
		}()
	}

	for index := range requests {
		jobs <- index
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	ordered := make([]batchItem, len(requests))
	for item := range results {
		ordered[item.index] = item
	}

	for _, item := range ordered {
		if item.err != nil {
			return result, fmt.Errorf("resolve origin %s: %w", debug.SanitizeOrigin(item.request.Origin), item.err)
		}
		if item.resolution.Warning != "" {
			result.Warnings = append(result.Warnings, item.resolution.Warning)
		}
		if item.resolution.LockChanged {
			result.LockChanged = true
			if lock != nil {
				lock[manifest.LockKey{Origin: item.request.Origin, Version: item.request.Version}] = item.resolution.Rev
			}
		}
		result.Paths[item.request.Origin] = item.resolution.Path
		result.Resolutions[item.request.Origin] = item.resolution
		applyWarning, err := ApplyOriginResolution(item.resolution)
		if err != nil {
			return result, err
		}
		if applyWarning != "" {
			result.Warnings = append(result.Warnings, applyWarning)
		}
	}

	return result, nil
}

func resolveOriginRequest(storeDir string, request OriginRequest, lock map[manifest.LockKey]string, strict bool) (OriginResolution, error) {
	debug.Logf("resolve origin origin=%s version=%s", debug.SanitizeOrigin(request.Origin), request.Version)
	if resolution, ok, err := resolvePrefetchedOrigin(request, lock); ok || err != nil {
		return resolution, err
	}

	return ResolveOriginRevision(
		storeDir,
		request.Origin,
		request.Version,
		request.ReplacePath,
		lockEntry(lock, request.Origin, request.Version),
		strict,
	)
}

func resolvePrefetchedOrigin(request OriginRequest, lock map[manifest.LockKey]string) (OriginResolution, bool, error) {
	if request.Prefetched == nil {
		return OriginResolution{}, false, nil
	}
	if request.Prefetched.Path == "" || request.Prefetched.Rev == "" {
		return OriginResolution{}, false, nil
	}
	if !module.IsPseudoVersion(request.Version) {
		return OriginResolution{}, false, nil
	}

	prefix := pseudoVersionRev(request.Version)
	if prefix == "" || !strings.HasPrefix(request.Prefetched.Rev, prefix) {
		return OriginResolution{}, false, nil
	}

	exists, err := CommitExists(request.Prefetched.Path, request.Prefetched.Rev)
	if err != nil || !exists {
		return OriginResolution{}, false, nil
	}

	key := manifest.LockKey{Origin: request.Origin, Version: request.Version}
	return OriginResolution{
		Path:         request.Prefetched.Path,
		Rev:          request.Prefetched.Rev,
		UsingReplace: request.Prefetched.UsingReplace,
		LockChanged:  lock == nil || lock[key] != request.Prefetched.Rev,
	}, true, nil
}

func lockEntry(lock map[manifest.LockKey]string, origin string, version string) map[manifest.LockKey]string {
	entries := map[manifest.LockKey]string{}
	if lock == nil {
		return entries
	}
	key := manifest.LockKey{Origin: origin, Version: version}
	if rev, ok := lock[key]; ok {
		entries[key] = rev
	}
	return entries
}

func resolveOriginFromStore(storeDir string, origin string, version string, lock map[manifest.LockKey]string, strict bool, warning string) (OriginResolution, error) {
	path := RepoPath(storeDir, origin)
	if err := EnsureRepo(path, origin); err != nil {
		return OriginResolution{}, err
	}

	rev, changed, err := ResolveRevision(path, origin, version, lock, strict)
	if err != nil {
		return OriginResolution{}, err
	}

	return OriginResolution{
		Path:        path,
		Rev:         rev,
		LockChanged: changed,
		Warning:     warning,
	}, nil
}

func ApplyOriginResolution(resolution OriginResolution) (string, error) {
	if resolution.Path == "" {
		return "", fmt.Errorf("resolved path is empty")
	}
	if resolution.UsingReplace {
		head, err := HeadHash(resolution.Path)
		if err != nil {
			return "", err
		}
		if head != resolution.Rev {
			return fmt.Sprintf("replace repo %s is at %s, expected %s", resolution.Path, head, resolution.Rev), nil
		}
		return "", nil
	}

	return "", CheckoutRevision(resolution.Path, resolution.Rev)
}
