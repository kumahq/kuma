package validators

import (
	"fmt"
	"strings"
)

const (
	MustNotBeEmpty   = "must not be empty"
	MustBeDefined    = "must be defined"
	MustNotBeDefined = "must not be defined"
)

func MustHaveOnlyOne(entity string, allowedValues ...string) string {
	return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}
