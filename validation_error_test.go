package ess

import (
	"errors"
	"testing"
)

func TestValidationError_Ok_returnsTrueIfThereAreNoErrors(t *testing.T) {
	err := NewValidationError()
	if got, want := err.Ok(), true; got != want {
		t.Errorf(`err.Ok() = %v; want %v`, got, want)
	}
}

func TestValidationError_Ok_returnsFalseIfThereAreErrors(t *testing.T) {
	err := NewValidationError().Add("test", "error")
	if got, want := err.Ok(), false; got != want {
		t.Errorf(`err.Ok() = %v; want %v`, got, want)
	}
}

func TestValidationError_Merge_mergesOtherErrorTypeUnderKeyAll(t *testing.T) {
	testError := errors.New("test error")
	err := NewValidationError().Merge(testError)
	if got, want := len(err.Errors["$all"]), 1; got != want {
		t.Errorf(`len(err.Errors["$all"]) = %v; want %v`, got, want)
	}

	if got, want := err.Errors["$all"][0], testError.Error(); got != want {
		t.Errorf(`err.Errors["$all"][0] = %v; want %v`, got, want)
	}
}

func TestValidationError_Return_returnsNilIfErrorIsOk(t *testing.T) {
	err := NewValidationError()
	if got, want := err.Return(), (error)(nil); got != want {
		t.Errorf(`err.Return() = %v; want %v`, got, want)
	}
}

func TestValidationError_Return_returnsSelfIfErrorIsNotOk(t *testing.T) {
	err := NewValidationError().Add("field", "error")
	if got, want := err.Return(), err; got != want {
		t.Errorf(`err.Return() = %v; want %v`, got, want)
	}
}
