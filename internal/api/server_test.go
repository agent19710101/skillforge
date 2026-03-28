package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agent19710101/skillforge/internal/catalog"
	"github.com/agent19710101/skillforge/internal/draft"
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

func TestCreateDraftEndpointSupportsCreateUpdateAndDelete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		operation string
		skillName string
		content   string
		setupRepo func(t *testing.T) string
	}{
		{
			name:      "create",
			id:        "create01",
			operation: "create",
			skillName: "new-skill",
			content:   validSkill("new-skill", "Fresh draft"),
			setupRepo: func(t *testing.T) string {
				repo := testRepo(t)
				return repo
			},
		},
		{
			name:      "update",
			id:        "update01",
			operation: "update",
			skillName: "example-skill",
			content:   validSkill("example-skill", "Updated draft"),
			setupRepo: func(t *testing.T) string {
				repo := testRepo(t)
				writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Original"))
				return repo
			},
		},
		{
			name:      "delete",
			id:        "delete01",
			operation: "delete",
			skillName: "example-skill",
			setupRepo: func(t *testing.T) string {
				repo := testRepo(t)
				writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Original"))
				return repo
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo(t)
			h := testDraftHandler(t, repo, tt.id, draft.SubmissionService{})

			payload := draftCreateRequestJSON(tt.operation, tt.skillName, tt.content)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(payload))
			res := httptest.NewRecorder()

			h.ServeHTTP(res, req)

			if res.Code != http.StatusCreated {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
			}
			var body draftResponse
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if body.ID != tt.id {
				t.Fatalf("unexpected draft id: %#v", body)
			}
			if body.Operation != tt.operation || body.SkillName != tt.skillName {
				t.Fatalf("unexpected body: %#v", body)
			}
			if body.BranchName != "skillforge/"+tt.operation+"/"+tt.skillName+"/"+tt.id {
				t.Fatalf("unexpected branch name: %#v", body)
			}
			if !body.Validation.Valid {
				t.Fatalf("expected valid draft, findings = %#v", body.Validation.Findings)
			}
		})
	}
}

