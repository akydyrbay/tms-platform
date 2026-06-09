package folder

import (
	"context"
	"errors"
	"strings"

	"tms-platform/internal/model"

	"github.com/google/uuid"
)

var (
	ErrNameRequired           = errors.New("folder name is required")
	ErrProjectRequired        = errors.New("project_id is required")
	ErrParentNotFound         = errors.New("parent folder not found")
	ErrParentDifferentProject = errors.New("parent folder belongs to a different project")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	ProjectID string
	ParentID  *string
	Name      string
	CreatedBy string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*model.Folder, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, ErrNameRequired
	}
	if in.ProjectID == "" {
		return nil, ErrProjectRequired
	}

	if in.ParentID != nil {
		parent, err := s.repo.GetByID(ctx, *in.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.ProjectID != in.ProjectID {
			return nil, ErrParentDifferentProject
		}
	}

	f := &model.Folder{
		ID:        uuid.NewString(),
		ProjectID: in.ProjectID,
		ParentID:  in.ParentID,
		Name:      name,
		CreatedBy: in.CreatedBy,
	}
	if err := s.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *Service) Tree(ctx context.Context, projectID string) ([]*model.Folder, error) {
	flat, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	byID := make(map[string]*model.Folder, len(flat))
	for i := range flat {
		flat[i].Children = []*model.Folder{}
		byID[flat[i].ID] = &flat[i]
	}

	roots := make([]*model.Folder, 0)
	for i := range flat {
		f := byID[flat[i].ID]
		if f.ParentID != nil {
			if parent, ok := byID[*f.ParentID]; ok {
				parent.Children = append(parent.Children, f)
				continue
			}
		}
		roots = append(roots, f)
	}
	return roots, nil
}
