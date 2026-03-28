package draft

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagedGitPublisherMaterializesCreateUpdateAndDelete(t *testing.T) {
	t.Run("create and publish", func(t *testing.T) {
		fixture := newGitFixture(t)
		workspace, err := Manager{RepoRoot: fixture.canonical, NewID: func() string { return "create01" }}.CreateWorkspace("create", "new-skill")
		if err != nil {
			t.Fatalf("CreateWorkspace() error = %v", err)
		}
		content := validSkill("new-skill", "Fresh draft")
		if err := workspace.CreateSkill(content); err != nil {
			t.Fatalf("CreateSkill() error = %v", err)
		}

		publisher := mustNewManagedGitPublisher(t, "main")
		commitHash, err := publisher.Commit(context.Background(), CommitRequest{
			RepoRoot:   workspace.RepoRoot,
			DraftRoot:  workspace.Root,
			BranchName: workspace.BranchName,
			BaseBranch: "main",
			Operation:  workspace.Operation,
			SkillName:  workspace.SkillName,
			Message:    CommitMessage{Subject: "add new skill"},
		})
		if err != nil {
			t.Fatalf("Commit() error = %v", err)
		}
		if strings.TrimSpace(commitHash) == "" {
			t.Fatal("Commit() returned empty hash")
		}
		if err := publisher.Publish(context.Background(), PublishRequest{
			RepoRoot:   workspace.RepoRoot,
			RemoteName: "origin",
			BranchName: workspace.BranchName,
		}); err != nil {
			t.Fatalf("Publish() error = %v", err)
		}

		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "new-skill", "SKILL.md"), "description: Fresh draft")
		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "other-skill", "SKILL.md"), "description: Other skill")
		remoteContent := gitOutput(t, "", "--git-dir", fixture.remote, "show", "refs/heads/"+workspace.BranchName+":skills/new-skill/SKILL.md")
		if !strings.Contains(remoteContent, "description: Fresh draft") {
			t.Fatalf("remote branch does not contain new skill content: %q", remoteContent)
		}
	})

	t.Run("update replaces targeted skill directory only", func(t *testing.T) {
		fixture := newGitFixture(t)
		workspace, err := Manager{RepoRoot: fixture.canonical, NewID: func() string { return "update01" }}.CreateWorkspace("update", "example-skill")
		if err != nil {
			t.Fatalf("CreateWorkspace() error = %v", err)
		}
		if err := workspace.UpdateSkill(validSkill("example-skill", "Updated draft")); err != nil {
			t.Fatalf("UpdateSkill() error = %v", err)
		}
		if err := os.Remove(filepath.Join(workspace.Root, "skills", "example-skill", "notes.txt")); err != nil {
			t.Fatalf("Remove(notes.txt) error = %v", err)
		}
		writeTestFile(t, filepath.Join(workspace.Root, "skills", "example-skill", "extra.md"), "extra draft file\n")

		publisher := mustNewManagedGitPublisher(t, "main")
		if _, err := publisher.Commit(context.Background(), CommitRequest{
			RepoRoot:   workspace.RepoRoot,
			DraftRoot:  workspace.Root,
			BranchName: workspace.BranchName,
			BaseBranch: "main",
			Operation:  workspace.Operation,
			SkillName:  workspace.SkillName,
			Message:    CommitMessage{Subject: "update example skill"},
		}); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}

		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "example-skill", "SKILL.md"), "description: Updated draft")
		if _, err := os.Stat(filepath.Join(fixture.canonical, "skills", "example-skill", "notes.txt")); !os.IsNotExist(err) {
			t.Fatalf("notes.txt should be removed, stat err = %v", err)
		}
		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "example-skill", "extra.md"), "extra draft file")
		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "other-skill", "SKILL.md"), "description: Other skill")
	})

	t.Run("delete removes only targeted skill directory", func(t *testing.T) {
		fixture := newGitFixture(t)
		workspace, err := Manager{RepoRoot: fixture.canonical, NewID: func() string { return "delete01" }}.CreateWorkspace("delete", "example-skill")
		if err != nil {
			t.Fatalf("CreateWorkspace() error = %v", err)
		}
		if err := workspace.DeleteSkill(); err != nil {
			t.Fatalf("DeleteSkill() error = %v", err)
		}

		publisher := mustNewManagedGitPublisher(t, "main")
		if _, err := publisher.Commit(context.Background(), CommitRequest{
			RepoRoot:   workspace.RepoRoot,
			DraftRoot:  workspace.Root,
			BranchName: workspace.BranchName,
			BaseBranch: "main",
			Operation:  workspace.Operation,
			SkillName:  workspace.SkillName,
			Message:    CommitMessage{Subject: "delete example skill"},
		}); err != nil {
			t.Fatalf("Commit() error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(fixture.canonical, "skills", "example-skill")); !os.IsNotExist(err) {
			t.Fatalf("example-skill should be removed, stat err = %v", err)
		}
		assertFileContains(t, filepath.Join(fixture.canonical, "skills", "other-skill", "SKILL.md"), "description: Other skill")
	})
}

