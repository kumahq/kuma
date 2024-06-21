package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("MeshService", func() {
	DescribeErrorCases(
		api.NewMeshServiceResource,
		Entry(
			"name too long",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `name`,
					Message: `must not be longer than 63 characters`,
				}},
				Name:     "meshservice-too-long-too-long-too-long-too-long-too-long-too-long-too-long-too-long-too-long",
				Resource: "",
			},
		),
	)
	DescribeValidCases(
		api.NewMeshServiceResource,
		Entry(
			"accepts valid resource",
			ResourceValidationCase{
				Name:     "meshservice",
				Resource: "",
			},
		),
	)
})
