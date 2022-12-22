package validators

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
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

func ValidateDurationGreaterThanZero(path PathBuilder, duration k8s.Duration) (err ValidationError) {
	if duration.Duration <= 0 {
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

func ValidateValueGreaterThanZero(path PathBuilder, value int32) (err ValidationError) {
	if value <= 0 {
		err.AddViolationAt(path, MustBeDefinedAndGreaterThanZero)
	}
	return
}

func ValidateIntPercentageOrNil(path PathBuilder, percentage *int32) (err ValidationError) {
	if percentage == nil {
		return
	}

	if *percentage < 0 || *percentage > 100 {
		err.AddViolationAt(path, HasToBeInUintPercentageRange)
	}

	return
}

func ValidateUInt32PercentageOrNil(path PathBuilder, percentage *uint32) (err ValidationError) {
	if percentage == nil {
		return
	}

	if *percentage > 100 {
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

func ValidatePathOrNil(path PathBuilder, filePath *string) (err ValidationError) {
	if filePath == nil {
		return
	}

	isFilePath, _ := govalidator.IsFilePath(*filePath)
	if !isFilePath {
		err.AddViolationAt(path, WhenDefinedHasToBeValidPath)
	}

	return
}

func ValidateStatusCode(path PathBuilder, status int32) (err ValidationError) {
	if status < 100 || status >= 600 {
		err.AddViolationAt(path, "must be in range [100, 600)")
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

	return ValidateIntegerGreaterThan(path, *value, 0)
}

func ValidateIntegerGreaterThan(path PathBuilder, value uint32, minValue uint32) (err ValidationError) {
	if value <= minValue {
		err.AddViolationAt(path, fmt.Sprintf("%s %d", HasToBeGreaterThan, minValue))
	}

	return
}
