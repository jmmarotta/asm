package source

import "testing"

func TestParseGitHubTreeURLLongestRefMatch(t *testing.T) {
	refs := map[string]struct{}{
		"feature":     {},
		"feature/foo": {},
	}

	input := "https://github.com/org/repo/tree/feature/foo/plugins/bar"
	spec, ok, err := ParseGitHubTreeURL(input, refs)
	if err != nil {
		t.Fatalf("ParseGitHubTreeURL: %v", err)
	}
	if !ok {
		t.Fatalf("expected github tree url to match")
	}
	if spec.Ref != "feature/foo" {
		t.Fatalf("expected ref feature/foo, got %q", spec.Ref)
	}
	if spec.Subdir != "plugins/bar" {
		t.Fatalf("expected subdir plugins/bar, got %q", spec.Subdir)
	}
}

func TestParseGitHubTreeURLNoMatch(t *testing.T) {
	refs := map[string]struct{}{"main": {}}

	input := "https://github.com/org/repo/tree/unknown/path"
	_, ok, err := ParseGitHubTreeURL(input, refs)
	if err == nil {
		t.Fatalf("expected error for unmatched ref")
	}
	if !ok {
		t.Fatalf("expected github tree url to match")
	}
}
