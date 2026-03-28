package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScannerDiscoversValidCanonicalSkill(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Example skill"))
	writeFile(t, filepath.Join(repo, "README.md"), "not relevant")

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("len(result.Skills) = %d, want 1", len(result.Skills))
	}
	got := result.Skills[0]
	if !got.Valid {
		t.Fatalf("skill should be valid, findings = %#v", got.Findings)
	}
	if got.Name != "example-skill" {
		t.Fatalf("got name %q, want example-skill", got.Name)
	}
	if got.Description != "Example skill" {
		t.Fatalf("got description %q", got.Description)
	}
	if got.Path != "skills/example-skill/SKILL.md" {
		t.Fatalf("got path %q", got.Path)
	}
	if result.ValidCount != 1 || result.ErrorCount != 0 {
		t.Fatalf("counts = (%d valid, %d errors), want (1,0)", result.ValidCount, result.ErrorCount)
	}
}

func TestScannerIgnoresNonCanonicalSkillMarkdown(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "misc", "SKILL.md"), validSkill("ignored-skill", "Ignored"))
	writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Example skill"))

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("len(result.Skills) = %d, want 1", len(result.Skills))
	}
	if result.Skills[0].Name != "example-skill" {
		t.Fatalf("got %q, want example-skill", result.Skills[0].Name)
	}
}

func TestScannerReportsMalformedFrontmatter(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "broken-skill", "SKILL.md"), "---\nname: broken-skill\ndescription: [oops\n---\nbody")

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 1 {
		t.Fatalf("len(result.Skills) = %d, want 1", len(result.Skills))
	}
	got := result.Skills[0]
	if got.Valid {
		t.Fatal("expected invalid skill")
	}
	assertFindingCode(t, got.Findings, "invalid_frontmatter")
}

func TestScannerReportsNameDirectoryMismatch(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("different-name", "Mismatch"))

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	got := result.Skills[0]
	if got.Valid {
		t.Fatal("expected invalid skill")
	}
	assertFindingCode(t, got.Findings, "name_directory_mismatch")
}

func TestScannerReportsMissingSkillMarkdownDirectory(t *testing.T) {
	repo := tempRepo(t)
	if err := os.MkdirAll(filepath.Join(repo, "skills", "missing-skill"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	got := result.Skills[0]
	if got.Valid {
		t.Fatal("expected invalid skill")
	}
	assertFindingCode(t, got.Findings, "missing_skill_md")
}

func TestScannerPreservesPartialSuccess(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "good-skill", "SKILL.md"), validSkill("good-skill", "Good skill"))
	writeFile(t, filepath.Join(repo, "skills", "bad-skill", "SKILL.md"), validSkill("wrong-name", "Bad skill"))

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(result.Skills) != 2 {
		t.Fatalf("len(result.Skills) = %d, want 2", len(result.Skills))
	}
	var valid, invalid int
	for _, skill := range result.Skills {
		if skill.Valid {
			valid++
		} else {
			invalid++
		}
	}
	if valid != 1 || invalid != 1 {
		t.Fatalf("valid=%d invalid=%d, want 1/1", valid, invalid)
	}
}

func TestScannerExtractsNormalizedMetadata(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "metadata-skill", "SKILL.md"), `---
name: metadata-skill
description: Metadata rich
license: MIT
compatibility:
  - openclaw
  - codex
allowed-tools:
  - read
  - exec
metadata:
  tags:
    - search
    - git
  owner: platform
---
Body text.`)

	result, err := Scanner{Root: repo}.Scan()
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	got := result.Skills[0]
	if got.License != "MIT" {
		t.Fatalf("got license %q", got.License)
	}
	if len(got.Compatibility) != 2 || got.Compatibility[0] != "codex" || got.Compatibility[1] != "openclaw" {
		t.Fatalf("unexpected compatibility: %#v", got.Compatibility)
	}
	if len(got.AllowedTools) != 2 || got.AllowedTools[0] != "exec" || got.AllowedTools[1] != "read" {
		t.Fatalf("unexpected allowed tools: %#v", got.AllowedTools)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "git" || got.Tags[1] != "search" {
		t.Fatalf("unexpected tags: %#v", got.Tags)
	}
	if got.Metadata["owner"] != "platform" {
		t.Fatalf("unexpected metadata owner: %#v", got.Metadata)
	}
}

func tempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func validSkill(name, description string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n# " + name + "\n"
}

func assertFindingCode(t *testing.T, findings []Finding, want string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Code == want {
			return
		}
	}
	t.Fatalf("finding code %q not found in %#v", want, findings)
}
