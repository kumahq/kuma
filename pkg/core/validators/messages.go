package validators

import (
	"fmt"
	"strings"
)

const (
	HasToBeGreaterThan                = "must be greater than"
	HasToBeLessThan                   = "must be less than"
	HasToBeGreaterOrEqualThen         = "must be greater or equal then"
	HasToBeGreaterThanZero            = "must be greater than 0"
	MustNotBeEmpty                    = "must not be empty"
	MustBeDefined                     = "must be defined"
	MustBeSet                         = "must be set"
	MustNotBeSet                      = "must not be set"
	MustNotBeDefined                  = "must not be defined"
	MustBeDefinedAndGreaterThanZero   = "must be defined and greater than zero"
	WhenDefinedHasToBeNonNegative     = "must not be negative when defined"
	WhenDefinedHasToBeGreaterThanZero = "must be greater than zero when defined"
	HasToBeInRangeFormat              = "must be in inclusive range [%v, %v]"
	WhenDefinedHasToBeValidPath       = "must be a valid path when defined"
	StringHasToBeValidNumber          = "string must be a valid number"
	MustHaveBPSUnit                   = "must be in kbps/Mbps/Gbps units"
)

var (
	HasToBeInPercentageRange     = fmt.Sprintf(HasToBeInRangeFormat, "0.0", "100.0")
	HasToBeInUintPercentageRange = fmt.Sprintf(HasToBeInRangeFormat, 0, 100)
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
