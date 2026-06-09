package suite

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

func (r *Repository) Create(ctx context.Context, s *model.Suite) error {
	const q = `
		INSERT INTO test_suites (id, project_id, name, description, created_by)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5::uuid)
		RETURNING created_at`
	return r.db.QueryRowxContext(ctx, q, s.ID, s.ProjectID, s.Name, s.Description, s.CreatedBy).
		Scan(&s.CreatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id string) (*model.Suite, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, name, description,
		       created_by::text AS created_by, created_at
		FROM test_suites
		WHERE id = $1::uuid`
	var s model.Suite
	if err := r.db.GetContext(ctx, &s, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	const casesQ = `
		SELECT test_case_id::text
		FROM suite_test_cases
		WHERE suite_id = $1::uuid
		ORDER BY test_case_id`
	s.CaseIDs = make([]string, 0)
	if err := r.db.SelectContext(ctx, &s.CaseIDs, casesQ, id); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]model.Suite, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, name, description,
		       created_by::text AS created_by, created_at
		FROM test_suites
		WHERE project_id = $1::uuid
		ORDER BY created_at`
	suites := make([]model.Suite, 0)
	if err := r.db.SelectContext(ctx, &suites, q, projectID); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r *Repository) CaseExists(ctx context.Context, caseID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM test_cases WHERE id = $1::uuid)`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, q, caseID); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) AddCase(ctx context.Context, suiteID, caseID string) error {
	const q = `
		INSERT INTO suite_test_cases (suite_id, test_case_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, q, suiteID, caseID)
	return err
}

func (r *Repository) RemoveCase(ctx context.Context, suiteID, caseID string) error {
	const q = `DELETE FROM suite_test_cases WHERE suite_id = $1::uuid AND test_case_id = $2::uuid`
	_, err := r.db.ExecContext(ctx, q, suiteID, caseID)
	return err
}
