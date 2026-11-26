package apperr

import (
	"errors"
	"testing"
)

func TestAppErr_Error_WithMessage(t *testing.T) {
	err := &AppErr{
		Message: "custom message",
	}

	if got := err.Error(); got != "custom message" {
		t.Fatalf("expected message %q, got %q", "custom message", got)
	}
}

func TestAppErr_Error_WithServiceError(t *testing.T) {
	baseErr := errors.New("base error")
	err := &AppErr{
		ServiceError: baseErr,
	}

	if got := err.Error(); got != "base error" {
		t.Fatalf("expected %q, got %q", "base error", got)
	}
}

func TestAppErr_Error_Unknown(t *testing.T) {
	err := &AppErr{}

	if got := err.Error(); got != "unknown error" {
		t.Fatalf("expected %q, got %q", "unknown error", got)
	}
}

func TestAppErr_Is_Match(t *testing.T) {
	err := &AppErr{
		ServiceError: ErrNotFound,
	}

	if !err.Is(ErrNotFound) {
		t.Fatalf("expected Is(ErrNotFound) == true")
	}
}

func TestAppErr_Is_NoMatch(t *testing.T) {
	err := &AppErr{
		ServiceError: ErrNotFound,
	}

	if err.Is(ErrAlreadyExists) {
		t.Fatalf("expected Is(ErrAlreadyExists) == false")
	}
}

func TestAppErr_Is_NilServiceError(t *testing.T) {
	err := &AppErr{}

	if err.Is(ErrNotFound) {
		t.Fatalf("expected Is() == false when ServiceError is nil")
	}
}

func TestErrorsIs_WorksWithAppErr(t *testing.T) {
	err := &AppErr{
		ServiceError: ErrInvalidStatus,
	}

	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected errors.Is(AppErr, ErrInvalidStatus) == true")
	}
}

func TestErrorsIs_FalseForDifferentError(t *testing.T) {
	err := &AppErr{
		ServiceError: ErrInvalidStatus,
	}

	if errors.Is(err, ErrInvalidPhone) {
		t.Fatalf("expected errors.Is(AppErr, ErrInvalidPhone) == false")
	}
}
