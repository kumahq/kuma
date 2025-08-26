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
		ErrorCase("spec.template empty", validators.Violation{
			Field:   `spec.ports[0].name`,
			Message: `must not be empty`,
		}, `
selector:
  meshService:
    matchLabels:
      app: xyz
ports:
- port: 123
  name: ''
  appProtocol: tcp
`, nil),
		ErrorCase("spec.template empty", validators.Violation{
			Field:   `spec.ports[0].appProtocol`,
			Message: `appProtocol must be one of: grpc, http, http2, kafka, tcp`,
		}, `
selector:
  meshService:
    matchLabels:
      app: xyz
ports:
- port: 123
  appProtocol: not_supported
`, nil),
		ErrorCases(
			"spec errors",
			[]validators.Violation{
				{
					Field:   "spec.selector.meshService.matchLabels",
					Message: "cannot be empty",
				},
				{
					Field:   "spec.ports",
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
ports:
- port: 123
  appProtocol: tcp
`),
	)
})
