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
		ErrorCase("negative port",
			validators.Violation{
				Field:   "spec.port",
				Message: "port must be a valid (1-65535)",
			}, `
address: 192.168.0.1
port: -1
`),
		ErrorCase("unspecified address",
			validators.Violation{
				Field:   "spec.address",
				Message: "must not be 0.0.0.0 or ::",
			}, `
address: 0.0.0.0
port: 10001
`),
		ErrorCase("address is neither an IP nor a domain name",
			validators.Violation{
				Field:   "spec.address",
				Message: "must be a valid IP address or domain name",
			}, `
address: "http://10.0.0.1:10001"
port: 10001
`),
	)

	DescribeValidCases(
		meshzoneaddress_api.NewMeshZoneAddressResource,
		Entry(
			"IPv4 address",
			`
address: 192.168.0.1
port: 10001
`),
		Entry(
			"IPv6 address",
			`
address: "2001:db8::1"
port: 65535
`),
		Entry(
			"hostname",
			`
address: a1b2c3.elb.us-east-1.amazonaws.com
port: 1
`),
	)
})
