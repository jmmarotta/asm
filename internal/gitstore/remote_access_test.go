package gitstore

import (
	"path/filepath"
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestResolveRemoteAccessStripsCredentials(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "empty"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
	t.Setenv("ASM_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("ASM_GIT_TOKEN", "")
	t.Setenv("ASM_GIT_USERNAME", "")
	t.Setenv("ASM_GIT_PASSWORD", "")

	access, err := ResolveRemoteAccess("https://user:token@github.com/org/repo")
	if err != nil {
		t.Fatalf("ResolveRemoteAccess: %v", err)
	}
	if access.URL != "https://github.com/org/repo" {
		t.Fatalf("expected sanitized url, got %q", access.URL)
	}
	auth, ok := access.Auth.(*githttp.BasicAuth)
	if !ok {
		t.Fatalf("expected basic auth")
	}
	if auth.Username != "user" || auth.Password != "token" {
		t.Fatalf("unexpected auth %q/%q", auth.Username, auth.Password)
	}
}

func TestResolveRemoteAccessUsesGitHubToken(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "empty"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("ASM_GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("ASM_GIT_TOKEN", "")
	t.Setenv("ASM_GIT_USERNAME", "")
	t.Setenv("ASM_GIT_PASSWORD", "")

	access, err := ResolveRemoteAccess("https://github.com/org/repo")
	if err != nil {
		t.Fatalf("ResolveRemoteAccess: %v", err)
	}
	auth, ok := access.Auth.(*githttp.BasicAuth)
	if !ok {
		t.Fatalf("expected basic auth")
	}
	if auth.Username != "x-access-token" || auth.Password != "test-token" {
		t.Fatalf("unexpected auth %q/%q", auth.Username, auth.Password)
	}
}
