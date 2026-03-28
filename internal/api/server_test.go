package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/agent19710101/skillforge/internal/catalog"
)

func TestListSkillsEndpoint(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills?validation=valid", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body struct {
		Skills []catalog.SkillRecord `json:"skills"`
		Total  int                   `json:"total"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Total != 1 || len(body.Skills) != 1 || body.Skills[0].Name != "alpha-skill" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestListSkillsRejectsInvalidQuery(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills?offset=-1", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestGetSkillByNameEndpoint(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills/alpha-skill", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body catalog.SkillRecord
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Name != "alpha-skill" {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestGetSkillByNameReturnsNotFound(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/skills/missing-skill", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func TestIndexStatusEndpoint(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/index/status", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
	var body catalog.IndexStatus
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.SkillCount != 2 || body.ValidCount != 1 || body.ErrorCount != 1 {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestSearchEndpointMatchesNameDescriptionAndTags(t *testing.T) {
	h := testHandler(t)
	tests := []struct {
		name  string
		query string
	}{
		{name: "name", query: "alpha"},
		{name: "description", query: "pdf"},
		{name: "tags", query: "git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+tt.query, nil)
			res := httptest.NewRecorder()

			h.ServeHTTP(res, req)

			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
			}
			var body struct {
				Query  string                `json:"query"`
				Skills []catalog.SkillRecord `json:"skills"`
				Total  int                   `json:"total"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if body.Query != tt.query || body.Total != 1 || len(body.Skills) != 1 || body.Skills[0].Name != "alpha-skill" {
				t.Fatalf("unexpected body: %#v", body)
			}
		})
	}
}

func TestSearchEndpointRejectsMissingQuery(t *testing.T) {
	h := testHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "skills", "alpha-skill", "SKILL.md"), richSkill("alpha-skill", "PDF automation helper", []string{"git", "search"}))
	writeFile(t, filepath.Join(repo, "skills", "broken-skill", "SKILL.md"), validSkill("wrong-name", "Broken"))

	index, err := catalog.BuildIndex(repo)
	if err != nil {
		t.Fatalf("BuildIndex() error = %v", err)
	}
	return NewServer(index).Handler()
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

func richSkill(name, description string, tags []string) string {
	content := "---\nname: " + name + "\ndescription: " + description + "\ntags:\n"
	for _, tag := range tags {
		content += "  - " + tag + "\n"
	}
	content += "---\n# " + name + "\n"
	return content
}
