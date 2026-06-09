package testrun

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

func (r *Repository) CreateRun(ctx context.Context, run *model.TestRun) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const runQ = `
		INSERT INTO test_runs (id, project_id, suite_id, name, created_by)
		SELECT $1::uuid, project_id, id, $3, $4::uuid
		FROM test_suites
		WHERE id = $2::uuid
		RETURNING project_id::text, status::text, created_at`
	err = tx.QueryRowxContext(ctx, runQ, run.ID, run.SuiteID, run.Name, run.CreatedBy).
		Scan(&run.ProjectID, &run.Status, &run.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSuiteNotFound
	}
	if err != nil {
		return err
	}

	const snapshotQ = `
		INSERT INTO test_run_results (id, test_run_id, test_case_id, test_case_version_id, status)
		SELECT gen_random_uuid(), $1::uuid, stc.test_case_id, cv.id, 'not_run'
		FROM suite_test_cases stc
		JOIN LATERAL (
			SELECT id
			FROM test_case_versions v
			WHERE v.test_case_id = stc.test_case_id
			ORDER BY version_number DESC
			LIMIT 1
		) cv ON true
		WHERE stc.suite_id = $2::uuid`
	if _, err := tx.ExecContext(ctx, snapshotQ, run.ID, run.SuiteID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetRun(ctx context.Context, id string) (*model.TestRun, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, suite_id::text AS suite_id,
		       name, status::text AS status, created_by::text AS created_by,
		       started_at, completed_at, created_at
		FROM test_runs
		WHERE id = $1::uuid`
	var run model.TestRun
	if err := r.db.GetContext(ctx, &run, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]model.TestRun, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, suite_id::text AS suite_id,
		       name, status::text AS status, created_by::text AS created_by,
		       started_at, completed_at, created_at
		FROM test_runs
		WHERE project_id = $1::uuid
		ORDER BY created_at DESC`
	runs := make([]model.TestRun, 0)
	if err := r.db.SelectContext(ctx, &runs, q, projectID); err != nil {
		return nil, err
	}
	return runs, nil
}

func (r *Repository) ListResults(ctx context.Context, runID string) ([]model.TestRunResult, error) {
	const q = `
		SELECT id::text AS id, test_run_id::text AS test_run_id, test_case_id::text AS test_case_id,
		       test_case_version_id::text AS test_case_version_id, status::text AS status,
		       comment, executed_by::text AS executed_by, executed_at
		FROM test_run_results
		WHERE test_run_id = $1::uuid
		ORDER BY test_case_id`
	results := make([]model.TestRunResult, 0)
	if err := r.db.SelectContext(ctx, &results, q, runID); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *Repository) MarkResult(ctx context.Context, runID, caseID, status string, comment, executedBy *string) (*model.TestRunResult, error) {
	const q = `
		UPDATE test_run_results
		SET status = $3::test_result_status_enum, comment = $4, executed_by = $5::uuid, executed_at = now()
		WHERE test_run_id = $1::uuid AND test_case_id = $2::uuid
		RETURNING id::text AS id, test_run_id::text AS test_run_id, test_case_id::text AS test_case_id,
		          test_case_version_id::text AS test_case_version_id, status::text AS status,
		          comment, executed_by::text AS executed_by, executed_at`
	var res model.TestRunResult
	if err := r.db.GetContext(ctx, &res, q, runID, caseID, status, comment, executedBy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrResultNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, runID, status string) (*model.TestRun, error) {
	const q = `
		UPDATE test_runs
		SET status = $2::test_run_status_enum,
		    started_at = CASE WHEN $2 = 'in_progress' AND started_at IS NULL THEN now() ELSE started_at END,
		    completed_at = CASE WHEN $2 = 'completed' THEN now() ELSE completed_at END
		WHERE id = $1::uuid
		RETURNING id::text AS id, project_id::text AS project_id, suite_id::text AS suite_id,
		          name, status::text AS status, created_by::text AS created_by,
		          started_at, completed_at, created_at`
	var run model.TestRun
	if err := r.db.GetContext(ctx, &run, q, runID, status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &run, nil
}

func (r *Repository) Stats(ctx context.Context, runID string) (model.TestRunStats, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'passed') AS passed,
			COUNT(*) FILTER (WHERE status = 'failed') AS failed,
			COUNT(*) FILTER (WHERE status = 'skipped') AS skipped,
			COUNT(*) FILTER (WHERE status = 'blocked') AS blocked,
			COUNT(*) FILTER (WHERE status = 'not_run') AS not_run
		FROM test_run_results
		WHERE test_run_id = $1::uuid`
	var s model.TestRunStats
	err := r.db.QueryRowxContext(ctx, q, runID).
		Scan(&s.Total, &s.Passed, &s.Failed, &s.Skipped, &s.Blocked, &s.NotRun)
	return s, err
}
