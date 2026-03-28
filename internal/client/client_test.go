package client

import (
	"context"
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
