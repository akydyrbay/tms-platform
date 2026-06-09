package testrun

import (
	"context"
	"errors"
	"strings"

	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("run not found")
	ErrSuiteNotFound  = errors.New("suite not found")
	ErrResultNotFound = errors.New("result not found for that case in this run")
	ErrNameRequired   = errors.New("run name is required")
	ErrInvalidStatus  = errors.New("invalid status")
)

var resultStatuses = map[string]bool{
	"passed": true, "failed": true, "skipped": true, "blocked": true,
}

var runStatuses = map[string]bool{
	"pending": true, "in_progress": true, "completed": true, "cancelled": true,
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	SuiteID   string
	Name      string
	CreatedBy string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.TestRun, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrNameRequired
	}
	run := &model.TestRun{
		ID:        uuid.NewString(),
		SuiteID:   in.SuiteID,
		Name:      name,
		CreatedBy: in.CreatedBy,
	}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) Get(ctx context.Context, id string) (*model.TestRun, error) {
	return s.repo.GetRun(ctx, id)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]model.TestRun, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Results(ctx context.Context, runID string) ([]model.TestRunResult, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	return s.repo.ListResults(ctx, runID)
}

type MarkInput struct {
	Status     string
	Comment    *string
	ExecutedBy string
}

func (s *Service) Mark(ctx context.Context, runID, caseID string, in MarkInput) (*model.TestRunResult, error) {
	if !resultStatuses[in.Status] {
		return nil, ErrInvalidStatus
	}
	executedBy := &in.ExecutedBy
	return s.repo.MarkResult(ctx, runID, caseID, in.Status, in.Comment, executedBy)
}

func (s *Service) SetStatus(ctx context.Context, runID, status string) (*model.TestRun, error) {
	if !runStatuses[status] {
		return nil, ErrInvalidStatus
	}
	return s.repo.UpdateStatus(ctx, runID, status)
}

func (s *Service) Stats(ctx context.Context, runID string) (model.TestRunStats, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return model.TestRunStats{}, err
	}
	return s.repo.Stats(ctx, runID)
}
