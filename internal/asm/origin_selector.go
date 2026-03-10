package asm

import (
	"fmt"
	"path/filepath"

	"github.com/jmmarotta/agent_skills_manager/internal/source"
)

func normalizeOriginSelector(value string) (string, error) {
	if source.IsGitHubTreeURL(value) {
		origin, ok, err := source.GitHubTreeOrigin(value)
		if err != nil {
			return "", fmt.Errorf("parse github tree origin: %w", err)
		}
		if ok {
			value = origin
		}
	}

	origin, _, err := source.NormalizeFileOrigin(value)
	if err != nil {
		return "", err
	}
	if err := source.ValidateOriginScheme(origin); err != nil {
		return "", err
	}

	if source.IsRemoteOrigin(origin) {
		origin, _ = source.ParseOriginRef(origin)
		return source.NormalizeOrigin(origin), nil
	}

	abs, err := filepath.Abs(origin)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func sameOrigin(left string, right string) bool {
	if left == right {
		return true
	}
	if source.IsRemoteOrigin(left) || source.IsRemoteOrigin(right) {
		return false
	}
	return sameLocalOriginPath(left, right)
}

func sameLocalOriginPath(left string, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	if leftErr != nil {
		leftAbs = filepath.Clean(left)
	}
	rightAbs, rightErr := filepath.Abs(right)
	if rightErr != nil {
		rightAbs = filepath.Clean(right)
	}
	if filepath.Clean(leftAbs) == filepath.Clean(rightAbs) {
		return true
	}

	leftResolved, leftErr := filepath.EvalSymlinks(leftAbs)
	rightResolved, rightErr := filepath.EvalSymlinks(rightAbs)
	if leftErr != nil || rightErr != nil {
		return false
	}
	return filepath.Clean(leftResolved) == filepath.Clean(rightResolved)
}
