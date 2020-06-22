package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {

	DescribeTable("should pass validation",
		func(dpYAML string) {
			// given
			dataplane := &core_mesh.DataplaneResource{}

			// when
			err := util_proto.FromYAML([]byte(dpYAML), &dataplane.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = dataplane.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("dataplane with inbounds", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  tags:
                    service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    service: redis`,
		),
		Entry("dataplane with full inbounds and outbounds", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 127.0.0.1
                  tags:
                    service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 127.0.0.1
                  tags:
                    service: redis`,
		),
		Entry("dataplane with legacy outbounds", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 127.0.0.1
                  tags:
                    service: backend
              outbound:
                - port: 3333
                  address: 127.0.0.1
                  service: redis`,
		),
		Entry("dataplane with gateway", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              gateway:
                tags:
                  service: backend
                  version: "1"
              outbound:
                - port: 3333
                  service: redis`,
		),
		Entry("dataplane with valid tags", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              gateway:
                tags:
                  service: backend
                  version: "1"
                  kuma.io/valid: abc.0123-789.under_score:90
              outbound:
                - port: 3333
                  tags:
                    service: redis`,
		),
		Entry("dataplane in ingress mode", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
                address: 192.168.0.1
                ingress:
                  availableServices:
                    - tags:
                        service: backend
                        version: "1"
                        region: us
                    - tags:
                        service: web
                        version: v2
                        region: eu
                inbound:
                  - port: 10001`,
		),
	)

	type testCase struct {
		dataplane string
		expected  string
	}
	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// setup
			dataplane := core_mesh.DataplaneResource{}

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), &dataplane.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := dataplane.Validate()
			// and
			actual, err := yaml.Marshal(verr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("networking: not enough inbound interfaces and no gateway", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking
                  message: has to contain at least one inbound interface or gateway`,
		}),
		Entry("networking: both inbounds and gateway are defined", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 8080
                      servicePort: 7777
                      tags:
                        service: backend
                        version: "1"
                  gateway:
                    tags:
                      service: kong
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking
                  message: inbound cannot be defined both with gateway`,
		}),
		Entry("networking: invalid address", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: invalid
                  inbound:
                    - port: 8080
                      servicePort: 7777
                      tags:
                        service: backend
                        version: "1"
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.address
                  message: address has to be valid IP address`,
		}),
		Entry("networking.inbound: port of the range", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - tags:
                        service: backend
                        version: "1"
                    - port: 65536
                      tags:
                        service: sub-backend
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].port
                  message: port has to be in range of [1, 65535]
                - field: networking.inbound[1].port
                  message: port has to be in range of [1, 65535]`,
		}),
		Entry("networking.inbound: servicePort out of the range", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      servicePort: 65536
                      tags:
                        service: backend
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].servicePort
                  message: servicePort has to be in range of [0, 65535]`,
		}),
		Entry("networking.inbound: invalid address", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      address: invalid-address
                      tags:
                        service: backend
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].address
                  message: address has to be valid IP address`,
		}),
		Entry("networking.inbound: empty service tag", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        version: "v1"
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["service"]
                  message: tag has to exist`,
		}),
		Entry("networking.inbound: empty tag value", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version:
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["version"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.inbound: `protocol` tag with an empty value", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        protocol:
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["protocol"]'
                  message: 'tag "protocol" has an invalid value "". Allowed values: http, tcp'
                - field: 'networking.inbound[0].tags["protocol"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.inbound: `protocol` tag with unsupported value", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        protocol: not-yet-supported-protocol
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["protocol"]'
                  message: 'tag "protocol" has an invalid value "not-yet-supported-protocol". Allowed values: http, tcp'`,
		}),
		Entry("networking.gateway: empty service tag", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    tags:
                      version: "v1"
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["service"]'
                  message: tag has to exist`,
		}),
		Entry("networking.gateway: empty tag value", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    tags:
                      service: backend
                      version:
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["version"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.outbound: empty service tag", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                  outbound:
                    - port: 3333`,
			expected: `
                violations:
                - field: networking.outbound[0].service
                  message: cannot be empty`,
		}),
		Entry("networking.outbound: empty service tag", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                  outbound:
                    - port: 3333
                      tags:
                        version: v1`,
			expected: `
                violations:
                - field: networking.outbound[0].tags["service"]
                  message: tag has to exist`,
		}),
		Entry("networking.outbound: port out of the range", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                  outbound:
                    - service: redis
                    - port: 65536
                      service: elastic`,
			expected: `
                violations:
                - field: networking.outbound[0].port
                  message: port has to be in range of [1, 65535]
                - field: networking.outbound[1].port
                  message: port has to be in range of [1, 65535]`,
		}),
		Entry("networking.outbound: invalid address", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                  outbound:
                    - port: 3333
                      address: invalid
                      service: elastic`,
			expected: `
                violations:
                - field: networking.outbound[0].address
                  message: address has to be valid IP address`,
		}),
		Entry("networking.outbound: invalid address", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                  outbound:
                    - port: 3333
                      address: invalid
                      service: elastic`,
			expected: `
                violations:
                - field: networking.outbound[0].address
                  message: address has to be valid IP address`,
		}),
		Entry("networking.inbound: tag name with invalid characters", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                        inv@lidT/gN%me: value
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["inv@lidT/gN%me"]
                  message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`,
		}),
		Entry("networking.inbound: tag value with invalid characters", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  inbound:
                    - port: 1234
                      tags:
                        service: backend
                        version: "v1"
                        invalidTagValue: inv@lid+t@g
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["invalidTagValue"]
                  message: tag value must consist of alphanumeric characters, dots, dashes and underscores`,
		}),
		Entry("networking.ingress: outbound is not empty", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  ingress:
                    availableServices:
                      - tags:
                          service: backend
                          version: "1"
                          region: us
                      - tags:
                          service: web
                          version: v2
                          region: eu
                  inbound:
                    - port: 10001
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking
                  message: dataplane cannot have outbounds in the ingress mode`,
		}),
		Entry("networking.ingress: gateway defined", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  ingress:
                    availableServices:
                      - tags:
                          service: backend
                          version: "1"
                          region: us
                      - tags:
                          service: web
                          version: v2
                          region: eu
                  gateway: {}
                  inbound:
                    - port: 10001`,
			expected: `
                violations:
                - field: networking
                  message: gateway cannot be defined in the ingress mode`,
		}),
		Entry("networking.ingress: no inbound defined", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  ingress:
                    availableServices:
                      - tags: 
                          service: backend
                          version: "1"
                          region: us
                      - tags:
                          service: web
                          version: v2
                          region: eu`,
			expected: `
                violations:
                - field: networking
                  message: dataplane must have one inbound interface`,
		}),
		Entry("networking.ingress: inbound with redundant fields", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  ingress:
                    availableServices:
                      - tags: 
                          service: backend
                          version: "1"
                          region: us
                      - tags:
                          service: web
                          version: v2
                          region: eu
                  inbound:
                    - port: 10001
                      servicePort: 5050
                      address: 1.1.1.1
                      tags:
                        name: ingress-dp`,
			expected: `
                violations:
                - field: networking.inbound[0].servicePort
                  message: cannot be defined in the ingress mode
                - field: networking.inbound[0].address
                  message: cannot be defined in the ingress mode`,
		}),
	)

})
