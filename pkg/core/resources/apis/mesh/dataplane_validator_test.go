package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis`,
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
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 127.0.0.1
                  tags:
                    kuma.io/service: redis`,
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
                    kuma.io/service: backend
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
                  kuma.io/service: backend
                  kuam.io/protocol: tcp
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
                  kuma.io/service: backend
                  version: "1"
                  kuma.io/valid: abc.0123-789.under_score:90
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis`,
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
                        kuma.io/service: backend
                        version: "1"
                        region: us
                    - tags:
                        kuma.io/service: web
                        version: v2
                        region: eu
                inbound:
                  - port: 10001`,
		),
		Entry("dataplane domain name in the address", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: example.com
              inbound:
                - port: 8080
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis`,
		),
		Entry("dataplane in ingress mode with protocol tag", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
                address: 192.168.0.1
                ingress:
                  availableServices:
                    - tags:
                        kuma.io/service: backend
                inbound:
                  - port: 10001
                    tags:
                      kuma.io/protocol: tcp`,
		),
		Entry("dataplane with probes", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis
            probes:
              port: 9000
              endpoints:
               - inboundPort: 8088
                 inboundPath: /healthz
                 path: /8080/healthz`,
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
		Entry("networking.address: empty", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend
                        version: "1"
                  outbound:
                    - port: 3333
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.address
                  message: address can't be empty`,
		}),
		Entry("networking.address: invalid format", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: ..>_<..
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend
                        version: "1"
                  outbound:
                    - port: 3333
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.address
                  message:  address has to be valid IP address or domain name`,
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
                        kuma.io/service: backend
                        version: "1"
                  gateway:
                    tags:
                      kuma.io/service: kong
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: networking
                  message: inbound cannot be defined both with gateway`,
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
                        kuma.io/service: backend
                        version: "1"
                    - port: 65536
                      tags:
                        kuma.io/service: sub-backend
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
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
                - field: networking.inbound[0].tags["kuma.io/service"]
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
                        kuma.io/protocol:
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "". Allowed values: grpc, http, http2, tcp'
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
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
                        kuma.io/service: backend
                        kuma.io/protocol: not-yet-supported-protocol
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "not-yet-supported-protocol". Allowed values: grpc, http, http2, tcp'`,
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
                - field: 'networking.gateway.tags["kuma.io/service"]'
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
                      kuma.io/service: backend
                      version:
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["version"]'
                  message: tag value cannot be empty`,
		}),
		Entry("networking.gateway: protocol http", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    tags:
                      kuma.io/service: backend
                      kuma.io/protocol: http
                  outbound:
                    - port: 3333
                      service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["kuma.io/protocol"]'
                  message: other values than TCP are not allowed`,
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
                        kuma.io/service: backend
                        version: "v1"
                  outbound:
                    - port: 3333`,
			expected: `
                violations:
                - field: networking.outbound[0].kuma.io/service
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
                        kuma.io/service: backend
                        version: "v1"
                  outbound:
                    - port: 3333
                      tags:
                        version: v1`,
			expected: `
                violations:
                - field: networking.outbound[0].tags["kuma.io/service"]
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
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
                        kuma.io/service: backend
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
                          kuma.io/service: backend
                          version: "1"
                          region: us
                      - tags:
                          kuma.io/service: web
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
                          kuma.io/service: backend
                          version: "1"
                          region: us
                      - tags:
                          kuma.io/service: web
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
                          kuma.io/service: backend
                          version: "1"
                          region: us
                      - tags:
                          kuma.io/service: web
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
                          kuma.io/service: backend
                          version: "1"
                          region: us
                      - tags:
                          kuma.io/service: web
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
		Entry("inbound service address", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
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
                type: Dataplane
                name: dp-1
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
                type: Dataplane
                name: dp-1
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
                type: Dataplane
                name: dp-1
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
                type: Dataplane
                name: dp-1
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
		Entry("inbound service address and ingress", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  ingress:
                    availableServices:
                      - tags: 
                          kuma.io/service: backend
                          version: "1"
                          region: us
                  inbound:
                    - port: 10001
                      serviceAddress: 192.168.0.2
                      servicePort: 5050
                      address: 1.1.1.1
                      tags:
                        name: ingress-dp
                        kuma.io/protocol: http`,
			expected: `
                violations:
                - field: networking.inbound[0].servicePort
                  message: cannot be defined in the ingress mode
                - field: networking.inbound[0].serviceAddress
                  message: cannot be defined in the ingress mode
                - field: networking.inbound[0].address
                  message: cannot be defined in the ingress mode
                - field: tags["kuma.io/protocol"]
                  message: other values than TCP are not allowed`,
		}),
		Entry("", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis
            probes:
              port: 0
              endpoints:
               - inboundPort: 8088
                 inboundPath: /healthz
                 path: /8080/healthz
               - inboundPort: 99999999
                 inboundPath: healthz
                 path: 8080/healthz
               - inboundPort: 1000
                 inboundPath: 
                 path: `,
			expected: `
                violations:
                - field: probes.port
                  message: port has to be in range of [1, 65535]
                - field: probes.endpoints[1].inboundPort
                  message: port has to be in range of [1, 65535]
                - field: probes.endpoints[1].inboundPath
                  message: should be a valid URL Path
                - field: probes.endpoints[1].path
                  message: should be a valid URL Path
                - field: probes.endpoints[2].inboundPath
                  message: should be a valid URL Path
                - field: probes.endpoints[2].path
                  message: should be a valid URL Path`,
		}),
	)

})
