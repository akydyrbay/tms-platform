package project

import (
	"context"
	"errors"
	"strings"
	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNotFound     = errors.New("project not found")
	ErrNameRequired = errors.New("project name is required")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name        string
	Description *string
	CreatedBy   string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrNameRequired
	}
	p := &model.Project{
		ID:          uuid.NewString(),
		Name:        name,
		Description: in.Description,
		CreatedBy:   in.CreatedBy,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) List(ctx context.Context) ([]model.Project, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*model.Project, error) {
	return s.repo.GetByID(ctx, id)
}
