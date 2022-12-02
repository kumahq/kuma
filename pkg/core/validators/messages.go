package validators

import (
	"fmt"
	"strings"
)

const (
	MustNotBeEmpty                    = "must not be empty"
	MustBeDefined                     = "must be defined"
	MustNotBeDefined                  = "must not be defined"
	MustBeDefinedAndGreaterThanZero   = "must be defined and greater than zero"
	WhenDefinedHasToBeNonNegative     = "must not be negative when defined"
	WhenDefinedHasToBeGreaterThanZero = "must be greater than zero when defined"
	HasToBeInPercentageRange          = "has to be in [0.0 - 100.0] range"
	HasToBeInUintPercentageRange      = "has to be in [0 - 100] range"
)

func MustHaveOnlyOne(entity string, allowedValues ...string) string {
	return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}
