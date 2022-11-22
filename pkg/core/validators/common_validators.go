package validators

import (
	"fmt"
	"time"

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

func ValidateDurationGreaterThan(path PathBuilder, duration *k8s.Duration, minDuration time.Duration) (err ValidationError) {
	if duration == nil {
		err.AddViolationAt(path, MustBeDefined)
		return
	}

	if duration.Duration <= minDuration {
		err.AddViolationAt(path, fmt.Sprintf("%s: %s", HasToBeGreaterThan, minDuration))
	}

	return
}

func ValidateIntegerGreaterThanZeroOrNil(path PathBuilder, value *uint32) (err ValidationError) {
	if value == nil {
		return
	}

	ValidateIntegerGreaterThan(path, *value, 0)
	return
}

func ValidateIntegerGreaterThan(path PathBuilder, value uint32, minValue uint32) (err ValidationError) {
	if value <= minValue {
		err.AddViolationAt(path, HasToBeGreaterThan)
	}

	return
}
