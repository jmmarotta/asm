package asm

import (
	"errors"
	"strings"
	"testing"
)

func TestUpgradeRunsGoInstallLatest(t *testing.T) {
	originalRunner := upgradeCommandRunner
	t.Cleanup(func() {
		upgradeCommandRunner = originalRunner
	})

	var gotName string
	var gotArgs []string
	upgradeCommandRunner = func(name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil, nil
	}

	report, err := Upgrade()
	if err != nil {
		t.Fatalf("Upgrade: %v", err)
	}
	if gotName != "go" {
		t.Fatalf("expected command %q, got %q", "go", gotName)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "install" || gotArgs[1] != upgradeInstallTarget {
		t.Fatalf("unexpected args: %v", gotArgs)
	}
	if report.Target != upgradeInstallTarget {
		t.Fatalf("expected target %q, got %q", upgradeInstallTarget, report.Target)
	}
}

func TestUpgradeIncludesGoInstallOutputOnFailure(t *testing.T) {
	originalRunner := upgradeCommandRunner
	t.Cleanup(func() {
		upgradeCommandRunner = originalRunner
	})

	upgradeCommandRunner = func(string, ...string) ([]byte, error) {
		return []byte("toolchain not available\n"), errors.New("exit status 1")
	}

	_, err := Upgrade()
	if err == nil {
		t.Fatalf("expected error")
	}
	message := err.Error()
	if !strings.Contains(message, upgradeInstallTarget) {
		t.Fatalf("expected target in error: %q", message)
	}
	if !strings.Contains(message, "toolchain not available") {
		t.Fatalf("expected command output in error: %q", message)
	}
	if !strings.Contains(message, "exit status 1") {
		t.Fatalf("expected wrapped error in message: %q", message)
	}
}
