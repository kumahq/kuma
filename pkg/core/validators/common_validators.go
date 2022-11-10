package validators

import (
	"github.com/kumahq/kuma/pkg/config/types"
)

func ValidateDurationNotNegativeOrNil(path PathBuilder, duration *types.Duration) (err ValidationError) {
	if duration == nil {
		return
	}

	if duration.Seconds() < 0 || duration.Nanoseconds() < 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeNonNegative)
	}

	return
}

func ValidateDurationGreaterThanZeroOrNil(path PathBuilder, duration *types.Duration) (err ValidationError) {
	if duration == nil {
		return
	}

	if duration.Seconds() <= 0 && duration.Nanoseconds() <= 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThanZero)
	}

	return
}
