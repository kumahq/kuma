package validators

import (
	"fmt"
)

type ValidationError struct {
	Violations []Violation
}

type Violation struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	msg := ""
	for _, violation := range v.Violations {
		if msg != "" {
			msg = fmt.Sprintf("%s; %s: %s", msg, violation.Field, violation.Message)
		} else {
			msg += fmt.Sprintf("%s: %s", violation.Field, violation.Message)
		}
	}
	return msg
}

func (v *ValidationError) HasViolations() bool {
	return len(v.Violations) > 0
}

func (v *ValidationError) OrNil() error {
	if v.HasViolations() {
		return v
	}
	return nil
}

func (v *ValidationError) AddViolation(field string, message string) {
	violation := Violation{
		Field:   field,
		Message: message,
	}
	v.Violations = append(v.Violations, violation)
}

func (v *ValidationError) AddError(rootField string, validationErr ValidationError) {
	rootPrefix := ""
	if rootField != "" {
		rootPrefix += fmt.Sprintf("%s.", rootField)
	}
	for _, violation := range validationErr.Violations {
		newViolation := Violation{
			Field:   fmt.Sprintf("%s%s", rootPrefix, violation.Field),
			Message: violation.Message,
		}
		v.Violations = append(v.Violations, newViolation)
	}
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}
