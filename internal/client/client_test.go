package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientListSkills(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/skills" {
			t.Fatalf("path = %q, want /api/v1/skills", r.URL.Path)
		}
		if got := r.URL.Query().Get("validation"); got != "valid" {
			t.Fatalf("validation = %q, want valid", got)
		}
		if got := r.URL.Query().Get("offset"); got != "10" {
			t.Fatalf("offset = %q, want 10", got)
		}
		if got := r.URL.Query().Get("limit"); got != "20" {
			t.Fatalf("limit = %q, want 20", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"skills":[{"name":"git-pr-review","description":"Review PRs","valid":true}],"total":1,"offset":10,"limit":20}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	resp, err := c.ListSkills(context.Background(), ListOptions{Validation: "valid", Offset: 10, Limit: 20})
	if err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
	if resp.Total != 1 || len(resp.Skills) != 1 || resp.Skills[0].Name != "git-pr-review" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientSearchSkills(t *testing.T) {
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

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	resp, err := c.SearchSkills(context.Background(), "git")
	if err != nil {
		t.Fatalf("SearchSkills() error = %v", err)
	}
	if resp.Query != "git" || len(resp.Skills) != 1 || resp.Skills[0].Name != "git-pr-review" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientGetSkill(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/skills/git-pr-review" {
			t.Fatalf("path = %q, want /api/v1/skills/git-pr-review", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"git-pr-review","description":"Review PRs","path":"skills/git-pr-review/SKILL.md","valid":true}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	skill, err := c.GetSkill(context.Background(), "git-pr-review")
	if err != nil {
		t.Fatalf("GetSkill() error = %v", err)
	}
	if skill.Name != "git-pr-review" || !skill.Valid {
		t.Fatalf("unexpected skill: %+v", skill)
	}
}

func TestClientReturnsTypedAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not_found","message":"skill not found"}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = c.GetSkill(context.Background(), "missing-skill")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("GetSkill() error = %v, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound || apiErr.Code != "not_found" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestClientCreateDraft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/drafts" {
			t.Fatalf("path = %q, want /api/v1/drafts", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q, want application/json", got)
		}
		var req DraftCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Operation != "create" || req.SkillName != "new-skill" || req.Content != "# New skill\n" {
			t.Fatalf("unexpected request: %+v", req)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"draft01","operation":"create","skillName":"new-skill","branchName":"draft/new-skill","createdAt":"2026-03-28T18:30:00Z","validation":{"valid":true},"submission":{"enabled":true,"baseBranch":"main"}}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	resp, err := c.CreateDraft(context.Background(), DraftCreateRequest{Operation: "create", SkillName: "new-skill", Content: "# New skill\n"})
	if err != nil {
		t.Fatalf("CreateDraft() error = %v", err)
	}
	if resp.ID != "draft01" || resp.SkillName != "new-skill" || !resp.Submission.Enabled {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientGetDraft(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/v1/drafts/draft01" {
			t.Fatalf("path = %q, want /api/v1/drafts/draft01", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"draft01","operation":"update","skillName":"git-pr-review","branchName":"draft/git-pr-review","createdAt":"2026-03-28T18:30:00Z","validation":{"valid":true},"submission":{"enabled":false,"reason":"not configured"}}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	resp, err := c.GetDraft(context.Background(), "draft01")
	if err != nil {
		t.Fatalf("GetDraft() error = %v", err)
	}
	if resp.ID != "draft01" || resp.Operation != "update" || resp.Submission.Reason != "not configured" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientSubmitDraft(t *testing.T) {
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

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	resp, err := c.SubmitDraft(context.Background(), "draft01")
	if err != nil {
		t.Fatalf("SubmitDraft() error = %v", err)
	}
	if resp.CommitHash != "abc123" || resp.PullRequest == nil || resp.PullRequest.Number != 17 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientSubmitDraftReturnsTypedSubmissionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":"submission_unavailable","message":"submission backend is not configured","submission":{"enabled":false,"reason":"submission backend is not configured"}}`))
	}))
	defer server.Close()

	c, err := New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = c.SubmitDraft(context.Background(), "draft01")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("SubmitDraft() error = %v, want *APIError", err)
	}
	if apiErr.Code != "submission_unavailable" || apiErr.Submission == nil || apiErr.Submission.Enabled {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
}

func TestClientPreservesBaseURLPathPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/skillforge/api/v1/skills" {
			t.Fatalf("path = %q, want /skillforge/api/v1/skills", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"skills":[],"total":0,"offset":0,"limit":0}`))
	}))
	defer server.Close()

	c, err := New(server.URL + "/skillforge")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := c.ListSkills(context.Background(), ListOptions{}); err != nil {
		t.Fatalf("ListSkills() error = %v", err)
	}
}
