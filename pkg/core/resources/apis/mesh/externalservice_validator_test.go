package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ExternalService", func() {

	DescribeTable("should pass validation",
		func(dpYAML string) {
			// given
			externalService := &core_mesh.ExternalServiceResource{}

			// when
			err := util_proto.FromYAML([]byte(dpYAML), &externalService.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = externalService.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("external service with inbound", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  tags:
                    kuma.io/service: backend
                    version: "1"`,
		),
		Entry("external service with full inbound", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 127.0.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"`,
		),
		Entry("external service with valid tags", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 127.0.0.1
                  tags:
                    kuma.io/service: backend
                    kuma.io/valid: abc.0123-789.under_score:90`,
		),
		Entry("external service domain name in the address", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: example.com
              inbound:
                - port: 8080
                  tags:
                    kuma.io/service: backend
                    version: "1"`,
		),
	)

	type testCase struct {
		dataplane string
		expected  string
	}
	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// setup
			externalService := core_mesh.ExternalServiceResource{}

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), &externalService.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := externalService.Validate()
			// and
			actual, err := yaml.Marshal(verr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("networking: not enough inbound interfaces", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking
                  message: has to contain at least one inbound interface`,
		}),
		Entry("networking.address: empty", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend
                        version: "1"`,
			expected: `
                violations:
                - field: networking.address
                  message: address can't be empty`,
		}),
		Entry("networking.address: invalid format", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: ..>_<..
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend
                        version: "1"`,
			expected: `
                violations:
                - field: networking.address
                  message:  address has to be valid IP address or domain name`,
		}),
		Entry("networking.inbound: port out of range", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - tags:
                        kuma.io/service: backend
                        version: "1"
                    - port: 65536
                      tags:
                        kuma.io/service: sub-backend`,
			expected: `
                violations:
                - field: networking.inbound[0].port
                  message: port has to be in range of [1, 65535]
                - field: networking.inbound[1].port
                  message: port has to be in range of [1, 65535]`,
		}),
		Entry("networking.inbound: servicePort out of the range", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      servicePort: 65536
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.inbound[0].servicePort
                  message: servicePort has to be in range of [0, 65535]`,
		}),
		Entry("networking.inbound: invalid address", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      address: invalid-address
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.inbound[0].address
                  message: address has to be valid IP address`,
		}),
		Entry("networking.inbound: empty service tag", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        version: "v1"`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["kuma.io/service"]
                  message: tag has to exist`,
		}),
		Entry("networking.inbound: empty tag value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        kuma.io/service: backend
                        version:`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["version"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.inbound: `protocol` tag with an empty value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        kuma.io/service: backend
                        kuma.io/protocol:`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "". Allowed values: grpc, http, http2, tcp'
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.inbound: `protocol` tag with unsupported value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        kuma.io/service: backend
                        kuma.io/protocol: not-yet-supported-protocol`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "not-yet-supported-protocol". Allowed values: grpc, http, http2, tcp'`,
		}),
		Entry("networking.inbound: tag name with invalid characters", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        kuma.io/service: backend
                        version: "v1"
                        inv@lidT/gN%me: value`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["inv@lidT/gN%me"]
                  message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`,
		}),
		Entry("networking.inbound: tag value with invalid characters", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        kuma.io/service: backend
                        version: "v1"
                        invalidTagValue: inv@lid+t@g`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["invalidTagValue"]
                  message: tag value must consist of alphanumeric characters, dots, dashes and underscores`,
		}),
		Entry("inbound service address", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 10001
                      serviceAddress: 192.168.0.2
                      servicePort: 5050
                      address: 1.1.1.1
                      tags:
                        kuma.io/service: backend`,
		}),
		Entry("inbound service address invalid", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 10001
                      serviceAddress: INVALID
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.inbound[0].serviceAddress
                  message: serviceAddress has to be valid IP address`,
		}),
		Entry("inbound service address overlap address", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 10001
                      serviceAddress: 192.168.0.1
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.inbound[0].serviceAddress
                  message: serviceAddress and servicePort has to differ from address and port`,
		}),
		Entry("inbound service address overlap inbound address", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 10001
                      address: 192.168.0.2
                      serviceAddress: 192.168.0.2
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.inbound[0].serviceAddress
                  message: serviceAddress and servicePort has to differ from address and port`,
		}),
		Entry("inbound service address different inbound address and port", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 10001
                      address: 192.168.0.2
                      serviceAddress: 192.168.0.2
                      servicePort: 10002
                      tags:
                        kuma.io/service: backend`,
		}),
	)

})
