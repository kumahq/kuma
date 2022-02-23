package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ZoneEgress", func() {

	DescribeTable("should pass validation",
		func(yaml string) {
			// given
			zoneEgress := core_mesh.NewZoneEgressResource()

			// when
			err := util_proto.FromYAML([]byte(yaml), zoneEgress.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = zoneEgress.Validate()

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("with address and port", `
            type: ZoneEgress
            name: ze-1
            networking:
              address: 192.168.0.1
              port: 10001`,
		),
		Entry("with admin port equal to port but different network interfaces", `
            type: ZoneEgress
            name: ze-1
            networking:
              admin:
                port: 10001
              address: 192.168.0.1
              port: 10001`,
		),
	)

	type testCase struct {
		dataplane string
		expected  string
	}

	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// given
			zoneEgress := core_mesh.NewZoneEgressResource()

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), zoneEgress.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := zoneEgress.Validate()
			actual, err := yaml.Marshal(verr)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("no port defined", testCase{
			dataplane: `
            type: ZoneEgress
            name: ze-1
            networking:
              address: 192.168.0.1`,
			expected: `
                violations:
                - field: networking.port
                  message: port must be in the range [1, 65535]`,
		}),
		Entry("admin port equal to port", testCase{
			dataplane: `
            type: ZoneEgress
            name: ze-1
            networking:
              admin:
                port: 10001
              address: 127.0.0.1
              port: 10001`,
			expected: `
                violations:
                - field: networking.admin.port
                  message: must differ from port`,
		}),
	)
})
