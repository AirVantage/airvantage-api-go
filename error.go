package airvantage

import (
	"fmt"
	"strings"
)

// AvError represents errors returned by the API.
type AvError struct {
	Path       string
	Code       string
	Parameters string
}

func avError(action, code string, parameters []string) error {
	params := strings.Join(parameters, ", ")
	return &AvError{Path: action, Code: code, Parameters: params}
}

func (e *AvError) Error() string {
	return fmt.Sprintf("%s  %s: %s", e.Path, e.Code, e.Parameters)
}
