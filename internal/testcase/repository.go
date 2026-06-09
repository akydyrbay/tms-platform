package testcase

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

// CreateCase writes the identity row, version 1 and that version's steps in one
// transaction. v.VersionNumber is set to 1.
func (r *Repository) CreateCase(ctx context.Context, c *model.TestCase, v *model.TestCaseVersion) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const caseQ = `
		INSERT INTO test_cases (id, project_id, folder_id, created_by)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		RETURNING created_at`
	if err := tx.QueryRowxContext(ctx, caseQ, c.ID, c.ProjectID, c.FolderID, c.CreatedBy).
		Scan(&c.CreatedAt); err != nil {
		return err
	}

	v.TestCaseID = c.ID
	v.VersionNumber = 1
	if err := insertVersionTx(ctx, tx, v); err != nil {
		return err
	}

	return tx.Commit()
}

// AddVersion inserts the next version (max+1) and its steps in one transaction.
// The previous version and its steps are left untouched.
func (r *Repository) AddVersion(ctx context.Context, v *model.TestCaseVersion) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const nextQ = `
		SELECT COALESCE(MAX(version_number), 0) + 1
		FROM test_case_versions
		WHERE test_case_id = $1::uuid`
	if err := tx.QueryRowxContext(ctx, nextQ, v.TestCaseID).Scan(&v.VersionNumber); err != nil {
		return err
	}

	if err := insertVersionTx(ctx, tx, v); err != nil {
		return err
	}

	return tx.Commit()
}

func insertVersionTx(ctx context.Context, tx *sqlx.Tx, v *model.TestCaseVersion) error {
	const versionQ = `
		INSERT INTO test_case_versions
			(id, test_case_id, version_number, title, description, preconditions,
			 expected_result, module, component, priority, type, status, created_by)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9,
		        $10::priority_enum, $11::case_type_enum, $12::case_status_enum, $13::uuid)
		RETURNING created_at`
	if err := tx.QueryRowxContext(ctx, versionQ,
		v.ID, v.TestCaseID, v.VersionNumber, v.Title, v.Description, v.Preconditions,
		v.ExpectedResult, v.Module, v.Component, v.Priority, v.Type, v.Status, v.CreatedBy,
	).Scan(&v.CreatedAt); err != nil {
		return err
	}

	const stepQ = `
		INSERT INTO test_case_steps (test_case_version_id, step_number, action, expected_result)
		VALUES ($1::uuid, $2, $3, $4)`
	for _, s := range v.Steps {
		if _, err := tx.ExecContext(ctx, stepQ, v.ID, s.StepNumber, s.Action, s.ExpectedResult); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetCase(ctx context.Context, id string) (*model.TestCase, error) {
	const q = `
		SELECT id::text AS id, project_id::text AS project_id, folder_id::text AS folder_id,
		       created_by::text AS created_by, created_at
		FROM test_cases
		WHERE id = $1::uuid`
	var c model.TestCase
	if err := r.db.GetContext(ctx, &c, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

// GetVersion loads a single version with its steps. If versionNumber <= 0 the
// current (highest) version is returned.
func (r *Repository) GetVersion(ctx context.Context, caseID string, versionNumber int) (*model.TestCaseVersion, error) {
	q := `
		SELECT id::text AS id, test_case_id::text AS test_case_id, version_number, title,
		       description, preconditions, expected_result, module, component,
		       priority::text AS priority, type::text AS type, status::text AS status,
		       created_by::text AS created_by, created_at
		FROM test_case_versions
		WHERE test_case_id = $1::uuid`
	args := []any{caseID}
	if versionNumber > 0 {
		q += ` AND version_number = $2`
		args = append(args, versionNumber)
	} else {
		q += ` ORDER BY version_number DESC LIMIT 1`
	}

	var v model.TestCaseVersion
	if err := r.db.GetContext(ctx, &v, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVersionNotFound
		}
		return nil, err
	}

	const stepsQ = `
		SELECT step_number, action, expected_result
		FROM test_case_steps
		WHERE test_case_version_id = $1::uuid
		ORDER BY step_number`
	v.Steps = make([]model.TestCaseStep, 0)
	if err := r.db.SelectContext(ctx, &v.Steps, stepsQ, v.ID); err != nil {
		return nil, err
	}
	return &v, nil
}

// ListByProject returns every case in a project with its current version.
func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]model.TestCaseSummary, error) {
	const q = `
		SELECT tc.id::text AS test_case_id, tc.folder_id::text AS folder_id,
		       cv.version_number, cv.title,
		       cv.priority::text AS priority, cv.status::text AS status
		FROM test_cases tc
		JOIN LATERAL (
			SELECT version_number, title, priority, status
			FROM test_case_versions v
			WHERE v.test_case_id = tc.id
			ORDER BY version_number DESC
			LIMIT 1
		) cv ON true
		WHERE tc.project_id = $1::uuid
		ORDER BY tc.created_at`
	cases := make([]model.TestCaseSummary, 0)
	if err := r.db.SelectContext(ctx, &cases, q, projectID); err != nil {
		return nil, err
	}
	return cases, nil
}

func (r *Repository) ListVersions(ctx context.Context, caseID string) ([]model.VersionSummary, error) {
	const q = `
		SELECT version_number, created_by::text AS created_by, created_at
		FROM test_case_versions
		WHERE test_case_id = $1::uuid
		ORDER BY version_number`
	versions := make([]model.VersionSummary, 0)
	if err := r.db.SelectContext(ctx, &versions, q, caseID); err != nil {
		return nil, err
	}
	return versions, nil
}
