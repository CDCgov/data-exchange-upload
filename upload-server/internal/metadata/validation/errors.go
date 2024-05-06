package validation

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrFailure  = errors.New("validation failure")
	ErrWarning  = errors.New("validation warning")
	ErrNotFound = errors.Join(ErrFailure, errors.New("manifest validation config file not found"))
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

type ValidationError struct {
	Err error
}

func (v *ValidationError) Error() string {
	return v.Err.Error()
}

func unwrap(e error) []error {
	errs := []error{}
	u, ok := e.(interface {
		Unwrap() []error
	})
	if ok {
		for _, err := range u.Unwrap() {
			errs = append(errs, unwrap(err)...)
		}
	} else {
		errs = append(errs, e)
		err := errors.Unwrap(e)
		if err != nil {
			errs = append(errs, unwrap(err)...)
		}
	}
	return errs
}

func (v *ValidationError) MarshalJSON() ([]byte, error) {
	errs := unwrap(v.Err)
	res := make([]any, len(errs))
	for i, e := range errs {
		res[i] = e.Error() // Fallback to the error string
	}
	return json.Marshal(res)
}
