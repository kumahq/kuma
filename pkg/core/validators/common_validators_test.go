package validators_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func validationError(msg string) validators.ValidationError {
	return validators.ValidationError{
		Violations: []validators.Violation{{
			Field:   "path",
			Message: msg,
		}},
	}
}

var _ = Describe("Common validators", func() {
	path := validators.RootedAt("path")
	invalidOtelAttributeName := validationError("must start with a lowercase letter, use only lowercase letters, digits, '.' or '_', avoid consecutive delimiters, and end with a letter or digit")

	DescribeTable("ValidateBandwidth",
		func(input string, expected validators.ValidationError) {
			// when
			actual := validators.ValidateBandwidth(path, input)

			// then
			Expect(actual).To(Equal(expected))
		},
		Entry("sanity", "1kbps", validators.ValidationError{}),
		Entry("without number", "Mbps", validators.ValidationError{}),
		Entry("not exact match", "1bpsp", validationError(validators.MustHaveBPSUnit)),
		Entry("bps is not allowed", "1bps", validationError(validators.MustHaveBPSUnit)),
		Entry("float point number is not supported", "0.1kbps", validationError(validators.MustHaveBPSUnit)),
		Entry("not defined", "", validationError(validators.MustBeDefined)),
	)

	DescribeTable("ValidateOtelAttributeName",
		func(input string, expected validators.ValidationError) {
			// when
			actual := validators.ValidateOtelAttributeName(path, input)

			// then
			Expect(actual).To(Equal(expected))
		},
		Entry("single segment", "mesh", validators.ValidationError{}),
		Entry("dotted segments", "service.version", validators.ValidationError{}),
		Entry("underscored segment", "process_command_args", validators.ValidationError{}),
		Entry("mixed dotted and underscored segments", "process.command_args", validators.ValidationError{}),
		Entry("placeholder key", "%KUMA_ZONE%", validationError("must be a static OpenTelemetry attribute name; placeholders are only supported in values")),
		Entry("reserved prefix", "otel.attribute", validationError("must not use the reserved OpenTelemetry prefix 'otel.'")),
		Entry("space", "my custom attribute", invalidOtelAttributeName),
		Entry("uppercase segment", "service.Version", invalidOtelAttributeName),
		Entry("hyphenated segment", "request-id", invalidOtelAttributeName),
		Entry("leading digit", "1service", invalidOtelAttributeName),
		Entry("consecutive delimiters", "service..version", invalidOtelAttributeName),
		Entry("mixed consecutive delimiters", "service._version", invalidOtelAttributeName),
		Entry("trailing dot", "service.", invalidOtelAttributeName),
		Entry("trailing underscore", "service_", invalidOtelAttributeName),
	)
})
