package model

import "time"

type TestRun struct {
	ID          string     `json:"id" db:"id"`
	ProjectID   string     `json:"project_id" db:"project_id"`
	SuiteID     string     `json:"suite_id" db:"suite_id"`
	Name        string     `json:"name" db:"name"`
	Status      string     `json:"status" db:"status"`
	CreatedBy   string     `json:"created_by" db:"created_by"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type TestRunResult struct {
	ID                string     `json:"id" db:"id"`
	TestRunID         string     `json:"test_run_id" db:"test_run_id"`
	TestCaseID        string     `json:"test_case_id" db:"test_case_id"`
	TestCaseVersionID string     `json:"test_case_version_id" db:"test_case_version_id"`
	Status            string     `json:"status" db:"status"`
	Comment           *string    `json:"comment,omitempty" db:"comment"`
	ExecutedBy        *string    `json:"executed_by,omitempty" db:"executed_by"`
	ExecutedAt        *time.Time `json:"executed_at,omitempty" db:"executed_at"`
	Title             string     `json:"title" db:"title"`
	VersionNumber     int        `json:"version_number" db:"version_number"`
}

type TestRunStats struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Blocked int `json:"blocked"`
	NotRun  int `json:"not_run"`
}
