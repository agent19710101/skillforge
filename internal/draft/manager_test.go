package draft

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agent19710101/skillforge/internal/catalog"
)

func TestCreateWorkspaceCopiesRepositoryAndAssignsBranchName(t *testing.T) {
	repo := testRepo(t)
	writeTestFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Example skill"))
	writeTestFile(t, filepath.Join(repo, "README.md"), "# skillforge")
	writeTestFile(t, filepath.Join(repo, ".git", "HEAD"), "ref: refs/heads/main\n")

	manager := Manager{
		RepoRoot: repo,
		NewID:    func() string { return "abc123" },
	}
	workspace, err := manager.CreateWorkspace("update", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	if workspace.BranchName != "skillforge/update/example-skill/abc123" {
		t.Fatalf("branch = %q", workspace.BranchName)
	}
	assertFileContains(t, filepath.Join(workspace.Root, "README.md"), "# skillforge")
	assertFileContains(t, filepath.Join(workspace.Root, "skills", "example-skill", "SKILL.md"), "name: example-skill")
	if _, err := os.Stat(filepath.Join(workspace.Root, ".git")); !os.IsNotExist(err) {
		t.Fatalf("workspace should not copy .git directory, stat err = %v", err)
	}
}

func TestCreateSkillAndValidateDraft(t *testing.T) {
	repo := testRepo(t)
	manager := Manager{
		RepoRoot: repo,
		NewID:    func() string { return "create01" },
	}
	workspace, err := manager.CreateWorkspace("create", "new-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.CreateSkill(validSkill("new-skill", "Fresh draft")); err != nil {
		t.Fatalf("CreateSkill() error = %v", err)
	}

	validation, err := workspace.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !validation.Valid {
		t.Fatalf("validation.Valid = false, findings = %#v", validation.Findings)
	}
	if len(validation.Scan.Skills) != 1 || validation.Scan.Skills[0].Name != "new-skill" {
		t.Fatalf("unexpected scan result: %#v", validation.Scan.Skills)
	}
}

func TestUpdateSkillCanProduceInvalidDraftDetectedByValidation(t *testing.T) {
	repo := testRepo(t)
	writeTestFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Example skill"))

	manager := Manager{
		RepoRoot: repo,
		NewID:    func() string { return "update01" },
	}
	workspace, err := manager.CreateWorkspace("update", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.UpdateSkill(validSkill("wrong-name", "Broken draft")); err != nil {
		t.Fatalf("UpdateSkill() error = %v", err)
	}

	validation, err := workspace.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validation.Valid {
		t.Fatal("expected invalid validation result")
	}
	assertFindingCode(t, validation.Findings, "name_directory_mismatch")
}

func TestDeleteSkillRemovesSkillDirectory(t *testing.T) {
	repo := testRepo(t)
	writeTestFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Example skill"))
	writeTestFile(t, filepath.Join(repo, "skills", "other-skill", "SKILL.md"), validSkill("other-skill", "Other skill"))

	manager := Manager{
		RepoRoot: repo,
		NewID:    func() string { return "delete01" },
	}
	workspace, err := manager.CreateWorkspace("delete", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if err := workspace.DeleteSkill(); err != nil {
		t.Fatalf("DeleteSkill() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(workspace.Root, "skills", "example-skill")); !os.IsNotExist(err) {
		t.Fatalf("deleted skill dir still present, stat err = %v", err)
	}
	validation, err := workspace.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !validation.Valid {
		t.Fatalf("expected remaining draft to validate, findings = %#v", validation.Findings)
	}
	if len(validation.Scan.Skills) != 1 || validation.Scan.Skills[0].Name != "other-skill" {
		t.Fatalf("unexpected remaining skills: %#v", validation.Scan.Skills)
	}
}

func TestDraftWorkspacesAreIsolated(t *testing.T) {
	repo := testRepo(t)
	original := validSkill("example-skill", "Original")
	writeTestFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), original)

	idValues := []string{"one", "two"}
	manager := Manager{
		RepoRoot: repo,
		NewID: func() string {
			id := idValues[0]
			idValues = idValues[1:]
			return id
		},
	}

	first, err := manager.CreateWorkspace("update", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace(first) error = %v", err)
	}
	second, err := manager.CreateWorkspace("update", "example-skill")
	if err != nil {
		t.Fatalf("CreateWorkspace(second) error = %v", err)
	}

	if err := first.UpdateSkill(validSkill("example-skill", "Changed only in first workspace")); err != nil {
		t.Fatalf("UpdateSkill(first) error = %v", err)
	}

	assertFileContains(t, filepath.Join(first.Root, "skills", "example-skill", "SKILL.md"), "Changed only in first workspace")
	assertFileContains(t, filepath.Join(second.Root, "skills", "example-skill", "SKILL.md"), "description: Original")
	assertFileContains(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), "description: Original")
}

func TestCreateWorkspaceRejectsInvalidOperationAndSkillName(t *testing.T) {
	repo := testRepo(t)
	manager := Manager{RepoRoot: repo}
	if _, err := manager.CreateWorkspace("publish", "example-skill"); err == nil {
		t.Fatal("expected invalid operation error")
	}
	if _, err := manager.CreateWorkspace("create", "Not Valid"); err == nil {
		t.Fatal("expected invalid skill name error")
	}
}

func testRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	return dir
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("file %q does not contain %q; got %q", path, want, string(content))
	}
}

func validSkill(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n# " + name + "\n"
}

func assertFindingCode(t *testing.T, findings []catalog.Finding, want string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Code == want {
			return
		}
	}
	t.Fatalf("finding code %q not found in %#v", want, findings)
}
