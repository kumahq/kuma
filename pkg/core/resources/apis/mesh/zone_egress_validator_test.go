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
			zoneegress := core_mesh.NewZoneEgressResource()

			// when
			err := util_proto.FromYAML([]byte(dpYAML), zoneegress.Spec)

			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			err = zoneegress.Validate()

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
	)

	type testCase struct {
		dataplane string
		expected  string
	}
	DescribeTable("should validate all fields and return as much individual errors as possible",
		func(given testCase) {
			// setup
			zoneegress := core_mesh.NewZoneEgressResource()

			// when
			err := util_proto.FromYAML([]byte(given.dataplane), zoneegress.Spec)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := zoneegress.Validate()
			// and
			actual, err := yaml.Marshal(verr)

			// then
			Expect(err).ToNot(HaveOccurred())
			// and
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
	)

})
