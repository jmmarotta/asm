package cli

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFindCommandOutputsResults(t *testing.T) {
	type request struct {
		path  string
		query url.Values
	}
	requests := make(chan request, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- request{path: r.URL.Path, query: r.URL.Query()}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"skills":[{"id":"vercel-labs/agent-skills/vercel-react-best-practices","skillId":"vercel-react-best-practices","name":"vercel-react-best-practices","installs":123,"source":"vercel-labs/agent-skills"}]}`)
	}))
	defer server.Close()

	t.Setenv("SKILLS_API_URL", server.URL)

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"find", "react", "performance"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("find: %v", err)
	}

	select {
	case req := <-requests:
		if req.path != "/api/search" {
			t.Fatalf("expected path /api/search, got %s", req.path)
		}
		if req.query.Get("q") != "react performance" {
			t.Fatalf("expected query \"react performance\", got %q", req.query.Get("q"))
		}
		if req.query.Get("limit") != "10" {
			t.Fatalf("expected limit 10, got %q", req.query.Get("limit"))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("expected request")
	}

	output := stdout.String()
	if !strings.Contains(output, "asm add https://github.com/vercel-labs/agent-skills --path skills/vercel-react-best-practices") {
		t.Fatalf("expected install command in output")
	}
	if !strings.Contains(output, "https://skills.sh/vercel-labs/agent-skills/vercel-react-best-practices") {
		t.Fatalf("expected skills.sh link in output")
	}
}

func TestFindCommandNoResults(t *testing.T) {
	requests := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"skills":[]}`)
	}))
	defer server.Close()

	t.Setenv("SKILLS_API_URL", server.URL)

	cmd, stdout, _ := newTestCommand()
	cmd.SetArgs([]string{"find", "unknown"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("find: %v", err)
	}

	select {
	case <-requests:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected request")
	}

	output := stdout.String()
	if !strings.Contains(output, "No skills found for \"unknown\".") {
		t.Fatalf("expected no results message")
	}
}
