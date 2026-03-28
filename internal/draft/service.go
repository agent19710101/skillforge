package draft

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrDraftNotFound = errors.New("draft not found")

type MutationRequest struct {
	Operation string
	SkillName string
	Content   string
}

type Draft struct {
	ID         string           `json:"id"`
	Operation  string           `json:"operation"`
	SkillName  string           `json:"skillName"`
	BranchName string           `json:"branchName"`
	CreatedAt  time.Time        `json:"createdAt"`
	Validation ValidationResult `json:"validation"`

	workspace *Workspace
}

type Service struct {
	Manager    Manager
	Submission SubmissionService

	mu     sync.RWMutex
	drafts map[string]*Draft
}

func NewService(manager Manager, submission SubmissionService) *Service {
	return &Service{
		Manager:    manager,
		Submission: submission,
		drafts:     make(map[string]*Draft),
	}
}

func (s *Service) Create(ctx context.Context, req MutationRequest) (*Draft, error) {
	_ = ctx

	workspace, err := s.Manager.CreateWorkspace(req.Operation, req.SkillName)
	if err != nil {
		return nil, err
	}
	if err := applyMutation(workspace, req); err != nil {
		return nil, err
	}

	validation, err := workspace.Validate()
	if err != nil {
		return nil, err
	}
	draft := &Draft{
		ID:         workspace.ID,
		Operation:  workspace.Operation,
		SkillName:  workspace.SkillName,
		BranchName: workspace.BranchName,
		CreatedAt:  workspace.CreatedAt,
		Validation: validation,
		workspace:  workspace,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.drafts == nil {
		s.drafts = make(map[string]*Draft)
	}
	s.drafts[draft.ID] = draft
	return cloneDraft(draft), nil
}

func (s *Service) Get(id string) (*Draft, bool) {
	s.mu.RLock()
	draft, ok := s.drafts[strings.TrimSpace(id)]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}

	copy := cloneDraft(draft)
	validation, err := draft.workspace.Validate()
	if err == nil {
		copy.Validation = validation
	}
	return copy, true
}

func (s *Service) Submit(ctx context.Context, id string) (SubmissionResult, error) {
	draft, ok := s.lookup(strings.TrimSpace(id))
	if !ok {
		return SubmissionResult{}, ErrDraftNotFound
	}
	return s.Submission.Submit(ctx, draft.workspace)
}

func (s *Service) lookup(id string) (*Draft, bool) {
	s.mu.RLock()
	draft, ok := s.drafts[id]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return draft, true
}

func applyMutation(workspace *Workspace, req MutationRequest) error {
	content := strings.TrimSpace(req.Content)
	switch strings.TrimSpace(req.Operation) {
	case "create":
		if content == "" {
			return fmt.Errorf("draft content is required for create")
		}
		return workspace.CreateSkill(req.Content)
	case "update":
		if content == "" {
			return fmt.Errorf("draft content is required for update")
		}
		return workspace.UpdateSkill(req.Content)
	case "delete":
		return workspace.DeleteSkill()
	default:
		return fmt.Errorf("unsupported draft operation %q", req.Operation)
	}
}

func cloneDraft(draft *Draft) *Draft {
	if draft == nil {
		return nil
	}
	copy := *draft
	return &copy
}
