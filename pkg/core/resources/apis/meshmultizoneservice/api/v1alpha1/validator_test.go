package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	. "github.com/kumahq/kuma/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		meshmzservice_api.NewMeshMultiZoneServiceResource,
		ErrorCases(
			"spec errors",
			[]validators.Violation{
				{
					Field:   "spec.selector.meshService.matchLabels",
					Message: "cannot be empty",
				},
			},
			``),
	)

	DescribeValidCases(
		meshmzservice_api.NewMeshMultiZoneServiceResource,
		Entry(
			"full spec",
			`
selector:
  meshService:
    matchLabels:
      app: xyz
`),
	)
})
