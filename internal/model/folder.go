package model

import "time"

type Folder struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	ParentID  *string   `json:"parent_id,omitempty" db:"parent_id"`
	Name      string    `json:"name" db:"name"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Children  []*Folder `json:"children" db:"-"`
}
