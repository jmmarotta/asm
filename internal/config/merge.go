package config

import (
	"sort"

	"github.com/jmmarotta/agent_skills_manager/internal/scope"
)

type ScopedSource struct {
	Source
	Scope scope.Scope `json:"scope"`
}

type ScopedTarget struct {
	Target
	Scope scope.Scope `json:"scope"`
}

func MergeSources(local []Source, global []Source) []ScopedSource {
	merged := map[string]ScopedSource{}
	for _, source := range global {
		merged[source.Name] = ScopedSource{Source: source, Scope: scope.ScopeGlobal}
	}
	for _, source := range local {
		merged[source.Name] = ScopedSource{Source: source, Scope: scope.ScopeLocal}
	}

	return sortedSources(merged)
}

func MergeTargets(local []Target, global []Target) []ScopedTarget {
	merged := map[string]ScopedTarget{}
	for _, target := range global {
		merged[target.Name] = ScopedTarget{Target: target, Scope: scope.ScopeGlobal}
	}
	for _, target := range local {
		merged[target.Name] = ScopedTarget{Target: target, Scope: scope.ScopeLocal}
	}

	return sortedTargets(merged)
}

func ScopedSourcesFrom(sources []Source, scoped scope.Scope) []ScopedSource {
	merged := map[string]ScopedSource{}
	for _, source := range sources {
		merged[source.Name] = ScopedSource{Source: source, Scope: scoped}
	}
	return sortedSources(merged)
}

func ScopedTargetsFrom(targets []Target, scoped scope.Scope) []ScopedTarget {
	merged := map[string]ScopedTarget{}
	for _, target := range targets {
		merged[target.Name] = ScopedTarget{Target: target, Scope: scoped}
	}
	return sortedTargets(merged)
}

func FindSource(sources []Source, name string) (Source, bool) {
	for _, source := range sources {
		if source.Name == name {
			return source, true
		}
	}
	return Source{}, false
}

func FindTarget(targets []Target, name string) (Target, bool) {
	for _, target := range targets {
		if target.Name == name {
			return target, true
		}
	}
	return Target{}, false
}

func sortedSources(values map[string]ScopedSource) []ScopedSource {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	results := make([]ScopedSource, 0, len(names))
	for _, name := range names {
		results = append(results, values[name])
	}
	return results
}

func sortedTargets(values map[string]ScopedTarget) []ScopedTarget {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)

	results := make([]ScopedTarget, 0, len(names))
	for _, name := range names {
		results = append(results, values[name])
	}
	return results
}
