package testcase

import (
	"context"
	"errors"
	"strings"

	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound        = errors.New("test case not found")
	ErrVersionNotFound = errors.New("test case version not found")
	ErrTitleRequired   = errors.New("title is required")
	ErrProjectRequired = errors.New("project_id is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Content is the full set of versioned fields for a test case. An edit supplies
// a complete new Content; it never patches an existing version in place.
type Content struct {
	Title          string
	Description    *string
	Preconditions  *string
	ExpectedResult *string
	Module         *string
	Component      *string
	Priority       string
	Type           string
	Status         string
	Steps          []model.TestCaseStep
}

type CreateInput struct {
	ProjectID string
	FolderID  *string
	CreatedBy string
	Content
}

// Create writes the identity row, version 1 and its steps in one transaction.
func (s *Service) Create(ctx context.Context, in CreateInput) (*model.TestCaseVersion, error) {
	if in.ProjectID == "" {
		return nil, ErrProjectRequired
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, ErrTitleRequired
	}

	c := &model.TestCase{
		ID:        uuid.NewString(),
		ProjectID: in.ProjectID,
		FolderID:  in.FolderID,
		CreatedBy: in.CreatedBy,
	}
	v := buildVersion(in.Content, in.CreatedBy)
	if err := s.repo.CreateCase(ctx, c, v); err != nil {
		return nil, err
	}
	return s.repo.GetVersion(ctx, c.ID, v.VersionNumber)
}

// Edit inserts a new version holding the full updated content; the previous
// version and its steps are left untouched.
func (s *Service) Edit(ctx context.Context, caseID, editedBy string, content Content) (*model.TestCaseVersion, error) {
	if strings.TrimSpace(content.Title) == "" {
		return nil, ErrTitleRequired
	}
	if _, err := s.repo.GetCase(ctx, caseID); err != nil {
		return nil, err
	}

	v := buildVersion(content, editedBy)
	v.TestCaseID = caseID
	if err := s.repo.AddVersion(ctx, v); err != nil {
		return nil, err
	}
	return s.repo.GetVersion(ctx, caseID, v.VersionNumber)
}

func (s *Service) GetCurrent(ctx context.Context, caseID string) (*model.TestCaseVersion, error) {
	if _, err := s.repo.GetCase(ctx, caseID); err != nil {
		return nil, err
	}
	return s.repo.GetVersion(ctx, caseID, 0)
}

func (s *Service) GetVersion(ctx context.Context, caseID string, versionNumber int) (*model.TestCaseVersion, error) {
	return s.repo.GetVersion(ctx, caseID, versionNumber)
}

func (s *Service) ListByProject(ctx context.Context, projectID string) ([]model.TestCaseSummary, error) {
	if projectID == "" {
		return nil, ErrProjectRequired
	}
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) History(ctx context.Context, caseID string) ([]model.VersionSummary, error) {
	if _, err := s.repo.GetCase(ctx, caseID); err != nil {
		return nil, err
	}
	return s.repo.ListVersions(ctx, caseID)
}

func buildVersion(c Content, createdBy string) *model.TestCaseVersion {
	steps := make([]model.TestCaseStep, len(c.Steps))
	for i, s := range c.Steps {
		s.StepNumber = i + 1
		steps[i] = s
	}
	return &model.TestCaseVersion{
		ID:             uuid.NewString(),
		Title:          strings.TrimSpace(c.Title),
		Description:    c.Description,
		Preconditions:  c.Preconditions,
		ExpectedResult: c.ExpectedResult,
		Module:         c.Module,
		Component:      c.Component,
		Priority:       defaultStr(c.Priority, "medium"),
		Type:           defaultStr(c.Type, "functional"),
		Status:         defaultStr(c.Status, "active"),
		CreatedBy:      createdBy,
		Steps:          steps,
	}
}

func defaultStr(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
