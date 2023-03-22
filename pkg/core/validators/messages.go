package validators

import (
	"fmt"
	"strings"
)

const (
	HasToBeGreaterThan                = "must be greater than"
	HasToBeGreaterThanZero            = "must be greater than 0"
	MustNotBeEmpty                    = "must not be empty"
	MustBeDefined                     = "must be defined"
	MustNotBeDefined                  = "must not be defined"
	MustBeDefinedAndGreaterThanZero   = "must be defined and greater than zero"
	WhenDefinedHasToBeNonNegative     = "must not be negative when defined"
	WhenDefinedHasToBeGreaterThanZero = "must be greater than zero when defined"
	HasToBeInPercentageRange          = "has to be in [0.0 - 100.0] range"
	HasToBeInUintPercentageRange      = "has to be in [0 - 100] range"
	WhenDefinedHasToBeValidPath       = "has to be a valid path when defined"
	StringHasToBeValidNumber          = "string has to be a valid number"
)

func MustHaveOnlyOne(entity string, allowedValues ...string) string {
	return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}

func MustHaveExactlyOneOf(entity string, allowedValues ...string) string {
	return fmt.Sprintf(`%s must have exactly one defined: %s`, entity, strings.Join(allowedValues, ", "))
}

func MustHaveAtLeastOne(allowedValues ...string) string {
	return fmt.Sprintf(`must have at least one defined: %s`, strings.Join(allowedValues, ", "))
}
