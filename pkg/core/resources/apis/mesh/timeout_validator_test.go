package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Timeouts", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(timeoutYAML string) {
				// setup
				timeout := NewTimeoutResource()

				// when
				err := util_proto.FromYAML([]byte(timeoutYAML), timeout.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := timeout.Validate()
				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full policy", `
                sources:
                - match:
                   kuma.io/service: frontend
                   kuma.io/protocol: http
                destinations:
                - match:
                   kuma.io/service: backend
                conf:
                  connectTimeout: 5s
                  tcp:
                    idleTimeout: 50s
                  http:
                    requestTimeout: 15s
                    idleTimeout: 1h
                  grpc:
                    streamIdleTimeout: 40s
                    maxStreamDuration: 30m`),
		)

		type testCase struct {
			timeout  string
			expected string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				timeout := NewTimeoutResource()

				// when
				err := util_proto.FromYAML([]byte(given.timeout), timeout.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := timeout.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("spec: empty", testCase{
				timeout: ``,
				expected: `
               violations:
               - field: sources
                 message: must have at least one element
               - field: destinations
                 message: must have at least one element
               - field: conf
                 message: has to be defined`}),
			Entry("conf.*: empty", testCase{
				timeout: `
                sources:
                - match:
                   kuma.io/service: frontend
                   kuma.io/protocol: http
                destinations:
                - match:
                   kuma.io/service: backend
                conf:
                  connectTimeout: 5s
                  tcp: {}
                  http: {}
                  grpc: {}`,
				expected: `
               violations:
               - field: conf.tcp
                 message: at least one timeout in section has to be defined
               - field: conf.http
                 message: at least one timeout in section has to be defined
               - field: conf.grpc
                 message: at least one timeout in section has to be defined`}),
		)
	})
})
