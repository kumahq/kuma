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

func ValidateDurationGreaterThanZero(path PathBuilder, duration *k8s.Duration) (err ValidationError) {
	if duration == nil || duration.Duration <= 0 {
		err.AddViolationAt(path, MustBeDefinedAndGreaterThanZero)
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

func ValidateValueGreaterThanZero(path PathBuilder, value *int32) (err ValidationError) {
	if value == nil || *value == 0 {
		err.AddViolationAt(path, MustBeDefinedAndGreaterThanZero)
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

func ValidateUintPercentageOrNil(path PathBuilder, percentage *int32) (err ValidationError) {
	if percentage == nil {
		return
	}

	if *percentage < 0 || *percentage > 100 {
		err.AddViolationAt(path, HasToBeInUintPercentageRange)
	}

	return
}

func ValidateStringDefined(path PathBuilder, value string) (err ValidationError) {
	if value == "" {
		err.AddViolationAt(path, MustBeDefined)
	}

	return
}
