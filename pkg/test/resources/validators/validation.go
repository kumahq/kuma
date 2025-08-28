package validators

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

// ResourceGenerator creates a resource of a pre-defined type.
type ResourceGenerator interface {
	New() core_model.Resource
}

// ResourceValidationCase captures a resource YAML and any corresponding validation error.
type ResourceValidationCase struct {
	Resource   string
	Name       string
	Violations []validators.Violation
	Labels     map[string]string
}

// DescribeValidCases creates a Ginkgo table test for the given entries,
// where each entry is a valid YAML resource. It ensures that each entry
// can be successfully validated.
func DescribeValidCases[T core_model.Resource](generator func() T, cases ...TableEntry) {
	DescribeTable(
		"should pass validation",
		func(anyGiven any) {
			var given ResourceValidationCase
			switch c := anyGiven.(type) {
			case string:
				given = ResourceValidationCase{
					Resource: c,
				}
			case ResourceValidationCase:
				given = c
			default:
				panic("invalid DescribeValidCases case")
			}
			// setup
			resource := generator()

			// when
			err := core_model.FromYAML([]byte(given.Resource), resource.GetSpec())
			if given.Name != "" || len(given.Labels) > 0 {
				resource.SetMeta(&test_model.ResourceMeta{
					Name:   given.Name,
					Labels: given.Labels,
				})
			}

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

			if given.Name != "" || len(given.Labels) > 0 {
				resource.SetMeta(&test_model.ResourceMeta{
					Name:   given.Name,
					Labels: given.Labels,
				})
			}

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
func ErrorCase(description string, err validators.Violation, yaml string, labels map[string]string) TableEntry {
	return Entry(
		description,
		ResourceValidationCase{
			Violations: []validators.Violation{err},
			Resource:   yaml,
			Labels:     labels,
		},
	)
}

// FErrorCase is a helper that generates a focused table entry for DescribeErrorCases.
func FErrorCase(description string, err validators.Violation, yaml string, labels map[string]string) TableEntry {
	// you need to manually set this to FEntry because `make check` will fail because it will try to un-focus this
	// and there is no way to exclude things from `ginkgo unfocus`
	return Entry(
		description,
		ResourceValidationCase{
			Violations: []validators.Violation{err},
			Resource:   yaml,
			Labels:     labels,
		},
	)
}

func ErrorCases(description string, errs []validators.Violation, yaml string) TableEntry {
	return Entry(
		description,
		ResourceValidationCase{
			Violations: errs,
			Resource:   yaml,
		},
	)
}
