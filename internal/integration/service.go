package integration

import (
	"context"
	"errors"
	"strings"

	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrResultNotFound  = errors.New("run result not found")
	ErrTrackerRequired = errors.New("tracker and external_id are required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type AttachInput struct {
	ResultID   string
	Tracker    string
	ExternalID string
	URL        *string
	Title      *string
	CreatedBy  string
}

func (s *Service) Attach(ctx context.Context, in AttachInput) (*model.Bug, error) {
	if strings.TrimSpace(in.Tracker) == "" || strings.TrimSpace(in.ExternalID) == "" {
		return nil, ErrTrackerRequired
	}
	exists, err := s.repo.ResultExists(ctx, in.ResultID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrResultNotFound
	}
	b := &model.Bug{
		ID:              uuid.NewString(),
		TestRunResultID: in.ResultID,
		Tracker:         strings.TrimSpace(in.Tracker),
		ExternalID:      strings.TrimSpace(in.ExternalID),
		URL:             in.URL,
		Title:           in.Title,
		CreatedBy:       in.CreatedBy,
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) ListByResult(ctx context.Context, resultID string) ([]model.Bug, error) {
	return s.repo.ListByResult(ctx, resultID)
}
