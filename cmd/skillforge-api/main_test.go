package main

import (
	"context"
	"errors"
	"testing"

	"github.com/agent19710101/skillforge/internal/draft"
)

func TestSubmissionConfigFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		configured bool
		want       draft.SubmissionConfig
	}{
		{
			name:       "empty env",
			configured: false,
			want:       draft.SubmissionConfig{},
		},
		{
			name: "trimmed configured values",
			env: map[string]string{
				envForgejoServerURL:  " https://forgejo.example ",
				envForgejoRemoteName: " origin ",
				envForgejoOwner:      " acme ",
				envForgejoRepo:       " skillforge ",
				envForgejoBaseBranch: " main ",
				envForgejoToken:      " secret ",
			},
			configured: true,
			want: draft.SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				RemoteName: "origin",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
				Token:      "secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, configured := submissionConfigFromEnv(mapGet(tt.env))
			if configured != tt.configured {
				t.Fatalf("configured = %v, want %v", configured, tt.configured)
			}
			if cfg != tt.want {
				t.Fatalf("cfg = %#v, want %#v", cfg, tt.want)
			}
		})
	}
}

func TestSubmissionServiceFromEnv(t *testing.T) {
	origGitBuilder := buildGitPublisher
	origForgejoBuilder := buildForgejoClient
	t.Cleanup(func() {
		buildGitPublisher = origGitBuilder
		buildForgejoClient = origForgejoBuilder
	})

	validEnv := map[string]string{
		envForgejoServerURL:  "https://forgejo.example",
		envForgejoRemoteName: "origin",
		envForgejoOwner:      "acme",
		envForgejoRepo:       "skillforge",
		envForgejoBaseBranch: "main",
		envForgejoToken:      "secret",
	}

	t.Run("unconfigured when env is empty", func(t *testing.T) {
		buildGitPublisher = func(draft.SubmissionConfig) (draft.GitPublisher, error) {
			t.Fatal("git builder should not run")
			return nil, nil
		}
		buildForgejoClient = func(draft.SubmissionConfig) (draft.ForgejoClient, error) {
			t.Fatal("forgejo builder should not run")
			return nil, nil
		}

		service, err := submissionServiceFromEnv(mapGet(nil))
		if err != nil {
			t.Fatalf("submissionServiceFromEnv() error = %v", err)
		}
		status := service.Status()
		if status.Enabled || status.Reason != "submission service is not configured" {
			t.Fatalf("status = %#v", status)
		}
	})

	t.Run("partial config reports invalid submission config", func(t *testing.T) {
		buildGitPublisher = func(draft.SubmissionConfig) (draft.GitPublisher, error) {
			t.Fatal("git builder should not run for invalid config")
			return nil, nil
		}
		buildForgejoClient = func(draft.SubmissionConfig) (draft.ForgejoClient, error) {
			t.Fatal("forgejo builder should not run for invalid config")
			return nil, nil
		}

		service, err := submissionServiceFromEnv(mapGet(map[string]string{envForgejoOwner: "acme"}))
		if err != nil {
			t.Fatalf("submissionServiceFromEnv() error = %v", err)
		}
		status := service.Status()
		if status.Enabled || status.Reason == "" || status.Reason == "submission service is not configured" {
			t.Fatalf("status = %#v", status)
		}
	})

	t.Run("complete config builds dependencies and enables submission", func(t *testing.T) {
		var gitBuilds, forgejoBuilds int
		buildGitPublisher = func(cfg draft.SubmissionConfig) (draft.GitPublisher, error) {
			gitBuilds++
			if cfg.ServerURL != "https://forgejo.example" || cfg.BaseBranch != "main" {
				t.Fatalf("unexpected git cfg: %#v", cfg)
			}
			return fakeGitPublisher{}, nil
		}
		buildForgejoClient = func(cfg draft.SubmissionConfig) (draft.ForgejoClient, error) {
			forgejoBuilds++
			if cfg.Owner != "acme" || cfg.Repo != "skillforge" {
				t.Fatalf("unexpected forgejo cfg: %#v", cfg)
			}
			return fakeForgejoClient{}, nil
		}

		service, err := submissionServiceFromEnv(mapGet(validEnv))
		if err != nil {
			t.Fatalf("submissionServiceFromEnv() error = %v", err)
		}
		if gitBuilds != 1 || forgejoBuilds != 1 {
			t.Fatalf("builder calls = git:%d forgejo:%d, want 1 each", gitBuilds, forgejoBuilds)
		}
		status := service.Status()
		if !status.Enabled || status.BaseBranch != "main" {
			t.Fatalf("status = %#v", status)
		}
	})

	t.Run("builder error surfaces on complete config", func(t *testing.T) {
		buildGitPublisher = func(draft.SubmissionConfig) (draft.GitPublisher, error) {
			return nil, errors.New("boom")
		}
		buildForgejoClient = func(draft.SubmissionConfig) (draft.ForgejoClient, error) {
			return fakeForgejoClient{}, nil
		}

		_, err := submissionServiceFromEnv(mapGet(validEnv))
		if err == nil || err.Error() != "build git publisher: boom" {
			t.Fatalf("submissionServiceFromEnv() error = %v", err)
		}
	})
}

func TestDefaultBuildForgejoClient(t *testing.T) {
	client, err := buildForgejoClient(draft.SubmissionConfig{
		ServerURL:  "https://forgejo.example",
		RemoteName: "origin",
		Owner:      "acme",
		Repo:       "skillforge",
		BaseBranch: "main",
		Token:      "secret",
	})
	if err != nil {
		t.Fatalf("buildForgejoClient() error = %v", err)
	}
	if client == nil {
		t.Fatal("buildForgejoClient() returned nil client")
	}
}

func mapGet(values map[string]string) func(string) string {
	return func(key string) string {
		if values == nil {
			return ""
		}
		return values[key]
	}
}

type fakeGitPublisher struct{}

func (fakeGitPublisher) Commit(_ context.Context, _ draft.CommitRequest) (string, error) {
	return "", nil
}
func (fakeGitPublisher) Publish(_ context.Context, _ draft.PublishRequest) error { return nil }

type fakeForgejoClient struct{}

func (fakeForgejoClient) CreatePullRequest(_ context.Context, _ draft.PullRequestRequest) (draft.PullRequest, error) {
	return draft.PullRequest{}, nil
}
