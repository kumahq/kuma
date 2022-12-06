package validators

import (
	"fmt"
	"strings"
)

const (
	HasToBeGreaterThan                = "must be greater than"
	MustNotBeEmpty                    = "must not be empty"
	MustBeDefined                     = "must be defined"
	MustNotBeDefined                  = "must not be defined"
	WhenDefinedHasToBeNonNegative     = "must not be negative when defined"
	WhenDefinedHasToBeGreaterThanZero = "must be greater than zero when defined"
)

func MustHaveOnlyOne(entity string, allowedValues ...string) string {
	return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}
