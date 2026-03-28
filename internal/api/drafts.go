package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agent19710101/skillforge/internal/catalog"
	"github.com/agent19710101/skillforge/internal/draft"
)

type draftCreateRequest struct {
	Operation string `json:"operation"`
	SkillName string `json:"skillName"`
	Content   string `json:"content,omitempty"`
}

type draftValidationResponse struct {
	Valid    bool              `json:"valid"`
	Findings []catalog.Finding `json:"findings,omitempty"`
}

type draftSubmissionStatusResponse struct {
	Enabled    bool   `json:"enabled"`
	BaseBranch string `json:"baseBranch,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type draftResponse struct {
	ID         string                        `json:"id"`
	Operation  string                        `json:"operation"`
	SkillName  string                        `json:"skillName"`
	BranchName string                        `json:"branchName"`
	CreatedAt  time.Time                     `json:"createdAt"`
	Validation draftValidationResponse       `json:"validation"`
	Submission draftSubmissionStatusResponse `json:"submission"`
}

type draftSubmissionResponse struct {
	ID          string                  `json:"id"`
	Operation   string                  `json:"operation"`
	SkillName   string                  `json:"skillName"`
	BranchName  string                  `json:"branchName"`
	BaseBranch  string                  `json:"baseBranch"`
	CommitHash  string                  `json:"commitHash,omitempty"`
	PullRequest *draft.PullRequestRef   `json:"pullRequest,omitempty"`
	Validation  draftValidationResponse `json:"validation"`
}

func (s *Server) handleDraftCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if s.drafts == nil {
		writeError(w, http.StatusServiceUnavailable, "drafts_unavailable", "draft service is not configured")
		return
	}

	req, err := decodeDraftCreateRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validateDraftCreateRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	draftRecord, err := s.drafts.Create(r.Context(), draft.MutationRequest{
		Operation: req.Operation,
		SkillName: req.SkillName,
		Content:   req.Content,
	})
	if err != nil {
		writeError(w, http.StatusConflict, "draft_conflict", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, draftResponseFrom(draftRecord, s.drafts.SubmissionStatus()))
}

func (s *Server) handleDraftItem(w http.ResponseWriter, r *http.Request) {
	if s.drafts == nil {
		writeError(w, http.StatusServiceUnavailable, "drafts_unavailable", "draft service is not configured")
		return
	}

	id, action, ok := splitDraftPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "draft not found")
		return
	}

	switch action {
	case "":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		draftRecord, found := s.drafts.Get(id)
		if !found {
			writeError(w, http.StatusNotFound, "not_found", "draft not found")
			return
		}
		writeJSON(w, http.StatusOK, draftResponseFrom(draftRecord, s.drafts.SubmissionStatus()))
	case "submit":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
			return
		}
		status := s.drafts.SubmissionStatus()
		if !status.Enabled {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"error":      "submission_unavailable",
				"message":    status.Reason,
				"submission": draftSubmissionStatusResponseFrom(status),
			})
			return
		}
		result, err := s.drafts.Submit(r.Context(), id)
		if err != nil {
			switch {
			case errors.Is(err, draft.ErrDraftNotFound):
				writeError(w, http.StatusNotFound, "not_found", "draft not found")
			case errors.Is(err, draft.ErrInvalidDraft):
				var validationErr draft.ValidationError
				if errors.As(err, &validationErr) {
					writeJSON(w, http.StatusConflict, map[string]any{
						"error":      "draft_invalid",
						"message":    validationErr.Error(),
						"validation": draftValidationResponseFrom(validationErr.Result),
						"submission": draftSubmissionStatusResponseFrom(status),
					})
					return
				}
				writeError(w, http.StatusConflict, "draft_invalid", err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "internal_error", "submit draft")
			}
			return
		}
		draftRecord, found := s.drafts.Get(id)
		if !found {
			writeError(w, http.StatusNotFound, "not_found", "draft not found")
			return
		}
		writeJSON(w, http.StatusOK, draftSubmissionResponse{
			ID:          draftRecord.ID,
			Operation:   draftRecord.Operation,
			SkillName:   draftRecord.SkillName,
			BranchName:  result.BranchName,
			BaseBranch:  result.BaseBranch,
			CommitHash:  result.CommitHash,
			PullRequest: result.PullRequest,
			Validation:  draftValidationResponseFrom(result.Validation),
		})
	default:
		writeError(w, http.StatusNotFound, "not_found", "draft not found")
	}
}

func decodeDraftCreateRequest(r *http.Request) (draftCreateRequest, error) {
	defer r.Body.Close()

	var req draftCreateRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return draftCreateRequest{}, fmt.Errorf("decode request body: %w", err)
	}
	return req, nil
}

func validateDraftCreateRequest(req draftCreateRequest) error {
	switch strings.TrimSpace(req.Operation) {
	case "create", "update", "delete":
	default:
		return fmt.Errorf("unsupported operation %q", req.Operation)
	}
	if !catalog.IsCanonicalSkillName(req.SkillName) {
		return fmt.Errorf("invalid skill name %q", req.SkillName)
	}
	if req.Operation != "delete" && strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("content is required for %s", req.Operation)
	}
	return nil
}

func splitDraftPath(path string) (id, action string, ok bool) {
	trimmed := strings.TrimPrefix(path, "/api/v1/drafts/")
	if trimmed == path || trimmed == "" {
		return "", "", false
	}
	parts := strings.Split(trimmed, "/")
	switch len(parts) {
	case 1:
		return parts[0], "", true
	case 2:
		return parts[0], parts[1], true
	default:
		return "", "", false
	}
}

func draftResponseFrom(record *draft.Draft, status draft.SubmissionStatus) draftResponse {
	return draftResponse{
		ID:         record.ID,
		Operation:  record.Operation,
		SkillName:  record.SkillName,
		BranchName: record.BranchName,
		CreatedAt:  record.CreatedAt,
		Validation: draftValidationResponseFrom(record.Validation),
		Submission: draftSubmissionStatusResponseFrom(status),
	}
}

func draftSubmissionStatusResponseFrom(status draft.SubmissionStatus) draftSubmissionStatusResponse {
	return draftSubmissionStatusResponse{
		Enabled:    status.Enabled,
		BaseBranch: status.BaseBranch,
		Reason:     status.Reason,
	}
}

func draftValidationResponseFrom(result draft.ValidationResult) draftValidationResponse {
	return draftValidationResponse{
		Valid:    result.Valid,
		Findings: result.Findings,
	}
}
