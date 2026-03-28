package draft

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agent19710101/skillforge/internal/catalog"
)

const (
	defaultGitBinary         = "git"
	defaultGitAuthorName     = "skillforge"
	defaultGitAuthorEmail    = "skillforge@example.invalid"
	defaultGitCommitterName  = defaultGitAuthorName
	defaultGitCommitterEmail = defaultGitAuthorEmail
)

type ManagedGitPublisher struct {
	BaseBranch string
	GitBinary  string
}

func NewManagedGitPublisher(baseBranch string) (*ManagedGitPublisher, error) {
	if strings.TrimSpace(baseBranch) == "" {
		return nil, errors.New("base branch is required")
	}
	return &ManagedGitPublisher{BaseBranch: strings.TrimSpace(baseBranch)}, nil
}

func (p *ManagedGitPublisher) Commit(ctx context.Context, req CommitRequest) (string, error) {
	repoRoot, err := filepath.Abs(strings.TrimSpace(req.RepoRoot))
	if err != nil {
		return "", fmt.Errorf("resolve canonical repo root: %w", err)
	}
	draftRoot, err := filepath.Abs(strings.TrimSpace(req.DraftRoot))
	if err != nil {
		return "", fmt.Errorf("resolve draft root: %w", err)
	}
	remoteName := strings.TrimSpace(req.RemoteName)
	if remoteName == "" {
		return "", errors.New("remote name is required")
	}
	branchName := strings.TrimSpace(req.BranchName)
	if branchName == "" {
		return "", errors.New("branch name is required")
	}
	baseBranch := strings.TrimSpace(req.BaseBranch)
	if baseBranch == "" {
		baseBranch = strings.TrimSpace(p.BaseBranch)
	}
	if baseBranch == "" {
		return "", errors.New("base branch is required")
	}
	operation := strings.TrimSpace(req.Operation)
	if err := validateOperation(operation); err != nil {
		return "", err
	}
	skillName := strings.TrimSpace(req.SkillName)
	if !catalog.IsCanonicalSkillName(skillName) {
		return "", fmt.Errorf("invalid skill name %q", req.SkillName)
	}
	if strings.TrimSpace(req.Message.Subject) == "" {
		return "", errors.New("commit message subject is required")
	}

	if err := p.resetBranch(ctx, repoRoot, remoteName, baseBranch, branchName); err != nil {
		return "", err
	}
	if err := materializeSkillChange(draftRoot, repoRoot, operation, skillName); err != nil {
		return "", err
	}

	skillPath := filepath.ToSlash(filepath.Join("skills", skillName))
	if err := p.runGit(ctx, repoRoot, "add", "-A", "--", skillPath); err != nil {
		return "", fmt.Errorf("stage skill changes: %w", err)
	}
	changed, err := p.hasStagedChanges(ctx, repoRoot, skillPath)
	if err != nil {
		return "", err
	}
	if !changed {
		return "", fmt.Errorf("no skill changes staged for %s", skillPath)
	}

	messageFile, err := writeCommitMessageFile(repoRoot, req.Message.String())
	if err != nil {
		return "", err
	}
	defer os.Remove(messageFile)

	if err := p.runGit(ctx, repoRoot, "commit", "-F", messageFile); err != nil {
		return "", fmt.Errorf("commit skill changes: %w", err)
	}
	commitHash, err := p.outputGit(ctx, repoRoot, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolve commit hash: %w", err)
	}
	return strings.TrimSpace(commitHash), nil
}

func (p *ManagedGitPublisher) Publish(ctx context.Context, req PublishRequest) error {
	repoRoot, err := filepath.Abs(strings.TrimSpace(req.RepoRoot))
	if err != nil {
		return fmt.Errorf("resolve canonical repo root: %w", err)
	}
	remoteName := strings.TrimSpace(req.RemoteName)
	if remoteName == "" {
		return errors.New("remote name is required")
	}
	branchName := strings.TrimSpace(req.BranchName)
	if branchName == "" {
		return errors.New("branch name is required")
	}
	if err := p.runGit(ctx, repoRoot, "push", "--force", remoteName, branchName+":refs/heads/"+branchName); err != nil {
		return fmt.Errorf("push branch %s to %s: %w", branchName, remoteName, err)
	}
	return nil
}

