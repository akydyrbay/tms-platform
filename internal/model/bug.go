package model

import "time"

type Bug struct {
	ID              string    `json:"id" db:"id"`
	TestRunResultID string    `json:"test_run_result_id" db:"test_run_result_id"`
	Tracker         string    `json:"tracker" db:"tracker"`
	ExternalID      string    `json:"external_id" db:"external_id"`
	URL             *string   `json:"url,omitempty" db:"url"`
	Title           *string   `json:"title,omitempty" db:"title"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}
