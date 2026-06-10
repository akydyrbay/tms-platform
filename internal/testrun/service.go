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

type CIEntry struct {
	TestCaseID string
	Status     string
	Comment    *string
}

type CIImportInput struct {
	Source      string
	PipelineURL *string
	Results     []CIEntry
	ImportedBy  string
}

type CIImportSummary struct {
	RunID    string             `json:"run_id"`
	Source   string             `json:"source"`
	Imported int                `json:"imported"`
	Skipped  int                `json:"skipped"` // cases not part of this run
	Stats    model.TestRunStats `json:"stats"`
}

func (s *Service) ImportCI(ctx context.Context, runID string, in CIImportInput) (*CIImportSummary, error) {
	if _, err := s.repo.GetRun(ctx, runID); err != nil {
		return nil, err
	}
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = "gitlab-ci"
	}

	summary := &CIImportSummary{RunID: runID, Source: source}
	executedBy := &in.ImportedBy

	for _, e := range in.Results {
		if !resultStatuses[e.Status] {
			return nil, ErrInvalidStatus
		}
		comment := ciComment(source, in.PipelineURL, e.Comment)
		_, err := s.repo.MarkResult(ctx, runID, e.TestCaseID, e.Status, comment, executedBy)
		if errors.Is(err, ErrResultNotFound) {
			summary.Skipped++
			continue
		}
		if err != nil {
			return nil, err
		}
		summary.Imported++
	}

	stats, err := s.repo.Stats(ctx, runID)
	if err != nil {
		return nil, err
	}
	summary.Stats = stats
	return summary, nil
}

func ciComment(source string, pipelineURL, comment *string) *string {
	prefix := "[" + source
	if pipelineURL != nil && strings.TrimSpace(*pipelineURL) != "" {
		prefix += " " + strings.TrimSpace(*pipelineURL)
	}
	prefix += "]"
	if comment != nil && strings.TrimSpace(*comment) != "" {
		prefix += " " + strings.TrimSpace(*comment)
	}
	return &prefix
}
