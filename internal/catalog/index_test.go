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
	writeFile(t, filepath.Join(repo, "skills", "beta-skill", "SKILL.md"), skillWithTags("beta-skill", "Beta", []string{"pdf"}))

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

func TestIndexSearchMatchesMetadata(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "name-skill", "SKILL.md"), validSkill("name-skill", "Example search target"))
	writeFile(t, filepath.Join(repo, "skills", "description-skill", "SKILL.md"), validSkill("description-skill", "Contains PDF tools"))
	writeFile(t, filepath.Join(repo, "skills", "tag-skill", "SKILL.md"), skillWithTags("tag-skill", "Tag search", []string{"pdf"}))

	index := mustBuildIndex(t, repo)

	if got := index.Search("name"); len(got) != 1 || got[0].Name != "name-skill" {
		t.Fatalf("unexpected name search results: %#v", got)
	}
	if got := index.Search("pdf"); len(got) != 2 {
		t.Fatalf("unexpected pdf search results: %#v", got)
	}
	if got := index.Search(""); len(got) != 0 {
		t.Fatalf("unexpected empty search results: %#v", got)
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

func TestIndexSearchMatchesNameDescriptionAndTags(t *testing.T) {
	repo := tempRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "alpha-skill", "SKILL.md"), richSkill("alpha-skill", "PDF automation helper", []string{"git", "search"}))
	writeFile(t, filepath.Join(repo, "skills", "beta-skill", "SKILL.md"), validSkill("beta-skill", "Other"))

	index := mustBuildIndex(t, repo)

	if got := index.Search("alpha"); len(got) != 1 || got[0].Name != "alpha-skill" {
		t.Fatalf("unexpected name search results: %#v", got)
	}
	if got := index.Search("pdf"); len(got) != 1 || got[0].Name != "alpha-skill" {
		t.Fatalf("unexpected description search results: %#v", got)
	}
	if got := index.Search("git"); len(got) != 1 || got[0].Name != "alpha-skill" {
		t.Fatalf("unexpected tag search results: %#v", got)
	}
	if got := index.Search("missing"); len(got) != 0 {
		t.Fatalf("unexpected empty search results: %#v", got)
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

func richSkill(name, description string, tags []string) string {
	content := "---\nname: " + name + "\ndescription: " + description + "\ntags:\n"
	for _, tag := range tags {
		content += "  - " + tag + "\n"
	}
	content += "---\n# " + name + "\n"
	return content
}

func skillWithTags(name, description string, tags []string) string {
	body := "---\nname: " + name + "\ndescription: " + description + "\ntags:\n"
	for _, tag := range tags {
		body += "  - " + tag + "\n"
	}
	return body + "---\n# " + name + "\n"
}
