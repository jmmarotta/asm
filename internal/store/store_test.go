package store

import "testing"

func TestRepoKeyStable(t *testing.T) {
	first := RepoKey("https://example.com/repo")
	second := RepoKey("https://example.com/repo")
	if first != second {
		t.Fatalf("expected stable repo key")
	}

	other := RepoKey("https://example.com/other")
	if first == other {
		t.Fatalf("expected different repo keys for different origins")
	}
}
