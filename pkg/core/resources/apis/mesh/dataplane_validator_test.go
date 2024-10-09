package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {
	DescribeTable("should pass validation",
		func(dpYAML string) {
			// given
			dataplane := core_mesh.NewDataplaneResource()

			// when
			err := util_proto.FromYAML([]byte(dpYAML), dataplane.Spec)

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
		Entry("dataplane with full inbounds and outbounds ipv6", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 0:0:0:0:0:FFFF:C0A8:0001
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: ::1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: ::1
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
                  tags:
                    kuma.io/service: redis`,
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
                  tags:
                    kuma.io/service: redis`,
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
		Entry("dataplane with service probes", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  serviceProbe:
                    interval: 1s
                    unhealthyThreshold: 5
                    tcp: {}
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis`,
		),
		Entry("dataplane with admin port", `
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
                    kuma.io/service: redis`),
		Entry("dataplane with admin port equal to inbound and outbound but different network interfaces", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              admin:
                port: 8080
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: redis`),
		Entry("dataplane with metrics backend defined and unique aggregate entries", `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              admin:
                port: 8080
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: redis
            metrics:
              type: prometheus
              conf:
                aggregate:
                - name: app
                  port: 123
                  path: "/stats"
                - name: sidecar
                  port: 999
                  path: "/metrics"`),
		Entry("dataplane with metrics backend type defined", `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  admin:
                    port: 8080
                  inbound:
                    - port: 8080
                      servicePort: 7777
                      address: 192.168.0.1
                      tags:
                        kuma.io/service: backend
                        version: "1"
                  outbound:
                    - port: 3333
                      address: 192.168.0.1
                      tags:
                        kuma.io/service: redis
                metrics:
                  type: prometheus`),
		Entry("dataplane with backend ref for MeshService", `
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
                  backendRef:
                    kind: MeshService
                    name: xyz
                    port: 80`,
		),
		Entry("dataplane with backend ref for MeshExternalService", `
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
                  backendRef:
                    kind: MeshExternalService
                    name: xyz`,
		),
		Entry("dataplane with reachableBackendRefs", `
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
              transparentProxying:
                reachableBackends:
                  refs:
                  - kind: MeshService
                    name: a
                    port: 9090
                  - kind: MeshExternalService
                    name: es
                    namespace: es1
                  - kind: MeshService
                    labels:
                      kuma.io/test: abc`,
		),
		Entry("dataplane with backend ref with labels", `
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
                  backendRef:
                    kind: MeshService
                    labels:
                      kuma.io/display-name: redis
                    port: 8080`,
		),
	)

	type testCase struct {
		dataplane string
		expected  string
	}
	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// setup
			dataplane := core_mesh.NewDataplaneResource()

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), dataplane.Spec)
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
		Entry("networking.address: can't use 0.0.0.0", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 0.0.0.0
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.address
                  message: 'must not be 0.0.0.0 or ::'`,
		}),
		Entry("networking.address: can't use ::", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: "::"
                  inbound:
                    - port: 8080
                      tags:
                        kuma.io/service: backend`,
			expected: `
                violations:
                - field: networking.address
                  message: 'must not be 0.0.0.0 or ::'`,
		}),
		Entry("networking: not enough inbound interfaces and no gateway", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  outbound:
                    - port: 3333
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking
                  message: has to contain at least one inbound interface or gateway`,
		}),
		Entry("missing networking", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default`,
			expected: `
                violations:
                - field: networking
                  message: must be defined`,
		}),
		Entry("networking empty", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking: {}`,
			expected: `
                violations:
                - field: networking
                  message: has to contain at least one inbound interface or gateway
                - field: networking.address
                  message: address can't be empty`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.inbound
                  message: inbound cannot be defined for delegated gateways`,
		}),
		Entry("networking: builtin gateway must not have inbounds", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    type: BUILTIN
                    tags:
                      kuma.io/service: kong
                  inbound:
                    - port: 3333
                      tags:
                        kuma.io/service: kong`,
			expected: `
                violations:
                - field: networking.inbound
                  message: inbound cannot be defined for builtin gateways`,
		}),
		Entry("networking: builtin gateway must not have outbounds", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    type: BUILTIN
                    tags:
                      kuma.io/service: kong
                  outbound:
                    - port: 3333
                      tags:
                        kuma.io/service: kong`,
			expected: `
                violations:
                - field: networking.outbound
                  message: outbound cannot be defined for builtin gateways`,
		}),
		Entry("networking: builtin gateway must not have probes", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    type: BUILTIN
                    tags:
                      kuma.io/service: kong
                probes:
                  port: 0
                  endpoints:
                   - inboundPort: 8088
                     inboundPath: /healthz
                     path: /8080/healthz`,
			expected: `
                violations:
                - field: networking.probes
                  message: probes cannot be defined for builtin gateways`,
		}),
		Entry("networking: builtin gateway must have a service tag", testCase{
			dataplane: `
                type: Dataplane
                name: dp-1
                mesh: default
                networking:
                  address: 192.168.0.1
                  gateway:
                    type: BUILTIN
                    tags:
                      foo: bar`,
			expected: `
                violations:
                - field: networking.gateway.tags
                  message: mandatory tag "kuma.io/service" is missing`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].port
                  message: port must be in the range [1, 65535]
                - field: networking.inbound[1].port
                  message: port must be in the range [1, 65535]`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].servicePort
                  message: port must be in the range [1, 65535]`,
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
                      tags:
                        kuma.io/service: redis`,
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
                      tags:
                        kuma.io/service: redis`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["version"]'
                  message: tag value must be non-empty`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "". Allowed values: grpc, http, http2, kafka, tcp'
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: tag value must be non-empty`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: 'networking.inbound[0].tags["kuma.io/protocol"]'
                  message: 'tag "kuma.io/protocol" has an invalid value "not-yet-supported-protocol". Allowed values: grpc, http, http2, kafka, tcp'`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.gateway.tags
                  message: mandatory tag "kuma.io/service" is missing`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["version"]'
                  message: tag value must be non-empty`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: 'networking.gateway.tags["kuma.io/protocol"]'
                  message: other values than tcp are not allowed, provided value "http"`,
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
                - field: networking.outbound[0].tags
                  message: mandatory tag "kuma.io/service" is missing`,
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
                - field: networking.outbound[0].tags
                  message: mandatory tag "kuma.io/service" is missing`,
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
                    - tags:
                        kuma.io/service: redis
                    - port: 65536
                      tags:
                        kuma.io/service: elastic`,
			expected: `
                violations:
                - field: networking.outbound[0].port
                  message: port must be in the range [1, 65535]
                - field: networking.outbound[1].port
                  message: port must be in the range [1, 65535]`,
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
                      tags:
                        kuma.io/service: elastic`,
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
                      tags:
                        kuma.io/service: elastic`,
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
                      tags:
                        kuma.io/service: redis`,
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
                      tags:
                        kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].tags["invalidTagValue"]
                  message: tag value must consist of alphanumeric characters, dots, dashes and underscores`,
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
		Entry("dataplane with virtual probe", testCase{
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
                  message: port must be in the range [1, 65535]
                - field: probes.endpoints[1].inboundPort
                  message: port must be in the range [1, 65535]
                - field: probes.endpoints[1].inboundPath
                  message: should be a valid URL Path
                - field: probes.endpoints[1].path
                  message: should be a valid URL Path
                - field: probes.endpoints[2].inboundPath
                  message: should be a valid URL Path
                - field: probes.endpoints[2].path
                  message: should be a valid URL Path`,
		}),
		Entry("dataplane with service probe", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  serviceProbe:
                    timeout: 1s
                    interval: "0"
                    healthyThreshold: 5
                    unhealthyThreshold: 0
                    tcp: {}
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: redis`,
			expected: `
                violations:
                - field: networking.inbound[0].serviceProbe.interval
                  message: must have a positive value
                - field: networking.inbound[0].serviceProbe.unhealthyThreshold
                  message: must have a positive value`,
		}),
		Entry("dataplane with admin port equal to inbound", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              admin:
                port: 8080
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
			expected: `
                violations:
                - field: networking.admin.port
                  message: must differ from inbound`,
		}),
		Entry("dataplane with admin port equal to outbound", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              admin:
                port: 3333
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
			expected: `
                violations:
                - field: networking.admin.port
                  message: must differ from outbound`,
		}),
		Entry("dataplane with duplicate metrics entry", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: redis
            metrics:
              type: prometheus
              conf:
                aggregate:
                - name: app
                  port: 123
                  path: "/stats"
                - name: app
                  port: 999
                  path: "/metrics"
                - name: duplicate-app
                  port: 12366
                  path: "/duplicate"
                - name: duplicate-app
                  port: 12345
                  path: "/other"  `,
			expected: `
                violations:
                - field: metrics.conf.aggregate[1].name
                  message: 'duplicate entry: app, values have to be unique'
                - field: metrics.conf.aggregate[3].name
                  message: 'duplicate entry: duplicate-app, values have to be unique'`,
		}),
		Entry("dataplane with not supported metrics type", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: redis
            metrics:
              type: custom-backend`,
			expected: `
                violations:
                - field: metrics.type
                  message: 'unknown backend type. Available backends: "prometheus"'`,
		}),
		Entry("dataplane with empty backend ref", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  backendRef: {}`,
			expected: `
                violations:
                - field: networking.outbound[0].backendRef.kind
                  message: 'invalid value. Available values are: MeshExternalService,MeshMultiZoneService,MeshService'
                - field: networking.outbound[0].backendRef
                  message:  either 'name' or 'labels' should be specified
                - field: networking.outbound[0].backendRef.port
                  message: port must be in the range [1, 65535]`,
		}),
		Entry("backend ref clashes with tags or service and missing port", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              outbound:
                - port: 3333
                  tags:
                    kuma.io/service: xyz
                  backendRef:
                    kind: MeshService
                    name: xyz
                    port: 8080
                - port: 3334
                  tags:
                    service: xyz
                  backendRef:
                    kind: MeshService
                    name: xyz`,
			expected: `
                violations:
                - field: networking.outbound[0].backendRef
                  message: both backendRef and tags/service cannot be defined
                - field: networking.outbound[1].backendRef.port
                  message: port must be in the range [1, 65535]
                - field: networking.outbound[1].backendRef
                  message: both backendRef and tags/service cannot be defined`,
		}),
		Entry("transparent proxy with reachable backend refs", testCase{
			dataplane: `
            type: Dataplane
            name: dp-1
            mesh: default
            networking:
              address: 192.168.0.1
              inbound:
                - port: 8080
                  servicePort: 7777
                  address: 192.168.0.1
                  tags:
                    kuma.io/service: backend
                    version: "1"
              transparentProxying:
                reachableBackends:
                  refs:
                  - kind: Something
                    name: first
                    labels:
                      kuma.io/test: test
                  - kind: MeshService
                    name: second
                    namespace: not-valid
                    labels:
                      kuma.io/test: test
                  - kind: MeshService
                    name: third
                    labels:
                      kuma.io/test: test
                  - kind: MeshService
                    name: first$-.kuma
                  - kind: MeshService
                  - kind: MeshService
                    namespace: xyz
`,
			expected: `
                violations:
                - field: networking.transparentProxing.reachableBackends.refs[0].kind
                  message: 'invalid value. Available values are: MeshExternalService,MeshMultiZoneService,MeshService'
                - field: networking.transparentProxing.reachableBackends.refs[0].labels
                  message: labels cannot be defined when name is specified
                - field: networking.transparentProxing.reachableBackends.refs[1].labels
                  message: labels cannot be defined when name is specified
                - field: networking.transparentProxing.reachableBackends.refs[2].labels
                  message: labels cannot be defined when name is specified
                - field: networking.transparentProxing.reachableBackends.refs[3].name
                  message: invalid characters. A lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character
                - field: networking.transparentProxing.reachableBackends.refs[4].name
                  message: name or labels are required
                - field: networking.transparentProxing.reachableBackends.refs[5].name
                  message: name is required, when namespace is defined`,
		}),
	)
})
