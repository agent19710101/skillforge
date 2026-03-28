package draft

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSubmissionConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		config SubmissionConfig
		want   string
	}{
		{
			name: "missing server url",
			config: SubmissionConfig{
				RemoteName: "origin",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
				Token:      "secret",
			},
			want: "forgejo server URL is required",
		},
		{
			name: "missing remote",
			config: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
				Token:      "secret",
			},
			want: "forgejo remote name is required",
		},
		{
			name: "missing token defaults to token auth",
			config: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				RemoteName: "origin",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
			},
			want: "forgejo token is required for token auth",
		},
		{
			name: "unsupported auth method",
			config: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				RemoteName: "origin",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
				Token:      "secret",
				AuthMethod: "basic",
			},
			want: "unsupported forgejo auth method \"basic\"",
		},
		{
			name: "no auth allowed for tests",
			config: SubmissionConfig{
				ServerURL:  "https://forgejo.example",
				RemoteName: "origin",
				Owner:      "acme",
				Repo:       "skillforge",
				BaseBranch: "main",
				AuthMethod: ForgejoAuthMethodNone,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.want == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.want {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSubmissionServiceStatus(t *testing.T) {
	tests := []struct {
		name    string
		service SubmissionService
		want    SubmissionStatus
	}{
		{
			name:    "unconfigured by default",
			service: SubmissionService{},
			want: SubmissionStatus{
				Enabled: false,
				Reason:  "submission service is not configured",
			},
		},
		{
			name: "missing forgejo client",
			service: SubmissionService{
				Config: testSubmissionConfig(),
				Git:    &fakeGitPublisher{},
			},
			want: SubmissionStatus{
				Enabled: false,
				Reason:  "forgejo client is not configured",
			},
		},
		{
			name: "enabled",
			service: SubmissionService{
				Config:  testSubmissionConfig(),
				Git:     &fakeGitPublisher{},
				Forgejo: &fakeForgejoClient{},
			},
			want: SubmissionStatus{
				Enabled:    true,
				BaseBranch: "main",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.service.Status()
			if got != tt.want {
				t.Fatalf("Status() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSubmitBlocksInvalidDraft(t *testing.T) {
	repo := testRepo(t)
	writeTestFile(t, repo+"/skills/example-skill/SKILL.md", validSkill("example-skill", "Example skill"))

	workspace, err := Manager{RepoRoot: repo, NewID: func() string { return "invalid01" }}.CreateWorkspace("update", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.UpdateSkill(validSkill("wrong-name", "Broken draft")); err != nil {
		t.Fatalf("UpdateSkill() error = %v", err)
	}

	git := &fakeGitPublisher{}
	forgejo := &fakeForgejoClient{}
	service := SubmissionService{
		Config:  testSubmissionConfig(),
		Git:     git,
		Forgejo: forgejo,
	}

	result, err := service.Submit(context.Background(), workspace)
	if err == nil {
		t.Fatal("Submit() error = nil, want validation error")
	}
	if !errors.Is(err, ErrInvalidDraft) {
		t.Fatalf("Submit() error = %v, want ErrInvalidDraft", err)
	}
	var validationErr ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Submit() error = %v, want ValidationError", err)
	}
	if validationErr.Result.Valid {
		t.Fatal("validation error reported a valid draft")
	}
	assertFindingCode(t, result.Validation.Findings, "name_directory_mismatch")
	if git.commitCalls != 0 || git.publishCalls != 0 {
		t.Fatalf("git should not run for invalid draft, got commit=%d publish=%d", git.commitCalls, git.publishCalls)
	}
	if forgejo.createCalls != 0 {
		t.Fatalf("forgejo should not run for invalid draft, got create=%d", forgejo.createCalls)
	}
}

func TestSubmitSuccess(t *testing.T) {
	repo := testRepo(t)
	workspace, err := Manager{RepoRoot: repo, NewID: func() string { return "submit01" }}.CreateWorkspace("create", "new-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("new-skill", "Fresh draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	git := &fakeGitPublisher{commitHash: "abc123"}
	forgejo := &fakeForgejoClient{pullRequest: PullRequest{Number: 17, ID: 42, URL: "https://forgejo.example/acme/skillforge/pulls/17"}}
	service := SubmissionService{
		Config:  testSubmissionConfig(),
		Git:     git,
		Forgejo: forgejo,
	}

	result, err := service.Submit(context.Background(), workspace)
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if result.BranchName != workspace.BranchName {
		t.Fatalf("result.BranchName = %q, want %q", result.BranchName, workspace.BranchName)
	}
	if result.BaseBranch != "main" {
		t.Fatalf("result.BaseBranch = %q, want main", result.BaseBranch)
	}
	if result.CommitHash != "abc123" {
		t.Fatalf("result.CommitHash = %q, want abc123", result.CommitHash)
	}
	if result.PullRequest == nil {
		t.Fatal("result.PullRequest = nil")
	}
	if result.PullRequest.Number != 17 || result.PullRequest.ID != 42 || result.PullRequest.URL == "" {
		t.Fatalf("unexpected pull request result: %#v", result.PullRequest)
	}
	if !result.Validation.Valid {
		t.Fatalf("expected valid submission result, findings = %#v", result.Validation.Findings)
	}
	if git.commitCalls != 1 || git.publishCalls != 1 {
		t.Fatalf("git calls = commit:%d publish:%d, want 1 each", git.commitCalls, git.publishCalls)
	}
	if git.lastCommit.RepoRoot != workspace.RepoRoot || git.lastCommit.DraftRoot != workspace.Root || git.lastCommit.BranchName != workspace.BranchName {
		t.Fatalf("unexpected commit request: %#v", git.lastCommit)
	}
	if git.lastCommit.BaseBranch != "main" || git.lastCommit.Operation != "create" || git.lastCommit.SkillName != "new-skill" {
		t.Fatalf("unexpected commit metadata: %#v", git.lastCommit)
	}
	if git.lastCommit.Message.Subject != "skillforge: create new-skill" {
		t.Fatalf("commit subject = %q", git.lastCommit.Message.Subject)
	}
	if got := git.lastCommit.Message.String(); !strings.Contains(got, "Operation: create") || !strings.Contains(got, "Skill: new-skill") {
		t.Fatalf("commit message = %q", got)
	}
	if git.lastPublish.RemoteName != "origin" || git.lastPublish.BranchName != workspace.BranchName {
		t.Fatalf("unexpected publish request: %#v", git.lastPublish)
	}
	if forgejo.createCalls != 1 {
		t.Fatalf("forgejo create calls = %d, want 1", forgejo.createCalls)
	}
	if forgejo.lastRequest.Owner != "acme" || forgejo.lastRequest.Repo != "skillforge" {
		t.Fatalf("unexpected forgejo request: %#v", forgejo.lastRequest)
	}
	if forgejo.lastRequest.HeadBranch != workspace.BranchName || forgejo.lastRequest.BaseBranch != "main" {
		t.Fatalf("unexpected forgejo branch request: %#v", forgejo.lastRequest)
	}
	if forgejo.lastRequest.Title != "skillforge: create new-skill" {
		t.Fatalf("pull request title = %q", forgejo.lastRequest.Title)
	}
}

func TestSubmitReturnsPublishFailure(t *testing.T) {
	repo := testRepo(t)
	workspace, err := Manager{RepoRoot: repo, NewID: func() string { return "publish01" }}.CreateWorkspace("create", "new-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("new-skill", "Fresh draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	git := &fakeGitPublisher{commitHash: "abc123", publishErr: errors.New("push failed")}
	forgejo := &fakeForgejoClient{}
	service := SubmissionService{
		Config:  testSubmissionConfig(),
		Git:     git,
		Forgejo: forgejo,
	}

	result, err := service.Submit(context.Background(), workspace)
	if err == nil || !strings.Contains(err.Error(), "publish draft branch") {
		t.Fatalf("Submit() error = %v, want publish failure", err)
	}
	if result.CommitHash != "abc123" {
		t.Fatalf("result.CommitHash = %q, want abc123", result.CommitHash)
	}
	if forgejo.createCalls != 0 {
		t.Fatalf("forgejo should not run after publish failure, got create=%d", forgejo.createCalls)
	}
}

func TestSubmitReturnsPullRequestFailure(t *testing.T) {
	repo := testRepo(t)
	workspace, err := Manager{RepoRoot: repo, NewID: func() string { return "prfail01" }}.CreateWorkspace("create", "new-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("new-skill", "Fresh draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	git := &fakeGitPublisher{commitHash: "abc123"}
	forgejo := &fakeForgejoClient{err: errors.New("api unavailable")}
	service := SubmissionService{
		Config:  testSubmissionConfig(),
		Git:     git,
		Forgejo: forgejo,
	}

	result, err := service.Submit(context.Background(), workspace)
	if err == nil || !strings.Contains(err.Error(), "create pull request") {
		t.Fatalf("Submit() error = %v, want pull request failure", err)
	}
	if result.CommitHash != "abc123" {
		t.Fatalf("result.CommitHash = %q, want abc123", result.CommitHash)
	}
	if result.PullRequest != nil {
		t.Fatalf("result.PullRequest = %#v, want nil", result.PullRequest)
	}
	if forgejo.createCalls != 1 {
		t.Fatalf("forgejo create calls = %d, want 1", forgejo.createCalls)
	}
}

func testSubmissionConfig() SubmissionConfig {
	return SubmissionConfig{
		ServerURL:  "https://forgejo.example",
		RemoteName: "origin",
		Owner:      "acme",
		Repo:       "skillforge",
		BaseBranch: "main",
		Token:      "secret",
	}
}

type fakeGitPublisher struct {
	commitHash   string
	commitErr    error
	publishErr   error
	commitCalls  int
	publishCalls int
	lastCommit   CommitRequest
	lastPublish  PublishRequest
}

func (f *fakeGitPublisher) Commit(_ context.Context, req CommitRequest) (string, error) {
	f.commitCalls++
	f.lastCommit = req
	if f.commitErr != nil {
		return "", f.commitErr
	}
	return f.commitHash, nil
}

func (f *fakeGitPublisher) Publish(_ context.Context, req PublishRequest) error {
	f.publishCalls++
	f.lastPublish = req
	return f.publishErr
}

type fakeForgejoClient struct {
	pullRequest PullRequest
	err         error
	createCalls int
	lastRequest PullRequestRequest
}

func (f *fakeForgejoClient) CreatePullRequest(_ context.Context, req PullRequestRequest) (PullRequest, error) {
	f.createCalls++
	f.lastRequest = req
	if f.err != nil {
		return PullRequest{}, f.err
	}
	return f.pullRequest, nil
}
