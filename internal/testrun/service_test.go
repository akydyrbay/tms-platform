package testrun

import (
	"context"
	"errors"
	"testing"
)

func TestMarkRejectsInvalidStatus(t *testing.T) {
	svc := NewService(nil)
	if _, err := svc.Mark(context.Background(), "run", "case", MarkInput{Status: "bogus"}); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("want ErrInvalidStatus, got %v", err)
	}
}

func TestSetStatusRejectsInvalidStatus(t *testing.T) {
	svc := NewService(nil)
	if _, err := svc.SetStatus(context.Background(), "run", "not-a-status"); !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("want ErrInvalidStatus, got %v", err)
	}
}

func TestResultStatusWhitelist(t *testing.T) {
	for _, s := range []string{"passed", "failed", "skipped", "blocked"} {
		if !resultStatuses[s] {
			t.Errorf("%q should be an allowed result status", s)
		}
	}
	if resultStatuses["not_run"] || resultStatuses["nonsense"] {
		t.Error("only the four execution outcomes should be markable")
	}
}

func TestCIComment(t *testing.T) {
	url := "https://gitlab.example.com/pipelines/42"
	comment := "job green"

	got := *ciComment("gitlab-ci", &url, &comment)
	want := "[gitlab-ci " + url + "] " + comment
	if got != want {
		t.Errorf("full comment: want %q, got %q", want, got)
	}

	if got := *ciComment("gitlab-ci", nil, nil); got != "[gitlab-ci]" {
		t.Errorf("bare comment: want %q, got %q", "[gitlab-ci]", got)
	}
}
