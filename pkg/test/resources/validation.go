package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// ResourceGenerator creates a resource of a pre-defined type.
type ResourceGenerator interface {
	New() core_model.Resource
}

// ResourceValidationCase captures a resource YAML and any corresponding validation error.
type ResourceValidationCase struct {
	Resource   string
	Violations []validators.Violation
}

// DescribeValidCases creates a Ginkgo table test for the given entries,
// where each entry is a valid YAML resource. It ensures that each entry
// can be successfully validated.
func DescribeValidCases[T core_model.Resource](generator func() T, cases ...TableEntry) {
	DescribeTable(
		"should pass validation",
		func(given string) {
			// setup
			resource := generator()

			// when
			err := core_model.FromYAML([]byte(given), resource.GetSpec())

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := core_model.Validate(resource)

			// then
			Expect(verr).ToNot(HaveOccurred())
		},
		cases)
}

// DescribeErrorCases creates a Ginkgo table test for the given entries, where each entry
// is a ResourceValidationCase that contains an invalid resource YAML and the corresponding
// validation error.
func DescribeErrorCases[T core_model.Resource](generator func() T, cases ...TableEntry) {
	DescribeTable(
		"should validate all fields and return as many individual errors as possible",
		func(given ResourceValidationCase) {
			// setup
			resource := generator()

			// when
			Expect(
				core_model.FromYAML([]byte(given.Resource), resource.GetSpec()),
			).ToNot(HaveOccurred())

			expected := validators.ValidationError{
				Violations: given.Violations,
			}

			// then
			err := core_model.Validate(resource)
			Expect(err).To(HaveOccurred())
			verr := err.(*validators.ValidationError)
			Expect(verr.Violations).To(ConsistOf(expected.Violations))
		},
		cases,
	)
}

// ErrorCase is a helper that generates a table entry for DescribeErrorCases.
func ErrorCase(description string, err validators.Violation, yaml string) TableEntry {
	return Entry(
		description,
		ResourceValidationCase{
			Violations: []validators.Violation{err},
			Resource:   yaml,
		},
	)
}

func ErrorCases(description string, errs []validators.Violation, yaml string) TableEntry {
	GinkgoHelper()
	return Entry(
		description,
		ResourceValidationCase{
			Violations: errs,
			Resource:   yaml,
		},
	)
}
