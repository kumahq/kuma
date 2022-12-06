package validators

import (
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
	if value == nil || *value <= 0 {
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
