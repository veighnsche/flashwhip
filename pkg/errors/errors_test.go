package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorFormatting(t *testing.T) {
	err1 := New(ErrCodeConfigInvalid, "invalid argument")
	if err1.Error() != "[FW-1001] invalid argument" {
		t.Errorf("unexpected error format: %s", err1.Error())
	}

	cause := errors.New("file not found")
	err2 := Wrap(ErrCodeToolFileNotFound, "cannot read file", cause)
	if err2.Error() != "[FW-5002] cannot read file: file not found" {
		t.Errorf("unexpected wrapped error format: %s", err2.Error())
	}

	err3 := Wrapf(ErrCodeDBOpenFailed, cause, "db %s failed", "init")
	if err3.Error() != "[FW-2001] db init failed: file not found" {
		t.Errorf("unexpected wrapf error format: %s", err3.Error())
	}
}

func TestErrorUnwrapAndIs(t *testing.T) {
	cause := errors.New("underlying cause")
	err := Wrap(ErrCodeAgentBuildFailed, "agent build failed", cause)

	if !errors.Is(err, cause) {
		t.Errorf("expected errors.Is to match cause error")
	}

	targetErr := New(ErrCodeAgentBuildFailed, "")
	if !Is(err, targetErr) {
		t.Errorf("expected Is to match error code")
	}

	if GetCode(err) != ErrCodeAgentBuildFailed {
		t.Errorf("expected code %d, got %d", ErrCodeAgentBuildFailed, GetCode(err))
	}

	if GetCode(fmt.Errorf("wrapped: %w", err)) != ErrCodeAgentBuildFailed {
		t.Errorf("expected code %d from nested fmt.Errorf, got %d", ErrCodeAgentBuildFailed, GetCode(fmt.Errorf("wrapped: %w", err)))
	}
}

func TestRegistryLookup(t *testing.T) {
	spec, ok := Lookup(ErrCodeDBOpenFailed)
	if !ok {
		t.Fatalf("expected ErrCodeDBOpenFailed to exist in registry")
	}
	if spec.Category != "Database" {
		t.Errorf("expected Category 'Database', got %s", spec.Category)
	}

	all := All()
	if len(all) == 0 {
		t.Errorf("expected All() to return registered specs")
	}

	// Verify all codes in registry have non-empty fields
	for _, s := range all {
		if s.Code <= 0 {
			t.Errorf("invalid code %d in registry", s.Code)
		}
		if s.Name == "" || s.Category == "" || s.Description == "" || s.Remedy == "" {
			t.Errorf("incomplete error spec for code FW-%04d (%s)", s.Code, s.Name)
		}
	}
}
