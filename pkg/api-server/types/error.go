package types

import "fmt"

type Error struct {
	Title   string  `json:"title"`
	Details string  `json:"details"`
	Causes  []Cause `json:"causes,omitempty"`
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("%s (%s)", e.Title, e.Details)
	for _, cause := range e.Causes {
		msg += fmt.Sprintf(";%s=%s ", cause.Field, cause.Message)
	}
	return msg
}

type Cause struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
