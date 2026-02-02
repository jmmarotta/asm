package gitstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNetrcCredentialsMachine(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	writeFile(t, home, ".netrc", "machine github.com login alice password secret\n")

	t.Setenv("HOME", home)
	t.Setenv("NETRC", "")

	user, pass, ok, err := netrcCredentials("github.com")
	if err != nil {
		t.Fatalf("netrcCredentials: %v", err)
	}
	if !ok {
		t.Fatalf("expected credentials")
	}
	if user != "alice" || pass != "secret" {
		t.Fatalf("unexpected credentials %q/%q", user, pass)
	}
}

func TestNetrcCredentialsDefault(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	writeFile(t, home, ".netrc", "default login token password pass\n")

	t.Setenv("HOME", home)
	t.Setenv("NETRC", "")

	user, pass, ok, err := netrcCredentials("example.com")
	if err != nil {
		t.Fatalf("netrcCredentials: %v", err)
	}
	if !ok {
		t.Fatalf("expected default credentials")
	}
	if user != "token" || pass != "pass" {
		t.Fatalf("unexpected credentials %q/%q", user, pass)
	}
}
