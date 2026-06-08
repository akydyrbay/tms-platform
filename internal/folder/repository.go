package folder

import (
	"context"
	"database/sql"
	"errors"

	"tms-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, f *model.Folder) error {
	const q = `
		INSERT INTO folders (id, project_id, parent_id, name, created_by)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5::uuid)
		RETURNING created_at`
	return r.db.QueryRowxContext(ctx, q, f.ID, f.ProjectID, f.ParentID, f.Name, f.CreatedBy).
		Scan(&f.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Folder, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, parent_id::text AS parent_id,
		       name, created_by::text AS created_by, created_at
		FROM folders
		WHERE id = $1::uuid`
	var f model.Folder
	if err := r.db.GetContext(ctx, &f, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrParentNotFound
		}
		return nil, err
	}
	return &f, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]model.Folder, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, parent_id::text AS parent_id,
		       name, created_by::text AS created_by, created_at
		FROM folders
		WHERE project_id = $1::uuid
		ORDER BY created_at`
	folders := make([]model.Folder, 0)
	if err := r.db.SelectContext(ctx, &folders, q, projectID); err != nil {
		return nil, err
	}
	return folders, nil
}
