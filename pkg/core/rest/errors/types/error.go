package types

import "fmt"

// Error following https://kong-aip.netlify.app/aip/193/
type Error struct {
	// Type a unique identifier for this error.
	Type string `json:"type"`
	// Status The HTTP status code of the error.
	Status int `json:"status"`
	// Title A short, human-readable summary of the problem.
	// It should not change between occurrences of a problem, except for localization.
	// Should be provided as "Sentence case" for direct use in the UI
	Title string `json:"title"`
	// Detail A human readable explanation specific to this occurrence of the problem.
	Detail string `json:"detail"`
	// Instance Used to return the correlation ID back to the user if present
	Instance string `json:"instance,omitempty"`
	// InvalidParameters
	InvalidParameters []InvalidParameter `json:"invalid_parameters,omitempty"`

	// Deprecated
	Details string `json:"details"`
	// Deprecated
	Causes []Cause `json:"causes,omitempty"`
}

type InvalidParameter struct {
	Field   string   `json:"field"`
	Reason  string   `json:"reason"`
	Rule    string   `json:"rule,omitempty"`
	Choices []string `json:"choices,omitempty"`
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("%s (%s)", e.Title, e.Detail)
	for _, cause := range e.InvalidParameters {
		msg += fmt.Sprintf(";%s=%s ", cause.Field, cause.Reason)
	}
	return msg
}

type Cause struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