func TestCreateDraftEndpointRejectsInvalidRequest(t *testing.T) {
	h := testDraftHandler(t, testRepo(t), "invalid01", draft.SubmissionService{})

	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "malformed json",
			payload: "{",
		},
		{
			name:    "unsupported operation",
			payload: draftCreateRequestJSON("publish", "example-skill", validSkill("example-skill", "Broken")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(tt.payload))
			res := httptest.NewRecorder()

			h.ServeHTTP(res, req)

			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestDraftCreateAndStatusExposeSubmissionCapability(t *testing.T) {
	repo := testRepo(t)
	h := testDraftHandler(t, repo, "capability01", draft.SubmissionService{})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("create", "new-skill", validSkill("new-skill", "Fresh draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	var created draftResponse
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("json.Unmarshal(create) error = %v", err)
	}
	if created.Submission.Enabled {
		t.Fatalf("expected submission to be disabled by default, got %#v", created.Submission)
	}
	if created.Submission.Reason == "" {
		t.Fatalf("expected submission reason in create response, got %#v", created.Submission)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/drafts/capability01", nil)
	statusRes := httptest.NewRecorder()
	h.ServeHTTP(statusRes, statusReq)
	if statusRes.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", statusRes.Code, statusRes.Body.String())
	}

	var statusBody draftResponse
	if err := json.Unmarshal(statusRes.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("json.Unmarshal(status) error = %v", err)
	}
	if statusBody.Submission.Enabled {
		t.Fatalf("expected submission to remain disabled, got %#v", statusBody.Submission)
	}
	if statusBody.Submission.Reason == "" {
		t.Fatalf("expected submission reason in status response, got %#v", statusBody.Submission)
	}
}

func TestDraftCreateAndStatusExposeEnabledSubmissionCapability(t *testing.T) {
	repo := testRepo(t)
	h := testDraftHandler(t, repo, "enabled01", draft.SubmissionService{
		Config: draft.SubmissionConfig{
			ServerURL:  "https://forgejo.example",
			RemoteName: "origin",
			Owner:      "acme",
			Repo:       "skillforge",
			BaseBranch: "main",
			Token:      "secret",
		},
		Git:     &fakeGitPublisher{},
		Forgejo: &fakeForgejoClient{},
	})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("create", "new-skill", validSkill("new-skill", "Fresh draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	var created draftResponse
	if err := json.Unmarshal(createRes.Body.Bytes(), &created); err != nil {
		t.Fatalf("json.Unmarshal(create) error = %v", err)
	}
	if !created.Submission.Enabled || created.Submission.BaseBranch != "main" {
		t.Fatalf("unexpected create submission status: %#v", created.Submission)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/drafts/enabled01", nil)
	statusRes := httptest.NewRecorder()
	h.ServeHTTP(statusRes, statusReq)
	if statusRes.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", statusRes.Code, statusRes.Body.String())
	}

	var statusBody draftResponse
	if err := json.Unmarshal(statusRes.Body.Bytes(), &statusBody); err != nil {
		t.Fatalf("json.Unmarshal(status) error = %v", err)
	}
	if !statusBody.Submission.Enabled || statusBody.Submission.BaseBranch != "main" {
		t.Fatalf("unexpected status submission: %#v", statusBody.Submission)
	}
}

func TestDraftStatusReturnsValidationFindings(t *testing.T) {
	repo := testRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Original"))
	h := testDraftHandler(t, repo, "status01", draft.SubmissionService{})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("update", "example-skill", validSkill("wrong-name", "Broken draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/v1/drafts/status01", nil)
	statusRes := httptest.NewRecorder()
	h.ServeHTTP(statusRes, statusReq)

	if statusRes.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", statusRes.Code, statusRes.Body.String())
	}
	var body draftResponse
	if err := json.Unmarshal(statusRes.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.Validation.Valid {
		t.Fatal("expected invalid draft status")
	}
	if !hasFindingCode(body.Validation.Findings, "name_directory_mismatch") {
		t.Fatalf("unexpected findings: %#v", body.Validation.Findings)
	}
}

func TestSubmitDraftReturnsBranchAndPullRequestMetadata(t *testing.T) {
	repo := testRepo(t)
	h := testDraftHandler(t, repo, "submit01", draft.SubmissionService{
		Config: draft.SubmissionConfig{
			ServerURL:  "https://forgejo.example",
			RemoteName: "origin",
			Owner:      "acme",
			Repo:       "skillforge",
			BaseBranch: "main",
			Token:      "secret",
		},
		Git: &fakeGitPublisher{commitHash: "abc123"},
		Forgejo: &fakeForgejoClient{pullRequest: draft.PullRequest{
			Number: 17,
			ID:     42,
			URL:    "https://forgejo.example/acme/skillforge/pulls/17",
		}},
	})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("create", "new-skill", validSkill("new-skill", "Fresh draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts/submit01/submit", nil)
	submitRes := httptest.NewRecorder()
	h.ServeHTTP(submitRes, submitReq)

	if submitRes.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", submitRes.Code, submitRes.Body.String())
	}
	var body draftSubmissionResponse
	if err := json.Unmarshal(submitRes.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body.BranchName != "skillforge/create/new-skill/submit01" {
		t.Fatalf("unexpected submit response: %#v", body)
	}
	if body.BaseBranch != "main" {
		t.Fatalf("unexpected base branch: %#v", body)
	}
	if body.PullRequest == nil || body.PullRequest.Number != 17 || body.PullRequest.ID != 42 || body.PullRequest.URL == "" {
		t.Fatalf("unexpected pull request: %#v", body.PullRequest)
	}
	if body.CommitHash != "abc123" {
		t.Fatalf("unexpected commit hash: %#v", body.CommitHash)
	}
}

func TestSubmitDraftReturnsUnavailableWhenSubmissionIsDisabled(t *testing.T) {
	repo := testRepo(t)
	h := testDraftHandler(t, repo, "disabled01", draft.SubmissionService{})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("create", "new-skill", validSkill("new-skill", "Fresh draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	submitReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts/disabled01/submit", nil)
	submitRes := httptest.NewRecorder()
	h.ServeHTTP(submitRes, submitReq)
	if submitRes.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", submitRes.Code, submitRes.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(submitRes.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if body["error"] != "submission_unavailable" {
		t.Fatalf("unexpected response: %#v", body)
	}
	submission, ok := body["submission"].(map[string]any)
	if !ok || submission["enabled"] != false {
		t.Fatalf("unexpected submission status: %#v", body)
	}
}

func TestSubmitDraftRejectsBlockedAndMissingDrafts(t *testing.T) {
	repo := testRepo(t)
	writeFile(t, filepath.Join(repo, "skills", "example-skill", "SKILL.md"), validSkill("example-skill", "Original"))
	h := testDraftHandler(t, repo, "blocked01", draft.SubmissionService{
		Config: draft.SubmissionConfig{
			ServerURL:  "https://forgejo.example",
			RemoteName: "origin",
			Owner:      "acme",
			Repo:       "skillforge",
			BaseBranch: "main",
			Token:      "secret",
		},
		Git:     &fakeGitPublisher{},
		Forgejo: &fakeForgejoClient{},
	})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts", strings.NewReader(draftCreateRequestJSON("update", "example-skill", validSkill("wrong-name", "Broken draft"))))
	createRes := httptest.NewRecorder()
	h.ServeHTTP(createRes, createReq)
	if createRes.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRes.Code, createRes.Body.String())
	}

	blockedReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts/blocked01/submit", nil)
	blockedRes := httptest.NewRecorder()
	h.ServeHTTP(blockedRes, blockedReq)
	if blockedRes.Code != http.StatusConflict {
		t.Fatalf("blocked status = %d, body = %s", blockedRes.Code, blockedRes.Body.String())
	}
	var blockedBody map[string]any
	if err := json.Unmarshal(blockedRes.Body.Bytes(), &blockedBody); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if blockedBody["error"] != "draft_invalid" {
		t.Fatalf("unexpected blocked response: %#v", blockedBody)
	}

	missingReq := httptest.NewRequest(http.MethodPost, "/api/v1/drafts/missing/submit", nil)
	missingRes := httptest.NewRecorder()
	h.ServeHTTP(missingRes, missingReq)
	if missingRes.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, body = %s", missingRes.Code, missingRes.Body.String())
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
	drafts := draft.NewService(draft.Manager{RepoRoot: repo}, draft.SubmissionService{})
	return NewServer(index, WithDraftService(drafts)).Handler()
}

func testDraftHandler(t *testing.T, repo, id string, submission draft.SubmissionService) http.Handler {
	t.Helper()
	index, err := catalog.BuildIndex(repo)
	if err != nil {
		t.Fatalf("BuildIndex() error = %v", err)
	}
	drafts := draft.NewService(draft.Manager{RepoRoot: repo, NewID: func() string { return id }}, submission)
	return NewServer(index, WithDraftService(drafts)).Handler()
}

func testRepo(t *testing.T) string {
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

func richSkill(name, description string, tags []string) string {
	content := "---\nname: " + name + "\ndescription: " + description + "\ntags:\n"
	for _, tag := range tags {
		content += "  - " + tag + "\n"
	}
	content += "---\n# " + name + "\n"
	return content
}

func draftCreateRequestJSON(operation, skillName, content string) string {
	body := map[string]string{
		"operation": operation,
		"skillName": skillName,
	}
	if content != "" {
		body["content"] = content
	}
	data, _ := json.Marshal(body)
	return string(data)
}

func hasFindingCode(findings []catalog.Finding, want string) bool {
	for _, finding := range findings {
		if finding.Code == want {
			return true
		}
	}
	return false
}

type fakeGitPublisher struct {
	commitHash   string
	commitErr    error
	publishErr   error
	commitCalls  int
	publishCalls int
	lastCommit   draft.CommitRequest
	lastPublish  draft.PublishRequest
}

func (f *fakeGitPublisher) Commit(_ context.Context, req draft.CommitRequest) (string, error) {
	f.commitCalls++
	f.lastCommit = req
	if f.commitErr != nil {
		return "", f.commitErr
	}
	return f.commitHash, nil
}

func (f *fakeGitPublisher) Publish(_ context.Context, req draft.PublishRequest) error {
	f.publishCalls++
	f.lastPublish = req
	return f.publishErr
}

type fakeForgejoClient struct {
	pullRequest draft.PullRequest
	err         error
	createCalls int
	lastRequest draft.PullRequestRequest
}

func (f *fakeForgejoClient) CreatePullRequest(_ context.Context, req draft.PullRequestRequest) (draft.PullRequest, error) {
	f.createCalls++
	f.lastRequest = req
	if f.err != nil {
		return draft.PullRequest{}, f.err
	}
	return f.pullRequest, nil
}
