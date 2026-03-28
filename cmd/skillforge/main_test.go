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

func TestRunDraftCreateHumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/drafts" {
			t.Fatalf("path = %q, want /api/v1/drafts", r.URL.Path)
		}
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["operation"] != "create" || req["skillName"] != "new-skill" {
			t.Fatalf("unexpected request: %#v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"draft01","operation":"create","skillName":"new-skill","branchName":"draft/new-skill","createdAt":"2026-03-28T18:30:00Z","validation":{"valid":true},"submission":{"enabled":true,"baseBranch":"main"}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"draft", "create", "--server", server.URL, "--operation", "create", "--skill", "new-skill", "--content", "# New skill\n"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Draft: draft01") || !strings.Contains(output, "Submission enabled: true") {
		t.Fatalf("unexpected stdout: %s", output)
	}
}

func TestRunDraftStatusJSONOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/drafts/draft01" {
			t.Fatalf("path = %q, want /api/v1/drafts/draft01", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"draft01","operation":"update","skillName":"git-pr-review","branchName":"draft/git-pr-review","createdAt":"2026-03-28T18:30:00Z","validation":{"valid":true},"submission":{"enabled":false,"reason":"not configured"}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"draft", "status", "--server", server.URL, "--json", "draft01"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if payload["id"] != "draft01" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRunDraftSubmitHumanOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/drafts/draft01/submit" {
			t.Fatalf("path = %q, want /api/v1/drafts/draft01/submit", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"draft01","operation":"create","skillName":"new-skill","branchName":"draft/new-skill","baseBranch":"main","commitHash":"abc123","pullRequest":{"number":17,"url":"https://forgejo.example/pr/17"},"validation":{"valid":true}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"draft", "submit", "--server", server.URL, "draft01"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "Submitted: yes") || !strings.Contains(output, "Pull request: #17") {
		t.Fatalf("unexpected stdout: %s", output)
	}
}

func TestRunDraftSubmitSubmissionUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"submission_unavailable","message":"submission backend is not configured","submission":{"enabled":false,"reason":"submission backend is not configured"}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"draft", "submit", "--server", server.URL, "draft01"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	if !strings.Contains(stdout.String(), "Submission unavailable") || !strings.Contains(stdout.String(), "submission backend is not configured") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "submission_unavailable") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestRunDraftSubmitDraftInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":"draft_invalid","message":"draft validation failed","validation":{"valid":false,"findings":[{"code":"missing_description","message":"description is required","path":"skills/new-skill/SKILL.md"}]},"submission":{"enabled":true,"baseBranch":"main"}}`))
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"draft", "submit", "--server", server.URL, "draft01"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", exitCode)
	}
	output := stdout.String()
	if !strings.Contains(output, "Submission blocked") || !strings.Contains(output, "description is required") || !strings.Contains(output, "Submission enabled: true") {
		t.Fatalf("unexpected stdout: %s", output)
	}
	if !strings.Contains(stderr.String(), "draft_invalid") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}
