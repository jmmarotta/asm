package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncCreatesSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	target := filepath.Join(root, "target")

	result, err := Sync([]Target{{Name: "t", Path: target}}, []Source{{Name: "foo", Path: source}})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.Linked != 1 {
		t.Fatalf("expected 1 link, got %d", result.Linked)
	}
	assertSymlink(t, filepath.Join(target, "foo"), source)
}

func TestSyncReplacesSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	wrong := filepath.Join(root, "wrong")
	if err := os.MkdirAll(wrong, 0o755); err != nil {
		t.Fatalf("mkdir wrong: %v", err)
	}
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.Symlink(wrong, filepath.Join(target, "foo")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	result, err := Sync([]Target{{Name: "t", Path: target}}, []Source{{Name: "foo", Path: source}})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.Linked != 1 {
		t.Fatalf("expected 1 link, got %d", result.Linked)
	}
	assertSymlink(t, filepath.Join(target, "foo"), source)
}

func TestSyncWarnsOnRealDir(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Join(target, "foo"), 0o755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}

	result, err := Sync([]Target{{Name: "t", Path: target}}, []Source{{Name: "foo", Path: source}})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestSyncNestedName(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	target := filepath.Join(root, "target")

	result, err := Sync([]Target{{Name: "t", Path: target}}, []Source{{Name: "author/skill", Path: source}})
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if result.Linked != 1 {
		t.Fatalf("expected 1 link, got %d", result.Linked)
	}
	assertSymlink(t, filepath.Join(target, "author", "skill"), source)
}

func TestCleanupRemovesSymlink(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.Symlink(source, filepath.Join(target, "foo")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	result, err := Cleanup(Target{Name: "t", Path: target}, []Source{{Name: "foo", Path: source}})
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if result.Removed != 1 {
		t.Fatalf("expected 1 removal, got %d", result.Removed)
	}
	if _, err := os.Lstat(filepath.Join(target, "foo")); !os.IsNotExist(err) {
		t.Fatalf("expected symlink removed")
	}
}

func TestCleanupWarnsOnFile(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "target", "foo")
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	result, err := Cleanup(Target{Name: "t", Path: filepath.Join(root, "target")}, []Source{{Name: "foo", Path: "ignored"}})
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func assertSymlink(t *testing.T, dest string, target string) {
	info, err := os.Lstat(dest)
	if err != nil {
		t.Fatalf("lstat %s: %v", dest, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s", dest)
	}

	link, err := os.Readlink(dest)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}

	resolved := link
	if !filepath.IsAbs(link) {
		resolved = filepath.Join(filepath.Dir(dest), link)
	}

	left, err := filepath.Abs(resolved)
	if err != nil {
		t.Fatalf("abs link: %v", err)
	}
	right, err := filepath.Abs(target)
	if err != nil {
		t.Fatalf("abs target: %v", err)
	}
	if filepath.Clean(left) != filepath.Clean(right) {
		t.Fatalf("expected link %s to point to %s", left, right)
	}
}
