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
		Entry("external service", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: 192.168.0.1:8080
            tags:
              kuma.io/service: backend
              version: "1"`,
		),
		Entry("external service with valid tags", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: 192.168.0.1:8080
            tags:
              kuma.io/service: backend
              kuma.io/valid: abc.0123-789.under_score:90`,
		),
		Entry("external service domain name in the address", `
            type: ExternalService
            name: es-1
            mesh: default
            networking:
              address: example.com:8080
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
		Entry("networking.address: empty", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address:
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
                  address: ..>_<..:8080
                tags:
                  kuma.io/service: backend
                  version: "1"`,
			expected: `
                violations:
                - field: networking.address
                  message:  address has to be valid IP address or domain name`,
		}),
		Entry("networking: port out of range", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:65536
                tags:
                  kuma.io/service: backend
                  version: "1"`,
			expected: `
                violations:
                - field: networking.address
                  message: port has to be in range of [1, 65535]`,
		}),
		Entry("tags: empty service tag", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  version: "1"`,
			expected: `
                violations:
                - field: tags["kuma.io/service"]
                  message: tag has to exist`,
		}),
		Entry("tags: empty tag value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  kuma.io/service: backend
                  version:`,
			expected: `
                violations:
                - field: tags["version"]
                  message: tag value cannot be empty`,
		}),
		Entry("tags: `protocol` tag with an empty value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  kuma.io/service: backend
                  kuma.io/protocol:`,
			expected: `
                violations:
                - field: tags["kuma.io/protocol"]
                  message: tag value cannot be empty
                - field: tags["kuma.io/protocol"]
                  message: 'tag "kuma.io/protocol" has an invalid value "". Allowed values: grpc, http, http2, tcp'
`,
		}),
		Entry("tags: `protocol` tag with unsupported value", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  kuma.io/service: backend
                  kuma.io/protocol: not-yet-supported-protocol`,
			expected: `
                violations:
                - field: tags["kuma.io/protocol"]
                  message: 'tag "kuma.io/protocol" has an invalid value "not-yet-supported-protocol". Allowed values: grpc, http, http2, tcp'`,
		}),
		Entry("tags: tag name with invalid characters", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  kuma.io/service: backend
                  version: "v1"
                  inv@lidT/gN%me: value`,
			expected: `
                violations:
                - field: tags["inv@lidT/gN%me"]
                  message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`,
		}),
		Entry("tags: tag value with invalid characters", testCase{
			dataplane: `
                type: ExternalService
                name: es-1
                mesh: default
                networking:
                  address: 192.168.0.1:8080
                tags:
                  kuma.io/service: backend
                  version: "v1"
                  invalidTagValue: inv@lid+t@g`,
			expected: `
                violations:
                - field: tags["invalidTagValue"]
                  message: tag value must consist of alphanumeric characters, dots, dashes and underscores`,
		}),
	)

})
