package suite

import (
	"context"
	"errors"
	"strings"

	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("suite not found")
	ErrNameRequired    = errors.New("suite name is required")
	ErrProjectRequired = errors.New("project_id is required")
	ErrCaseNotFound    = errors.New("test case not found")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	ProjectID   string
	Name        string
	Description *string
	CreatedBy   string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Suite, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrNameRequired
	}
	if in.ProjectID == "" {
		return nil, ErrProjectRequired
	}
	suite := &model.Suite{
		ID:          uuid.NewString(),
		ProjectID:   in.ProjectID,
		Name:        name,
		Description: in.Description,
		CreatedBy:   in.CreatedBy,
		CaseIDs:     []string{},
	}
	if err := s.repo.Create(ctx, suite); err != nil {
		return nil, err
	}
	return suite, nil
}

func (s *Service) Get(ctx context.Context, id string) (*model.Suite, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]model.Suite, error) {
	if projectID == "" {
		return nil, ErrProjectRequired
	}
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) AddCase(ctx context.Context, suiteID, caseID string) (*model.Suite, error) {
	if _, err := s.repo.GetByID(ctx, suiteID); err != nil {
		return nil, err
	}
	exists, err := s.repo.CaseExists(ctx, caseID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCaseNotFound
	}
	if err := s.repo.AddCase(ctx, suiteID, caseID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, suiteID)
}

func (s *Service) RemoveCase(ctx context.Context, suiteID, caseID string) (*model.Suite, error) {
	if _, err := s.repo.GetByID(ctx, suiteID); err != nil {
		return nil, err
	}
	if err := s.repo.RemoveCase(ctx, suiteID, caseID); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, suiteID)
}
