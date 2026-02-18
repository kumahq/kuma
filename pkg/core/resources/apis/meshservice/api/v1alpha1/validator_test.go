package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
	. "github.com/kumahq/kuma/v2/pkg/test/resources/validators"
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
		Entry(
			"multiple selectors specified",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `spec.selector`,
					Message: `must specify only one of: dataplaneTags, dataplaneRef, or dataplaneLabels`,
				}},
				Name: "meshservice",
				Resource: `
selector:
  dataplaneTags:
    app: redis
  dataplaneLabels:
    matchLabels:
      app: redis
`,
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
		Entry(
			"accepts dataplaneTags selector",
			ResourceValidationCase{
				Name: "meshservice",
				Resource: `
selector:
  dataplaneTags:
    app: redis
`,
			},
		),
		Entry(
			"accepts dataplaneRef selector",
			ResourceValidationCase{
				Name: "meshservice",
				Resource: `
selector:
  dataplaneRef:
    name: redis-01
`,
			},
		),
		Entry(
			"accepts dataplaneLabels selector",
			ResourceValidationCase{
				Name: "meshservice",
				Resource: `
selector:
  dataplaneLabels:
    matchLabels:
      app: redis
`,
			},
		),
	)
})
