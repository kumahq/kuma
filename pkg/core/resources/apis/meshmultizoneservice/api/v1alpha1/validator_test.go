package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	. "github.com/kumahq/kuma/v3/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		meshmzservice_api.NewMeshMultiZoneServiceResource,
		ErrorCase("spec.template empty",
			validators.Violation{
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
`),
		ErrorCase("spec.template empty",
			validators.Violation{
				Field:   `spec.ports[0].appProtocol`,
				Message: `appProtocol must be one of: grpc, http, http2, tcp`,
			}, `
selector:
  meshService:
    matchLabels:
      app: xyz
ports:
- port: 123
  appProtocol: not_supported
`),
		Entry(
			"name does not conform to RFC 1035",
			ResourceValidationCase{
				Violations: []validators.Violation{{
					Field:   `name`,
					Message: `a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')`,
				}},
				Name: "global.backend",
				Resource: `
selector:
  meshService:
    matchLabels:
      app: xyz
ports:
- port: 123
  appProtocol: tcp
`,
			},
		),
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
