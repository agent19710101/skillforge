package catalog

import (
	"path/filepath"
	"testing"
)

func TestIndexListFiltersAndPaginates(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "alpha-skill", "SKILL.md"), validSkill("alpha-skill", "Alpha"))
	writeFile(t, filepath.Join(repo, "skills", "broken-skill", "SKILL.md"), validSkill("wrong-name", "Broken"))

	index := mustBuildIndex(t, repo)

	valid := index.List(ListOptions{Validation: "valid"})
	if len(valid) != 1 || valid[0].Name != "alpha-skill" {
		t.Fatalf("unexpected valid results: %#v", valid)
	}

	invalid := index.List(ListOptions{Validation: "invalid"})
	if len(invalid) != 1 || invalid[0].Name != "wrong-name" {
		t.Fatalf("unexpected invalid results: %#v", invalid)
	}

	paged := index.List(ListOptions{Offset: 1, Limit: 1})
	if len(paged) != 1 {
		t.Fatalf("unexpected paged results: %#v", paged)
	}
}

func TestIndexGetByName(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "alpha-skill", "SKILL.md"), validSkill("alpha-skill", "Alpha"))

	index := mustBuildIndex(t, repo)
	got, ok := index.Get("alpha-skill")
	if !ok {
		t.Fatal("expected skill lookup to succeed")
	}
	if got.Name != "alpha-skill" {
		t.Fatalf("got %#v", got)
	}
	if _, ok := index.Get("missing-skill"); ok {
		t.Fatal("unexpected skill lookup success")
	}
}

func TestIndexStatusReflectsScanResult(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "alpha-skill", "SKILL.md"), validSkill("alpha-skill", "Alpha"))
	writeFile(t, filepath.Join(repo, "skills", "broken-skill", "SKILL.md"), validSkill("wrong-name", "Broken"))

	index := mustBuildIndex(t, repo)
	status := index.Status()
	if status.SkillCount != 2 || status.ValidCount != 1 || status.ErrorCount != 1 {
		t.Fatalf("unexpected status: %#v", status)
	}
}

func TestParseListOptions(t *testing.T) {
	opts, err := ParseListOptions("valid", "5", "10")
	if err != nil {
		t.Fatalf("ParseListOptions() error = %v", err)
	}
	if opts.Validation != "valid" || opts.Offset != 5 || opts.Limit != 10 {
		t.Fatalf("unexpected options: %#v", opts)
	}
	if _, err := ParseListOptions("", "-1", ""); err == nil {
		t.Fatal("expected invalid offset error")
	}
	if _, err := ParseListOptions("", "", "-1"); err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func mustBuildIndex(t *testing.T, root string) *Index {
	t.Helper()
	index, err := BuildIndex(root)
	if err != nil {
		t.Fatalf("BuildIndex() error = %v", err)
	}
	return index
}
