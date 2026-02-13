package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/workload/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
	. "github.com/kumahq/kuma/v2/pkg/test/resources/validators"
)

var _ = Describe("Workload", func() {
	DescribeErrorCases(
		api.NewWorkloadResource,
		Entry(
			"name too long",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `name`,
					Message: `must not be longer than 63 characters`,
				}},
				Name:     "workload-" + string(make([]byte, 63)),
				Resource: "",
			},
		),
	)
	DescribeValidCases(
		api.NewWorkloadResource,
		Entry(
			"accepts valid resource",
			ResourceValidationCase{
				Name:     "workload",
				Resource: "",
			},
		),
		Entry(
			"accepts max length name",
			ResourceValidationCase{
				Name:     string(make([]byte, 63)),
				Resource: "",
			},
		),
	)
})
