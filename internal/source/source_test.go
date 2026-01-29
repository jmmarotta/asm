package source

import "testing"

func TestIsRemoteOrigin(t *testing.T) {
	tests := []struct {
		origin string
		want   bool
	}{
		{origin: "https://github.com/org/repo", want: true},
		{origin: "ssh://github.com/org/repo", want: true},
		{origin: "git://github.com/org/repo", want: true},
		{origin: "git@github.com:org/repo", want: true},
		{origin: "file:///tmp/repo", want: false},
		{origin: "/tmp/repo", want: false},
	}

	for _, test := range tests {
		if got := IsRemoteOrigin(test.origin); got != test.want {
			t.Fatalf("IsRemoteOrigin(%q)=%t, want %t", test.origin, got, test.want)
		}
	}
}

func TestValidateOriginScheme(t *testing.T) {
	if err := ValidateOriginScheme("https://github.com/org/repo"); err != nil {
		t.Fatalf("expected https to be allowed, got %v", err)
	}
	if err := ValidateOriginScheme("file:///tmp/repo"); err != nil {
		t.Fatalf("expected file scheme to be allowed, got %v", err)
	}
	if err := ValidateOriginScheme("s3://bucket/repo"); err == nil {
		t.Fatalf("expected unsupported scheme error")
	}
}
