package validators

import (
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ValidateDurationNotNegativeOrNil(path PathBuilder, duration *k8s.Duration) (err ValidationError) {
	if duration == nil {
		return
	}

	if duration.Duration < 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeNonNegative)
	}

	return
}

func ValidateDurationGreaterThanZeroOrNil(path PathBuilder, duration *k8s.Duration) (err ValidationError) {
	if duration == nil {
		return
	}

	if duration.Duration <= 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThanZero)
	}

	return
}

func ValidateValueGreaterThanZeroOrNil(path PathBuilder, value *uint32) (err ValidationError) {
	if value == nil {
		return
	}

	if *value == 0 {
		err.AddViolationAt(path, WhenDefinedHasToBeGreaterThanZero)
	}
	return
}

func ValidatePercentageOrNil(path PathBuilder, percentage *float32) (err ValidationError) {
    if percentage == nil {
        return
    }

    if *percentage < 0.0 || *percentage > 100.0 {
        err.AddViolationAt(path, HasToBeInPercentageRange)
    }

    return
}

func ValidateStringDefined(path PathBuilder, value string) (err ValidationError) {
	if value == "" {
		err.AddViolationAt(path, MustBeDefined)
	}

	return
}
