package draft

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type ForgejoAuthMethod string

const (
	ForgejoAuthMethodToken ForgejoAuthMethod = "token"
	ForgejoAuthMethodNone  ForgejoAuthMethod = "none"
)

type SubmissionConfig struct {
	RemoteName string            `json:"remoteName"`
	Owner      string            `json:"owner"`
	Repo       string            `json:"repo"`
	BaseBranch string            `json:"baseBranch"`
	Token      string            `json:"-"`
	AuthMethod ForgejoAuthMethod `json:"authMethod"`
}

func (c SubmissionConfig) Validate() error {
	if strings.TrimSpace(c.RemoteName) == "" {
		return errors.New("forgejo remote name is required")
	}
	if strings.TrimSpace(c.Owner) == "" {
		return errors.New("forgejo owner is required")
	}
	if strings.TrimSpace(c.Repo) == "" {
		return errors.New("forgejo repo is required")
	}
	if strings.TrimSpace(c.BaseBranch) == "" {
		return errors.New("forgejo base branch is required")
	}

	switch c.authMethod() {
	case ForgejoAuthMethodToken:
		if strings.TrimSpace(c.Token) == "" {
			return errors.New("forgejo token is required for token auth")
		}
	case ForgejoAuthMethodNone:
		return nil
	default:
		return fmt.Errorf("unsupported forgejo auth method %q", c.AuthMethod)
	}

	return nil
}

func (c SubmissionConfig) authMethod() ForgejoAuthMethod {
	if strings.TrimSpace(string(c.AuthMethod)) == "" {
		return ForgejoAuthMethodToken
	}
	return ForgejoAuthMethod(strings.ToLower(strings.TrimSpace(string(c.AuthMethod))))
}

type CommitMessage struct {
	Subject string   `json:"subject"`
	Body    []string `json:"body,omitempty"`
}

func (m CommitMessage) String() string {
	subject := strings.TrimSpace(m.Subject)
	if len(m.Body) == 0 {
		return subject
	}

	body := make([]string, 0, len(m.Body))
	for _, line := range m.Body {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			body = append(body, trimmed)
		}
	}
	if len(body) == 0 {
		return subject
	}
	return subject + "\n\n" + strings.Join(body, "\n")
}

