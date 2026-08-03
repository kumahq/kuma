package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"

	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	. "github.com/kumahq/kuma/v3/pkg/test/resources/validators"
)

var _ = Describe("validation", func() {
	DescribeErrorCases(
		meshzoneaddress_api.NewMeshZoneAddressResource,
		ErrorCases(
			"empty spec",
			[]validators.Violation{
				{
					Field:   "spec.address",
					Message: "must not be empty",
				},
				{
					Field:   "spec.port",
					Message: "port must be a valid (1-65535)",
				},
			},
			``),
		ErrorCase("port out of range",
			validators.Violation{
				Field:   "spec.port",
				Message: "port must be a valid (1-65535)",
			}, `
address: 192.168.0.1
port: 70000
`),
	)

	DescribeValidCases(
		meshzoneaddress_api.NewMeshZoneAddressResource,
		Entry(
			"full spec",
			`
address: 192.168.0.1
port: 10001
`),
	)
})
