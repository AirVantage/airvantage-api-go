package airvantage

import (
	"fmt"
)

// AvError represents errors returned by the API.
type AvError struct {
	Path       string
	Code       string
	Parameters string
}

func avError(action, code, parameters string) error {
	return &AvError{Path: action, Code: code, Parameters: parameters}
}

func (e *AvError) Error() string {
	return fmt.Sprintf("%s  %s: %s", e.Path, e.Code, e.Parameters)
}
