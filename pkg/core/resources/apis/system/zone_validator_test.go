package system_test

import (
	"fmt"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Zone", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(zoneYAML string) {
				// setup
				zone := system.ZoneResource{}

				// when
				err := util_proto.FromYAML([]byte(zoneYAML), &zone.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())
				fmt.Printf("%v", zone.Spec)

				// when
				verr := zone.Validate()
				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("valid zone", `
            ingress:
              address: 192.168.0.2:10001`),
		)

		type testCase struct {
			zone     string
			expected string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				zone := system.ZoneResource{}

				// when
				err := util_proto.FromYAML([]byte(given.zone), &zone.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := zone.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("spec: empty", testCase{
				zone: ``,
				expected: `
               violations:
                 - field: address
                   message: cannot be empty`}),
			Entry("wrong format", testCase{
				zone: `
               ingress:
                 address: 192.168.0.2`,
				expected: `
               violations:
                 - field: address
                   message: "invalid address: address 192.168.0.2: missing port in address"`}),
			Entry("spec: empty", testCase{
				zone: ``,
				expected: `
               violations:
                 - field: address
                   message: cannot be empty`}),
			Entry("url instead of address", testCase{
				zone: `
               ingress:
                 address: grpcs://192.168.0.2:1234`,
				expected: `
               violations:
                 - field: address
                   message: should not be URL. Expected format is hostname:port`}),
		)
	})
})
