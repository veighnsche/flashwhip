package errors

import (
	"errors"
	"fmt"
)

// Re-export standard library errors functions for convenience.
var (
	Is     = errors.Is
	As     = errors.As
	Unwrap = errors.Unwrap
)

// Error represents a structured error with a stable numeric error code.
type Error struct {
	Code    Code
	Message string
	Err     error
}

// Error formats the error into a clean string with the stable error code (e.g. "[FW-1001] invalid config").
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Code > 0 {
		if e.Err != nil {
			return fmt.Sprintf("[FW-%04d] %s: %v", e.Code, e.Message, e.Err)
		}
		return fmt.Sprintf("[FW-%04d] %s", e.Code, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying cause error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is returns true if target is an *Error matching e.Code.
func (e *Error) Is(target error) bool {
	if target == nil {
		return e == nil
	}
	var t *Error
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// New creates a new Error with a stable numeric code and message.
func New(code Code, msg string) *Error {
	return &Error{
		Code:    code,
		Message: msg,
	}
}

// Errorf creates a new Error with a stable numeric code and formatted message.
func Errorf(code Code, format string, args ...any) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with a stable numeric code and context message.
func Wrap(code Code, msg string, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// Wrapf wraps an existing error with a stable numeric code and formatted context message.
func Wrapf(code Code, err error, format string, args ...any) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// GetCode extracts the numeric error code from an error chain, returning ErrCodeUnknown (0) if none is found.
func GetCode(err error) Code {
	if err == nil {
		return 0
	}
	var customErr *Error
	if errors.As(err, &customErr) {
		return customErr.Code
	}
	return ErrCodeUnknown
}