func (p *ManagedGitPublisher) resetBranch(ctx context.Context, repoRoot, remoteName, baseBranch, branchName string) error {
	if err := p.runGit(ctx, repoRoot, "reset", "--hard"); err != nil {
		return fmt.Errorf("reset working copy: %w", err)
	}
	if err := p.runGit(ctx, repoRoot, "clean", "-fd"); err != nil {
		return fmt.Errorf("clean working copy: %w", err)
	}
	remoteBaseRef := remoteName + "/" + baseBranch
	if err := p.runGit(ctx, repoRoot, "fetch", "--prune", remoteName, baseBranch); err != nil {
		return fmt.Errorf("fetch base branch %s from %s: %w", baseBranch, remoteName, err)
	}
	if err := p.runGit(ctx, repoRoot, "checkout", "-B", branchName, remoteBaseRef); err != nil {
		return fmt.Errorf("checkout branch %s from %s: %w", branchName, remoteBaseRef, err)
	}
	if err := p.runGit(ctx, repoRoot, "reset", "--hard", remoteBaseRef); err != nil {
		return fmt.Errorf("reset branch %s to %s: %w", branchName, remoteBaseRef, err)
	}
	if err := p.runGit(ctx, repoRoot, "clean", "-fd"); err != nil {
		return fmt.Errorf("clean branch %s: %w", branchName, err)
	}
	return nil
}

func (p *ManagedGitPublisher) hasStagedChanges(ctx context.Context, repoRoot, path string) (bool, error) {
	_, err := p.outputGit(ctx, repoRoot, "diff", "--cached", "--quiet", "--", path)
	if err == nil {
		return false, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true, nil
	}
	return false, fmt.Errorf("check staged diff for %s: %w", path, err)
}

func materializeSkillChange(draftRoot, repoRoot, operation, skillName string) error {
	sourceDir := filepath.Join(draftRoot, "skills", skillName)
	targetDir := filepath.Join(repoRoot, "skills", skillName)

	switch operation {
	case "create", "update":
		info, err := os.Stat(sourceDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("draft skill directory %s does not exist", sourceDir)
			}
			return fmt.Errorf("stat draft skill directory %s: %w", sourceDir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("draft skill path %s is not a directory", sourceDir)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("reset target skill directory %s: %w", targetDir, err)
		}
		if err := copyRepoTree(sourceDir, targetDir); err != nil {
			return fmt.Errorf("copy draft skill directory %s to %s: %w", sourceDir, targetDir, err)
		}
	case "delete":
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("remove target skill directory %s: %w", targetDir, err)
		}
	default:
		return fmt.Errorf("unsupported draft operation %q", operation)
	}
	return nil
}

func writeCommitMessageFile(repoRoot, message string) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", errors.New("commit message must not be empty")
	}
	file, err := os.CreateTemp(repoRoot, "skillforge-commit-*.txt")
	if err != nil {
		return "", fmt.Errorf("create commit message file: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(message + "\n"); err != nil {
		return "", fmt.Errorf("write commit message file: %w", err)
	}
	return file.Name(), nil
}

func (p *ManagedGitPublisher) runGit(ctx context.Context, repoRoot string, args ...string) error {
	_, err := p.outputGit(ctx, repoRoot, args...)
	return err
}

func (p *ManagedGitPublisher) outputGit(ctx context.Context, repoRoot string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, p.gitBinary(), args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+defaultGitAuthorName,
		"GIT_AUTHOR_EMAIL="+defaultGitAuthorEmail,
		"GIT_COMMITTER_NAME="+defaultGitCommitterName,
		"GIT_COMMITTER_EMAIL="+defaultGitCommitterEmail,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, trimmed)
	}
	return string(output), nil
}

func (p *ManagedGitPublisher) gitBinary() string {
	if strings.TrimSpace(p.GitBinary) == "" {
		return defaultGitBinary
	}
	return strings.TrimSpace(p.GitBinary)
}
