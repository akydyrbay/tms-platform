package testcase

import (
	"context"
	"errors"
	"testing"

	"tms-platform/internal/model"
)

func TestBuildVersionAppliesDefaultsAndRenumbersSteps(t *testing.T) {
	v := buildVersion(Content{
		Title: "  Login works  ",
		Steps: []model.TestCaseStep{
			{Action: "open page"},
			{Action: "submit"},
		},
	}, "user-1")

	if v.Title != "Login works" {
		t.Errorf("title should be trimmed, got %q", v.Title)
	}
	if v.Priority != "medium" || v.Type != "functional" || v.Status != "active" {
		t.Errorf("defaults not applied: %s/%s/%s", v.Priority, v.Type, v.Status)
	}
	if v.CreatedBy != "user-1" {
		t.Errorf("created_by: got %q", v.CreatedBy)
	}
	if v.ID == "" {
		t.Error("a version id should be generated")
	}
	if len(v.Steps) != 2 || v.Steps[0].StepNumber != 1 || v.Steps[1].StepNumber != 2 {
		t.Errorf("steps should be renumbered from 1, got %+v", v.Steps)
	}
}

func TestBuildVersionKeepsExplicitValues(t *testing.T) {
	v := buildVersion(Content{
		Title:    "x",
		Priority: "high",
		Type:     "smoke",
		Status:   "draft",
	}, "u")
	if v.Priority != "high" || v.Type != "smoke" || v.Status != "draft" {
		t.Errorf("explicit values should win over defaults: %s/%s/%s", v.Priority, v.Type, v.Status)
	}
}

// Validation happens before any DB call, so a nil repo is safe here.
func TestCreateValidation(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()

	if _, err := svc.Create(ctx, CreateInput{Content: Content{Title: "ok"}}); !errors.Is(err, ErrProjectRequired) {
		t.Errorf("missing project_id: want ErrProjectRequired, got %v", err)
	}
	if _, err := svc.Create(ctx, CreateInput{ProjectID: "p1", Content: Content{Title: "   "}}); !errors.Is(err, ErrTitleRequired) {
		t.Errorf("blank title: want ErrTitleRequired, got %v", err)
	}
}
