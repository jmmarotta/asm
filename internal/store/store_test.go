package store

import "testing"

func TestRepoKeyStable(t *testing.T) {
	first := RepoKey("https://example.com/repo", "main")
	second := RepoKey("https://example.com/repo", "main")
	if first != second {
		t.Fatalf("expected stable repo key")
	}

	other := RepoKey("https://example.com/repo", "dev")
	if first == other {
		t.Fatalf("expected different repo keys for different refs")
	}
}
