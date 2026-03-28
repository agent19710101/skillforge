package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/agent19710101/skillforge/internal/api"
	"github.com/agent19710101/skillforge/internal/catalog"
	"github.com/agent19710101/skillforge/internal/draft"
)

const (
	envListenAddr        = "SKILLFORGE_LISTEN_ADDR"
	envRepoRoot          = "SKILLFORGE_REPO_ROOT"
	envForgejoServerURL  = "SKILLFORGE_FORGEJO_SERVER_URL"
	envForgejoRemoteName = "SKILLFORGE_FORGEJO_REMOTE_NAME"
	envForgejoOwner      = "SKILLFORGE_FORGEJO_OWNER"
	envForgejoRepo       = "SKILLFORGE_FORGEJO_REPO"
	envForgejoBaseBranch = "SKILLFORGE_FORGEJO_BASE_BRANCH"
	envForgejoToken      = "SKILLFORGE_FORGEJO_TOKEN"
	envForgejoAuthMethod = "SKILLFORGE_FORGEJO_AUTH_METHOD"
)

var (
	buildGitPublisher = func(cfg draft.SubmissionConfig) (draft.GitPublisher, error) {
		return draft.NewManagedGitPublisher(cfg.BaseBranch)
	}
	buildForgejoClient = func(cfg draft.SubmissionConfig) (draft.ForgejoClient, error) {
		return draft.NewForgejoClient(cfg)
	}
)

func main() {
	addr := envOrDefault(envListenAddr, ":8080")
	repoRoot := envOrDefault(envRepoRoot, ".")

	index, err := catalog.BuildIndex(repoRoot)
	if err != nil {
		log.Fatalf("build catalog index: %v", err)
	}

	submission, err := submissionServiceFromEnv(os.Getenv)
	if err != nil {
		log.Fatalf("build submission service: %v", err)
	}
	drafts := draft.NewService(draft.Manager{RepoRoot: repoRoot}, submission)
	handler := api.NewServer(index, api.WithDraftService(drafts)).Handler()
	log.Printf("skillforge-api listening on %s (repo root: %s, submission enabled: %t)", addr, repoRoot, submission.Status().Enabled)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func submissionServiceFromEnv(getenv func(string) string) (draft.SubmissionService, error) {
	cfg, configured := submissionConfigFromEnv(getenv)
	if !configured {
		return draft.SubmissionService{}, nil
	}

	service := draft.SubmissionService{Config: cfg}
	if err := cfg.Validate(); err != nil {
		return service, nil
	}

	gitPublisher, err := buildGitPublisher(cfg)
	if err != nil {
		return service, fmt.Errorf("build git publisher: %w", err)
	}
	forgejoClient, err := buildForgejoClient(cfg)
	if err != nil {
		return service, fmt.Errorf("build forgejo client: %w", err)
	}
	service.Git = gitPublisher
	service.Forgejo = forgejoClient
	return service, nil
}

func submissionConfigFromEnv(getenv func(string) string) (draft.SubmissionConfig, bool) {
	cfg := draft.SubmissionConfig{
		ServerURL:  trimmedEnv(getenv, envForgejoServerURL),
		RemoteName: trimmedEnv(getenv, envForgejoRemoteName),
		Owner:      trimmedEnv(getenv, envForgejoOwner),
		Repo:       trimmedEnv(getenv, envForgejoRepo),
		BaseBranch: trimmedEnv(getenv, envForgejoBaseBranch),
		Token:      trimmedEnv(getenv, envForgejoToken),
		AuthMethod: draft.ForgejoAuthMethod(trimmedEnv(getenv, envForgejoAuthMethod)),
	}
	return cfg, !isZeroSubmissionEnvConfig(cfg)
}

func isZeroSubmissionEnvConfig(cfg draft.SubmissionConfig) bool {
	return strings.TrimSpace(cfg.ServerURL) == "" &&
		strings.TrimSpace(cfg.RemoteName) == "" &&
		strings.TrimSpace(cfg.Owner) == "" &&
		strings.TrimSpace(cfg.Repo) == "" &&
		strings.TrimSpace(cfg.BaseBranch) == "" &&
		strings.TrimSpace(cfg.Token) == "" &&
		strings.TrimSpace(string(cfg.AuthMethod)) == ""
}

func trimmedEnv(getenv func(string) string, key string) string {
	return strings.TrimSpace(getenv(key))
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