func TestManagedGitPublisherBasesBranchFromRequestedBaseBranch(t *testing.T) {
	fixture := newGitFixture(t)
	gitRun(t, fixture.canonical, "checkout", "-b", "release", "origin/release")
	gitRun(t, fixture.canonical, "checkout", "main")

	workspace, err := Manager{RepoRoot: fixture.canonical, NewID: func() string { return "release01" }}.CreateWorkspace("create", "branch-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("branch-skill", "Branch based draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	publisher := mustNewManagedGitPublisher(t, "release")
	if _, err := publisher.Commit(context.Background(), CommitRequest{
		RepoRoot:   workspace.RepoRoot,
		DraftRoot:  workspace.Root,
		BranchName: workspace.BranchName,
		BaseBranch: "release",
		Operation:  workspace.Operation,
		SkillName:  workspace.SkillName,
		Message:    CommitMessage{Subject: "branch from release"},
	}); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	baseHash := strings.TrimSpace(gitOutput(t, fixture.canonical, "rev-parse", "release"))
	parentHash := strings.TrimSpace(gitOutput(t, fixture.canonical, "rev-parse", workspace.BranchName+"^"))
	if parentHash != baseHash {
		t.Fatalf("submission branch parent = %q, want %q", parentHash, baseHash)
	}
	assertFileContains(t, filepath.Join(fixture.canonical, "BASE.txt"), "release base")
}

func TestManagedGitPublisherSurfacesPublishFailure(t *testing.T) {
	fixture := newGitFixture(t)
	workspace, err := Manager{RepoRoot: fixture.canonical, NewID: func() string { return "fail01" }}.CreateWorkspace("create", "new-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("new-skill", "Fresh draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	publisher := mustNewManagedGitPublisher(t, "main")
	if _, err := publisher.Commit(context.Background(), CommitRequest{
		RepoRoot:   workspace.RepoRoot,
		DraftRoot:  workspace.Root,
		BranchName: workspace.BranchName,
		BaseBranch: "main",
		Operation:  workspace.Operation,
		SkillName:  workspace.SkillName,
		Message:    CommitMessage{Subject: "add new skill"},
	}); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	err = publisher.Publish(context.Background(), PublishRequest{
		RepoRoot:   workspace.RepoRoot,
		RemoteName: "missing",
		BranchName: workspace.BranchName,
	})
	if err == nil {
		t.Fatal("Publish() error = nil, want push failure")
	}
	if !strings.Contains(err.Error(), "push branch "+workspace.BranchName+" to missing") {
		t.Fatalf("Publish() error = %v", err)
	}
}

type gitFixture struct {
	remote    string
	canonical string
}

func newGitFixture(t *testing.T) gitFixture {
	t.Helper()
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	seed := filepath.Join(root, "seed")
	canonical := filepath.Join(root, "canonical")

	gitRun(t, "", "init", "--bare", remote)
	gitRun(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/main")
	gitRun(t, "", "init", "-b", "main", seed)

	writeTestFile(t, filepath.Join(seed, "README.md"), "# skillforge\n")
	writeTestFile(t, filepath.Join(seed, "BASE.txt"), "main base\n")
	writeTestFile(t, filepath.Join(seed, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Original"))
	writeTestFile(t, filepath.Join(seed, "skills", "example-skill", "notes.txt"), "base notes\n")
	writeTestFile(t, filepath.Join(seed, "skills", "other-skill", "SKILL.md"), validSkill("other-skill", "Other skill"))
	gitRun(t, seed, "add", ".")
	gitRun(t, seed, "commit", "-m", "seed main")
	gitRun(t, seed, "remote", "add", "origin", remote)
	gitRun(t, seed, "push", "origin", "main")

	gitRun(t, seed, "checkout", "-b", "release")
	writeTestFile(t, filepath.Join(seed, "BASE.txt"), "release base\n")
	gitRun(t, seed, "add", "BASE.txt")
	gitRun(t, seed, "commit", "-m", "seed release")
	gitRun(t, seed, "push", "origin", "release")
	gitRun(t, seed, "checkout", "main")

	gitRun(t, "", "clone", remote, canonical)
	return gitFixture{remote: remote, canonical: canonical}
}

func mustNewManagedGitPublisher(t *testing.T, baseBranch string) *ManagedGitPublisher {
	t.Helper()
	publisher, err := NewManagedGitPublisher(baseBranch)
	if err != nil {
		t.Fatalf("NewManagedGitPublisher() error = %v", err)
	}
	return publisher
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME="+defaultGitAuthorName,
		"GIT_AUTHOR_EMAIL="+defaultGitAuthorEmail,
		"GIT_COMMITTER_NAME="+defaultGitCommitterName,
		"GIT_COMMITTER_EMAIL="+defaultGitCommitterEmail,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output)
}
