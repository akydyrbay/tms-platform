package model

import "time"

type TestCase struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	FolderID  *string   `json:"folder_id,omitempty" db:"folder_id"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TestCaseVersion struct {
	ID             string         `json:"id" db:"id"`
	TestCaseID     string         `json:"test_case_id" db:"test_case_id"`
	VersionNumber  int            `json:"version_number" db:"version_number"`
	Title          string         `json:"title" db:"title"`
	Description    *string        `json:"description,omitempty" db:"description"`
	Preconditions  *string        `json:"preconditions,omitempty" db:"preconditions"`
	ExpectedResult *string        `json:"expected_result,omitempty" db:"expected_result"`
	Module         *string        `json:"module,omitempty" db:"module"`
	Component      *string        `json:"component,omitempty" db:"component"`
	Priority       string         `json:"priority" db:"priority"`
	Type           string         `json:"type" db:"type"`
	Status         string         `json:"status" db:"status"`
	CreatedBy      string         `json:"created_by" db:"created_by"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	Steps          []TestCaseStep `json:"steps" db:"-"`
}

type TestCaseStep struct {
	StepNumber     int     `json:"step_number" db:"step_number"`
	Action         string  `json:"action" db:"action"`
	ExpectedResult *string `json:"expected_result,omitempty" db:"expected_result"`
}

type VersionSummary struct {
	VersionNumber int       `json:"version_number" db:"version_number"`
	CreatedBy     string    `json:"created_by" db:"created_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type TestCaseSummary struct {
	TestCaseID    string  `json:"test_case_id" db:"test_case_id"`
	FolderID      *string `json:"folder_id,omitempty" db:"folder_id"`
	VersionNumber int     `json:"version_number" db:"version_number"`
	Title         string  `json:"title" db:"title"`
	Priority      string  `json:"priority" db:"priority"`
	Status        string  `json:"status" db:"status"`
}
