package validators

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Violations []Violation `json:"violations"`
}

type Violation struct {
	Field   string `json:"field"`
	Message string `json:"message"`
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

func (v *ValidationError) AddViolationAt(path PathBuilder, message string) {
	v.AddViolation(path.String(), message)
}

func (v *ValidationError) AddViolation(field string, message string) {
	violation := Violation{
		Field:   field,
		Message: message,
	}
	v.Violations = append(v.Violations, violation)
}

func (v *ValidationError) Add(err ValidationError) {
	v.AddError("", err)
}

func (v *ValidationError) AddErrorAt(path PathBuilder, validationErr ValidationError) {
	v.AddError(path.String(), validationErr)
}

func (v *ValidationError) AddError(rootField string, validationErr ValidationError) {
	rootPrefix := ""
	if rootField != "" {
		rootPrefix += fmt.Sprintf("%s.", rootField)
	}
	for _, violation := range validationErr.Violations {
		field := ""
		if violation.Field == "" {
			field = rootField
		} else {
			field = fmt.Sprintf("%s%s", rootPrefix, violation.Field)
		}
		newViolation := Violation{
			Field:   field,
			Message: violation.Message,
		}
		v.Violations = append(v.Violations, newViolation)
	}
}

// Transform returns a new ValidationError with every violation
// transformed by a given transformFunc.
func (v *ValidationError) Transform(transformFunc func(Violation) Violation) *ValidationError {
	if v == nil {
		return nil
	}
	if transformFunc == nil || len(v.Violations) == 0 {
		return &(*v) // we want to guarantee that Transform always returns a new object
	}
	result := ValidationError{
		Violations: make([]Violation, len(v.Violations)),
	}
	for i := range v.Violations {
		result.Violations[i] = transformFunc(v.Violations[i])
	}
	return &result
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

type PathBuilder []string

func RootedAt(name string) PathBuilder {
	return PathBuilder{name}
}

func (p PathBuilder) Field(name string) PathBuilder {
	return append(p, fmt.Sprintf(".%s", name))
}

func (p PathBuilder) Index(index int) PathBuilder {
	return append(p, fmt.Sprintf("[%d]", index))
}

func (p PathBuilder) Key(key string) PathBuilder {
	return append(p, fmt.Sprintf("[%q]", key))
}

func (p PathBuilder) String() string {
	return strings.Join(p, "")
}
