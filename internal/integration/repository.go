package integration

import (
	"context"

	"tms-platform/internal/model"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ResultExists(ctx context.Context, resultID string) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM test_run_results WHERE id = $1::uuid)`
	var exists bool
	if err := r.db.GetContext(ctx, &exists, q, resultID); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) Create(ctx context.Context, b *model.Bug) error {
	const q = `
		INSERT INTO bugs (id, test_run_result_id, tracker, external_id, url, title, created_by)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7::uuid)
		RETURNING created_at`
	return r.db.QueryRowxContext(ctx, q,
		b.ID, b.TestRunResultID, b.Tracker, b.ExternalID, b.URL, b.Title, b.CreatedBy).
		Scan(&b.CreatedAt)
}

func (r *Repository) ListByResult(ctx context.Context, resultID string) ([]model.Bug, error) {
	const q = `
		SELECT id::text AS id, test_run_result_id::text AS test_run_result_id,
		       tracker, external_id, url, title, created_by::text AS created_by, created_at
		FROM bugs
		WHERE test_run_result_id = $1::uuid
		ORDER BY created_at`
	bugs := make([]model.Bug, 0)
	if err := r.db.SelectContext(ctx, &bugs, q, resultID); err != nil {
		return nil, err
	}
	return bugs, nil
}
