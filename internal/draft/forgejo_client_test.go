package draft

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewForgejoClient(t *testing.T) {
	tests := []struct {
		name string
		cfg  SubmissionConfig
		want string
	}{
		{
			name: "missing server URL",
			cfg: SubmissionConfig{
				Token: "secret",
			},
			want: "forgejo client: forgejo server URL is required",
		},
		{
			name: "missing token for token auth",
			cfg: SubmissionConfig{
				ServerURL: "https://forgejo.example",
			},
			want: "forgejo client: forgejo token is required for token auth",
		},
		{
			name: "unsupported auth method",
			cfg: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				AuthMethod: "basic",
			},
			want: "forgejo client: unsupported auth method \"basic\"",
		},
		{
			name: "no auth allowed",
			cfg: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				AuthMethod: ForgejoAuthMethodNone,
			},
		},
		{
			name: "token auth allowed",
			cfg: SubmissionConfig{
				ServerURL: "https://forgejo.example",
				Token:     "secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewForgejoClient(tt.cfg)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("NewForgejoClient() error = %v", err)
				}
				if client == nil {
					t.Fatal("NewForgejoClient() returned nil client")
				}
				return
			}
			if err == nil || err.Error() != tt.want {
				t.Fatalf("NewForgejoClient() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestForgejoClientCreatePullRequestShapesRequestAndParsesResponse(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotAuthorization string
	var gotAccept string
	var gotContentType string
	var gotUserAgent string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuthorization = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotContentType = r.Header.Get("Content-Type")
		gotUserAgent = r.Header.Get("User-Agent")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("Decode(request) error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"number":17,"id":42,"html_url":"https://forgejo.example/acme/skillforge/pulls/17"}`))
	}))
	defer server.Close()

	client, err := NewForgejoClient(SubmissionConfig{
		ServerURL: server.URL + "/forgejo/",
		Token:     "secret",
	})
	if err != nil {
		t.Fatalf("NewForgejoClient() error = %v", err)
	}

	pr, err := client.CreatePullRequest(context.Background(), PullRequestRequest{
		Owner:      "acme",
		Repo:       "skillforge",
		HeadBranch: "skillforge/create/new-skill/submit01",
		BaseBranch: "main",
		Title:      "skillforge: create new-skill",
		Body:       "Operation: create\nSkill: new-skill",
	})
	if err != nil {
		t.Fatalf("CreatePullRequest() error = %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/forgejo/api/v1/repos/acme/skillforge/pulls" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuthorization != "token secret" {
		t.Fatalf("authorization = %q", gotAuthorization)
	}
	if gotAccept != "application/json" {
		t.Fatalf("accept = %q", gotAccept)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content type = %q", gotContentType)
	}
	if gotUserAgent != defaultForgejoUserAgent {
		t.Fatalf("user agent = %q", gotUserAgent)
	}
	if gotBody["head"] != "skillforge/create/new-skill/submit01" {
		t.Fatalf("unexpected head: %#v", gotBody)
	}
	if gotBody["base"] != "main" {
		t.Fatalf("unexpected base: %#v", gotBody)
	}
	if gotBody["title"] != "skillforge: create new-skill" {
		t.Fatalf("unexpected title: %#v", gotBody)
	}
	if gotBody["body"] != "Operation: create\nSkill: new-skill" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
	if pr.Number != 17 || pr.ID != 42 || pr.URL != "https://forgejo.example/acme/skillforge/pulls/17" {
		t.Fatalf("unexpected pull request = %#v", pr)
	}
}

func TestForgejoClientCreatePullRequestReturnsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"head branch does not exist"}`))
	}))
	defer server.Close()

	client, err := NewForgejoClient(SubmissionConfig{ServerURL: server.URL, Token: "secret"})
	if err != nil {
		t.Fatalf("NewForgejoClient() error = %v", err)
	}

	_, err = client.CreatePullRequest(context.Background(), PullRequestRequest{
		Owner:      "acme",
		Repo:       "skillforge",
		HeadBranch: "feature-branch",
		BaseBranch: "main",
		Title:      "test",
	})
	if err == nil {
		t.Fatal("CreatePullRequest() error = nil, want failure")
	}
	want := "forgejo create pull request: response 422 Unprocessable Entity: head branch does not exist"
	if err.Error() != want {
		t.Fatalf("CreatePullRequest() error = %q, want %q", err, want)
	}
}

func TestForgejoClientCreatePullRequestReturnsMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"number":"wrong"}`))
	}))
	defer server.Close()

	client, err := NewForgejoClient(SubmissionConfig{ServerURL: server.URL, AuthMethod: ForgejoAuthMethodNone})
	if err != nil {
		t.Fatalf("NewForgejoClient() error = %v", err)
	}

	_, err = client.CreatePullRequest(context.Background(), PullRequestRequest{
		Owner:      "acme",
		Repo:       "skillforge",
		HeadBranch: "feature-branch",
		BaseBranch: "main",
		Title:      "test",
	})
	if err == nil {
		t.Fatal("CreatePullRequest() error = nil, want decode failure")
	}
	if got, want := err.Error(), "forgejo create pull request: decode response:"; len(got) < len(want) || got[:len(want)] != want {
		t.Fatalf("CreatePullRequest() error = %q, want prefix %q", got, want)
	}
}