func (m CommitMessage) PullRequestBody() string {
	if len(m.Body) == 0 {
		return ""
	}
	lines := make([]string, 0, len(m.Body))
	for _, line := range m.Body {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return strings.Join(lines, "\n")
}

func GenerateCommitMessage(workspace *Workspace) CommitMessage {
	if workspace == nil {
		return CommitMessage{}
	}
	return CommitMessage{
		Subject: fmt.Sprintf("skillforge: %s %s", workspace.Operation, workspace.SkillName),
		Body: []string{
			fmt.Sprintf("Operation: %s", workspace.Operation),
			fmt.Sprintf("Skill: %s", workspace.SkillName),
			fmt.Sprintf("Branch: %s", workspace.BranchName),
		},
	}
}

type CommitRequest struct {
	RepoRoot   string        `json:"repoRoot"`
	BranchName string        `json:"branchName"`
	Message    CommitMessage `json:"message"`
}

type PublishRequest struct {
	RepoRoot   string `json:"repoRoot"`
	RemoteName string `json:"remoteName"`
	BranchName string `json:"branchName"`
}

type GitPublisher interface {
	Commit(ctx context.Context, req CommitRequest) (string, error)
	Publish(ctx context.Context, req PublishRequest) error
}

type PullRequestRequest struct {
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	HeadBranch string `json:"headBranch"`
	BaseBranch string `json:"baseBranch"`
	Title      string `json:"title"`
	Body       string `json:"body,omitempty"`
}

type PullRequest struct {
	Number int    `json:"number,omitempty"`
	ID     int64  `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
}

type ForgejoClient interface {
	CreatePullRequest(ctx context.Context, req PullRequestRequest) (PullRequest, error)
}

type PullRequestRef struct {
	Number int    `json:"number,omitempty"`
	ID     int64  `json:"id,omitempty"`
	URL    string `json:"url,omitempty"`
}

type SubmissionResult struct {
	BranchName  string           `json:"branchName"`
	BaseBranch  string           `json:"baseBranch"`
	CommitHash  string           `json:"commitHash,omitempty"`
	PullRequest *PullRequestRef  `json:"pullRequest,omitempty"`
	Validation  ValidationResult `json:"validation"`
}

type SubmissionService struct {
	Config  SubmissionConfig
	Git     GitPublisher
	Forgejo ForgejoClient
}

type SubmissionStatus struct {
	Enabled    bool   `json:"enabled"`
	BaseBranch string `json:"baseBranch,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

func (s SubmissionService) Status() SubmissionStatus {
	if s.Git == nil && s.Forgejo == nil && isZeroSubmissionConfig(s.Config) {
		return SubmissionStatus{Enabled: false, Reason: "submission service is not configured"}
	}
	if err := s.Config.Validate(); err != nil {
		return SubmissionStatus{Enabled: false, Reason: fmt.Sprintf("invalid submission config: %v", err)}
	}
	if s.Git == nil {
		return SubmissionStatus{Enabled: false, Reason: "git publisher is not configured"}
	}
	if s.Forgejo == nil {
		return SubmissionStatus{Enabled: false, Reason: "forgejo client is not configured"}
	}
	return SubmissionStatus{Enabled: true, BaseBranch: s.Config.BaseBranch}
}

type ValidationError struct {
	Result ValidationResult
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("draft validation failed with %d finding(s)", len(e.Result.Findings))
}

func (e ValidationError) Unwrap() error {
	return ErrInvalidDraft
}

var ErrInvalidDraft = errors.New("invalid draft")

func isZeroSubmissionConfig(c SubmissionConfig) bool {
	return strings.TrimSpace(c.RemoteName) == "" &&
		strings.TrimSpace(c.Owner) == "" &&
		strings.TrimSpace(c.Repo) == "" &&
		strings.TrimSpace(c.BaseBranch) == "" &&
		strings.TrimSpace(c.Token) == "" &&
		strings.TrimSpace(string(c.AuthMethod)) == ""
}

func (s SubmissionService) Submit(ctx context.Context, workspace *Workspace) (SubmissionResult, error) {
	if workspace == nil {
		return SubmissionResult{}, errors.New("workspace is required")
	}
	status := s.Status()
	if !status.Enabled {
		return SubmissionResult{}, errors.New(status.Reason)
	}

	validation, err := workspace.Validate()
	if err != nil {
		return SubmissionResult{}, fmt.Errorf("validate workspace: %w", err)
	}

	result := SubmissionResult{
		BranchName: workspace.BranchName,
		BaseBranch: s.Config.BaseBranch,
		Validation: validation,
	}
	if !validation.Valid {
		return result, ValidationError{Result: validation}
	}

	message := GenerateCommitMessage(workspace)
	commitHash, err := s.Git.Commit(ctx, CommitRequest{
		RepoRoot:   workspace.Root,
		BranchName: workspace.BranchName,
		Message:    message,
	})
	if err != nil {
		return result, fmt.Errorf("commit draft: %w", err)
	}
	result.CommitHash = strings.TrimSpace(commitHash)

	if err := s.Git.Publish(ctx, PublishRequest{
		RepoRoot:   workspace.Root,
		RemoteName: s.Config.RemoteName,
		BranchName: workspace.BranchName,
	}); err != nil {
		return result, fmt.Errorf("publish draft branch: %w", err)
	}

	pr, err := s.Forgejo.CreatePullRequest(ctx, PullRequestRequest{
		Owner:      s.Config.Owner,
		Repo:       s.Config.Repo,
		HeadBranch: workspace.BranchName,
		BaseBranch: s.Config.BaseBranch,
		Title:      message.Subject,
		Body:       message.PullRequestBody(),
	})
	if err != nil {
		return result, fmt.Errorf("create pull request: %w", err)
	}

	result.PullRequest = &PullRequestRef{
		Number: pr.Number,
		ID:     pr.ID,
		URL:    pr.URL,
	}
	return result, nil
}
