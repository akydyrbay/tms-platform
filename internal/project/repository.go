package project

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

func (r *Repository) Create(ctx context.Context, p *model.Project) error {
	const q = `
		INSERT INTO projects (id, name, description, created_by)
		VALUES ($1::uuid, $2, $3, $4::uuid)
		RETURNING created_at`
	return r.db.QueryRowxContext(ctx, q, p.ID, p.Name, p.Description, p.CreatedBy).
		Scan(&p.CreatedAt)
}

func (r *Repository) List(ctx context.Context) ([]model.Project, error) {
	const q = `
		SELECT id::text AS id, name, description, created_by::text AS created_by, created_at
		FROM projects
		ORDER BY created_at DESC`
	projects := make([]model.Project, 0)
	if err := r.db.SelectContext(ctx, &projects, q); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Project, error) {
	const q = `
		SELECT id::text AS id, name, description, created_by::text AS created_by, created_at
		FROM projects
		WHERE id = $1::uuid`
	var p model.Project
	if err := r.db.GetContext(ctx, &p, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}
