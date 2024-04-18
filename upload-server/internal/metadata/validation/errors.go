package validation

import (
	"errors"
	"fmt"
)

var (
	ErrFailure = errors.New("validation failure")
	ErrWarning = errors.New("validation warning")
)

type ErrorMissing struct {
	Field string
}

func (e *ErrorMissing) Error() string {
	return fmt.Sprintf("field %s was missing", e.Field)
}

type ErrorNotAnAllowedValue struct {
	field string
	value string
}

func (e *ErrorNotAnAllowedValue) Error() string {
	return fmt.Sprintf("%s had disallowed value %s", e.field, e.value)
}
