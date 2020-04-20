package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("FaultInjection", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(faultInjectionYAML string) {
				// setup
				faultInjection := FaultInjectionResource{}

				// when
				err := util_proto.FromYAML([]byte(faultInjectionYAML), &faultInjection.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := faultInjection.Validate()
				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
                sources:
                - match:
                    service: frontend
                    protocol: http
                destinations:
                - match:
                    service: backend
                    protocol: http
                    region: eu
                    valid: abcd.123-456.under_score_.
                conf:
                  delay:
                    percentage: 50
                    value: 10ms
                  abort:
                    percentage: 40
                    httpStatus: 500
                  responseBandwidth:
                    percentage: 40
                    limit: 50kbps`),
			Entry("match any", `
                sources:
                - match:
                    service: frontend
                    protocol: http
                destinations:
                - match:
                    service: backend
                    protocol: http
                    region: "*"
                conf:
                  delay:
                    percentage: 50
                    value: 10ms
                  abort:
                    percentage: 40
                    httpStatus: 500
                  responseBandwidth:
                    percentage: 40
                    limit: 50kbps`),
		)

		type testCase struct {
			faultInjection string
			expected       string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				faultInjection := FaultInjectionResource{}

				// when
				err := util_proto.FromYAML([]byte(given.faultInjection), &faultInjection.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := faultInjection.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("spec: empty", testCase{
				faultInjection: ``,
				expected: `
               violations:
               - field: sources
                 message: must have at least one element
               - field: destinations
                 message: must have at least one element
               - field: conf
                 message: must have at least one of the faults configured`}),
			Entry("conf.*: empty", testCase{
				faultInjection: `
                sources:
                - match:
                   service: frontend
                   protocol: http
                destinations:
                - match:
                   service: backend
                   protocol: http
                   region: eu
                conf:
                  delay: {}
                  abort: {}
                  responseBandwidth: {}`,
				expected: `
               violations:
               - field: conf
                 message: must have at least one of the faults configured`}),
			Entry("conf.*.percentage: empty", testCase{
				faultInjection: `
                sources:
                - match:
                   service: frontend
                   protocol: http
                destinations:
                - match:
                   service: backend
                   protocol: http
                   region: eu
                conf:
                  delay:
                    value: 50ms
                  abort:
                    httpStatus: 500
                  responseBandwidth: 
                    limit: 50kbps`,
				expected: `
               violations:
               - field: conf.delay.percentage
                 message: cannot be empty
               - field: conf.abort.percentage
                 message: cannot be empty
               - field: conf.responseBandwidth.percentage
                 message: cannot be empty`}),
			Entry("conf.* main value: empty", testCase{
				faultInjection: `
                sources:
                - match:
                   service: frontend
                   protocol: http
                destinations:
                - match:
                   service: backend
                   protocol: http
                   region: eu
                conf:
                  delay:
                    percentage: 30
                  abort:
                    percentage: 30
                  responseBandwidth: 
                    percentage: 30`,
				expected: `
               violations:
               - field: conf.delay.value
                 message: cannot be empty
               - field: conf.abort.httpStatus
                 message: cannot be empty
               - field: conf.responseBandwidth.limit
                 message: cannot be empty`}),
			Entry("conf.abort: wrong format", testCase{
				faultInjection: `
                sources:
                - match:
                   service: frontend
                   protocol: http
                destinations:
                - match:
                   service: backend
                   protocol: http
                   region: eu
                conf:
                  abort:
                    httpStatus: 600
                    percentage: 100`,
				expected: `
               violations:
               - field: conf.abort.httpStatus
                 message: http status code is incorrect`}),
			Entry("conf.responseBandwidth: wrong format", testCase{
				faultInjection: `
                sources:
                - match:
                   service: frontend
                   protocol: http
                destinations:
                - match:
                   service: backend
                   protocol: http
                   region: eu
                conf:
                  responseBandwidth:
                    limit: 50mb
                    percentage: 101`,
				expected: `
               violations:
               - field: conf.responseBandwidth.percentage
                 message: has to be in [0.0 - 100.0] range
               - field: conf.responseBandwidth.limit
                 message: has to be in kbps/mbps/gbps units`}),
			Entry("protocol: wrong format", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                    protocol: tcp
                destinations:
                - match:
                    service: backend
                    region: eu
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: sources[0].match["protocol"]
                 message: must be one of the [http]
               - field: destinations[0].match
                 message: protocol must be specified`}),
			Entry("tag value: invalid character set", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                    protocol: http
                    invalidTag: v@/u^e
                destinations:
                - match:
                    service: backend
                    protocol: http
                    invalidTag: v@/u^e#!
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: sources[0].match["invalidTag"]
                 message: value must consist of alphanumeric characters, dots, dashes and underscores or be "*"
               - field: destinations[0].match["invalidTag"]
                 message: value must consist of alphanumeric characters, dots, dashes and underscores or be "*"`}),
			Entry("tag name: invalid character set", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                    protocol: http
                    inv@lidT@g#: value
                destinations:
                - match:
                    service: backend
                    protocol: http
                    inv@lidT@g#: value
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: sources[0].match["inv@lidT@g#"]
                 message: key must consist of alphanumeric characters, dots, dashes and underscores
               - field: destinations[0].match["inv@lidT@g#"]
                 message: key must consist of alphanumeric characters, dots, dashes and underscores`}),
		)
	})
})
