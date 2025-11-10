package errors

import "errors"

var (
	
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrBlueprintNotFound    = errors.New("blueprint not found")
	ErrBlueprintNameMismatch = errors.New("blueprint name mismatch")
	ErrMissingRequiredFields = errors.New("missing required fields")
	ErrInvalidRequest       = errors.New("invalid request")
)
