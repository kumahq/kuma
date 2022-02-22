package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Dataplane", func() {

	DescribeTable("should pass validation",
		func(dpYAML string) {
			// given
			zoneingress := core_mesh.NewZoneIngressResource()

			// when
			err := util_proto.FromYAML([]byte(dpYAML), zoneingress.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = zoneingress.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("with advertised address and port", `
            type: ZoneIngress
            name: zi-1
            networking:
              address: 192.168.0.1
              advertisedAddress: 10.0.0.1
              port: 10001
              advertisedPort: 1234
            availableServices:
              - tags:
                  kuma.io/service: backend
                  version: "1"
                  region: us
              - tags:
                  kuma.io/service: web
                  version: v2
                  region: eu`,
		),
		Entry("with advertised ipv6 address and port", `
            type: ZoneIngress
            name: zi-1
            networking:
              address: 192.168.0.1
              advertisedAddress: ::ffff:0a00:0001
              port: 10001
              advertisedPort: 1234
            availableServices:
              - tags:
                  kuma.io/service: backend
                  version: "1"
                  region: us
              - tags:
                  kuma.io/service: web
                  version: v2
                  region: eu`,
		),
		// no advertised address and port is valid because we may be waiting for Kubernetes to reconcile it
		Entry("without advertised address and port", `
            type: ZoneIngress
            name: zi-1
            networking:
              address: 192.168.0.1
              port: 10001
            availableServices: []`,
		),
		Entry("with admin port equal to port but different network interfaces", `
            type: ZoneIngress
            name: zi-1
            networking:
              admin:
                port: 10001
              address: 192.168.0.1
              advertisedAddress: 10.0.0.1
              port: 10001
              advertisedPort: 1234
            availableServices:
              - tags:
                  kuma.io/service: backend
                  version: "1"
                  region: us
              - tags:
                  kuma.io/service: web
                  version: v2
                  region: eu`,
		),
	)

	type testCase struct {
		dataplane string
		expected  string
	}
	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// setup
			zoneingress := core_mesh.NewZoneIngressResource()

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), zoneingress.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := zoneingress.Validate()
			// and
			actual, err := yaml.Marshal(verr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("no port defined", testCase{
			dataplane: `
            type: ZoneIngress
            name: zi-1
            networking:
              address: 192.168.0.1
              advertisedAddress: 10.0.0.1
              advertisedPort: 1234
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
                - field: networking.port
                  message: port must be in the range [1, 65535]`,
		}),
		Entry("invalid advertised address, port and available service", testCase{
			dataplane: `
            type: ZoneIngress
            name: zi-1
            networking:
              address: 192.168.0.1
              advertisedAddress: "!@#"
              port: 10001
              advertisedPort: 100000
            availableServices:
              - tags:
                  kuma.io/service: backend
                  version: "1"
                  region: us
              - tags:
                  kuma.io/service: web
                  version: v2
                  region: eu
              - tags:
                  version: v2
                  region: eu
              - tags:
                  kuma.io/service: ""
              - tags:
                  version: ""`,
			expected: `
                violations:
                - field: networking.advertisedAddress.address
                  message: address has to be valid IP address or domain name
                - field: networking.advertisedPort
                  message: port must be in the range [1, 65535]
                - field: availableService[2].tags
                  message: mandatory tag "kuma.io/service" is missing
                - field: availableService[3].tags["kuma.io/service"]
                  message: tag value must be non-empty
                - field: availableService[4].tags["version"]
                  message: tag value must be non-empty
                - field: availableService[4].tags
                  message: mandatory tag "kuma.io/service" is missing`,
		}),
		Entry("admin port equal to port", testCase{
			dataplane: `
            type: ZoneIngress
            name: zi-1
            networking:
              admin:
                port: 10001
              address: 127.0.0.1
              advertisedAddress: 10.0.0.1
              port: 10001
              advertisedPort: 1234
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
                - field: networking.admin.port
                  message: must differ from port`,
		}),
	)

})
