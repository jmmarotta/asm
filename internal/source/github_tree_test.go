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

func TestParseGitHubTreeURLLoose(t *testing.T) {
	input := "https://github.com/org/repo/tree/main/plugins/foo"
	spec, ok, err := ParseGitHubTreeURLLoose(input)
	if err != nil {
		t.Fatalf("ParseGitHubTreeURLLoose: %v", err)
	}
	if !ok {
		t.Fatalf("expected github tree url to match")
	}
	if spec.Origin != "https://github.com/org/repo" {
		t.Fatalf("expected origin https://github.com/org/repo, got %q", spec.Origin)
	}
	if spec.Ref != "main" {
		t.Fatalf("expected ref main, got %q", spec.Ref)
	}
	if spec.Subdir != "plugins/foo" {
		t.Fatalf("expected subdir plugins/foo, got %q", spec.Subdir)
	}
}

func TestParseGitHubTreeURLLooseMissingRef(t *testing.T) {
	_, ok, err := ParseGitHubTreeURLLoose("https://github.com/org/repo/tree/")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ok {
		t.Fatalf("expected github tree url to not match")
	}
}
