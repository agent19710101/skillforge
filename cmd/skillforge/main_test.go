package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunListHumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/skills" {
			t.Fatalf("path = %q, want /api/v1/skills", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"skills":[{"name":"git-pr-review","description":"Review pull requests","valid":true}],"total":1,"offset":0,"limit":50}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"list", "--server", server.URL}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "git-pr-review") || !strings.Contains(output, "Review pull requests") {
		t.Fatalf("unexpected stdout: %s", output)
	}
}

func TestRunSearchJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/search" {
			t.Fatalf("path = %q, want /api/v1/search", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "git" {
			t.Fatalf("q = %q, want git", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"query":"git","skills":[{"name":"git-pr-review","valid":true}],"total":1}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"search", "--server", server.URL, "--json", "git"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload["query"] != "git" {
		t.Fatalf("query = %#v, want git", payload["query"])
	}
}

func TestRunGetHumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/skills/git-pr-review" {
			t.Fatalf("path = %q, want /api/v1/skills/git-pr-review", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"git-pr-review","path":"skills/git-pr-review/SKILL.md","description":"Review pull requests","tags":["git","review"],"body":"Use gh pr review.","valid":true}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"get", "--server", server.URL, "git-pr-review"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Name: git-pr-review") || !strings.Contains(output, "Tags: git, review") || !strings.Contains(output, "Use gh pr review.") {
		t.Fatalf("unexpected stdout: %s", output)
	}
}

func TestRunReturnsUsageErrorForUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"wat"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("exitCode = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}
