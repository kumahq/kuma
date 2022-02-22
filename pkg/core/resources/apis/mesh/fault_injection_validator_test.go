package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("FaultInjection", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(faultInjectionYAML string) {
				// setup
				faultInjection := NewFaultInjectionResource()

				// when
				err := util_proto.FromYAML([]byte(faultInjectionYAML), faultInjection.Spec)
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
                    kuma.io/protocol: http
                destinations:
                - match:
                    service: backend
                    kuma.io/protocol: http
                    region: eu
                    kuma.io/valid: abcd.123-456.under_score_.:80
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
			Entry("http2", `
                sources:
                - match:
                    service: frontend
                destinations:
                - match:
                    service: backend
                    kuma.io/protocol: http2
                conf:
                  delay:
                    percentage: 50
                    value: 10ms`),
			Entry("match any", `
                sources:
                - match:
                    service: frontend
                    kuma.io/protocol: http
                destinations:
                - match:
                    service: backend
                    kuma.io/protocol: grpc
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
				faultInjection := NewFaultInjectionResource()

				// when
				err := util_proto.FromYAML([]byte(given.faultInjection), faultInjection.Spec)
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
                   kuma.io/protocol: http
                destinations:
                - match:
                   service: backend
                   kuma.io/protocol: http
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
                   kuma.io/protocol: http
                destinations:
                - match:
                   service: backend
                   kuma.io/protocol: http
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
                   kuma.io/protocol: http
                destinations:
                - match:
                   service: backend
                   kuma.io/protocol: http
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
                   kuma.io/protocol: http
                destinations:
                - match:
                   service: backend
                   kuma.io/protocol: http
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
                   kuma.io/protocol: http
                destinations:
                - match:
                   service: backend
                   kuma.io/protocol: http
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
			Entry("kuma.io/protocol: not specified", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
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
               - field: destinations[0].match
                 message: protocol must be specified`}),
			Entry("kuma.io/protocol: wrong protocol", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                destinations:
                - match:
                    service: backend
                    region: eu
                    kuma.io/protocol: tcp
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: destinations[0].match["kuma.io/protocol"]
                 message: must be one of the [http, http2, grpc]`}),
			Entry("tag value: invalid character set", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                    kuma.io/protocol: http
                    invalidTag: v@/u^e
                destinations:
                - match:
                    service: backend
                    kuma.io/protocol: http
                    invalidTag: v@/u^e#!
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: sources[0].match["invalidTag"]
                 message: tag value must consist of alphanumeric characters, dots, dashes, slashes and underscores or be "*"
               - field: destinations[0].match["invalidTag"]
                 message: tag value must consist of alphanumeric characters, dots, dashes, slashes and underscores or be "*"`}),
			Entry("tag name: invalid character set", testCase{
				faultInjection: `
                sources:
                - match:
                    service: frontend
                    kuma.io/protocol: http
                    inv@lidT@g#: value
                destinations:
                - match:
                    service: backend
                    kuma.io/protocol: http
                    inv@lidT@g#: value
                conf:
                  responseBandwidth:
                    limit: 50mbps
                    percentage: 100`,
				expected: `
               violations:
               - field: sources[0].match["inv@lidT@g#"]
                 message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores
               - field: destinations[0].match["inv@lidT@g#"]
                 message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`}),
		)
	})
})
