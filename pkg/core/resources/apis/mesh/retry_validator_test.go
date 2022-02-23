package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Retry", func() {
	Describe("Validate()", func() {
		type testCase struct {
			retry    string
			expected string
		}

		type testCaseWithNoErrors struct {
			retry string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				retry := NewRetryResource()

				// when
				err := util_proto.FromYAML([]byte(given.retry), retry.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := retry.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				retry: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
                - field: conf
                  message: has to be defined
`,
			}),
			Entry("selectors without tags", testCase{
				retry: `
                sources:
                - match: {}
                destinations:
                - match: {}
`,
				expected: `
                violations:
                - field: sources[0].match
                  message: must have at least one tag
                - field: sources[0].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: conf
                  message: has to be defined
`,
			}),
			Entry("selectors with empty tags values", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                destinations:
                - match:
                    kuma.io/service:
                    region:
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: conf
                  message: has to be defined`,
			}),
			Entry("empty conf", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf: {}
`,
				expected: `
                violations:
                - field: conf
                  message: missing protocol [grpc|http|tcp] configuration
`,
			}),
			Entry("empty conf.http", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    http: {}
`,
				expected: `
                violations:
                - field: conf.http
                  message: field cannot be empty
`,
			}),
			Entry("invalid conf.http", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    http:
                        numRetries: 0
                        perTryTimeout: "0"
                        backOff: {}
                        retriableMethods:
                        - NONE
`,
				expected: `
                violations:
                - field: conf.http.numRetries
                  message: has to be greater than 0 when defined
                - field: conf.http.perTryTimeout
                  message: has to be greater than 0 when defined
                - field: conf.http.backOff.baseInterval
                  message: has to be defined
                - field: conf.http.retriableMethods[0]
                  message: field cannot be empty
`,
			}),
			Entry("empty conf.grpc", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    grpc: {}
`,
				expected: `
                violations:
                - field: conf.grpc
                  message: field cannot be empty
`,
			}),
			Entry("invalid conf.grpc", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    grpc:
                        numRetries: 0
                        backOff: {}
                        perTryTimeout: "0"
                        retryOn:
                        - cancelled
                        - resource_exhausted
                        - cancelled
                        - cancelled
                        - resource_exhausted
                        - internal
                        - internal
                        - internal
`,
				expected: `
                violations:
                - field: conf.grpc.retryOn
                  message: repeated value "cancelled" at indexes [0, 2, 3]
                - field: conf.grpc.retryOn
                  message: repeated value "internal" at indexes [5, 6, 7]
                - field: conf.grpc.retryOn
                  message: repeated value "resource_exhausted" at indexes [1, 4]
                - field: conf.grpc.numRetries
                  message: has to be greater than 0 when defined
                - field: conf.grpc.perTryTimeout
                  message: has to be greater than 0 when defined
                - field: conf.grpc.backOff.baseInterval
                  message: has to be defined
`,
			}),
			Entry("conf.grpc.backOff.baseInterval equal 0", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    grpc:
                        backOff:
                            baseInterval: "0"
`,
				expected: `
                violations:
                - field: conf.grpc.backOff.baseInterval
                  message: has to be greater than 0
`,
			}),
			Entry("conf.grpc.backOff.maxInterval equal 0s", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    grpc:
                        backOff:
                            baseInterval: 20ms
                            maxInterval: 0s
`,
				expected: `
                violations:
                - field: conf.grpc.backOff.maxInterval
                  message: has to be greater than 0 when defined
`,
			}),
			Entry("empty conf.tcp", testCase{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    tcp: {}
`,
				expected: `
                violations:
                - field: conf.tcp.maxConnectAttempts
                  message: has to be greater than 0
`,
			}),
		)

		DescribeTable("should validate all fields and return no errors if all are valid",
			func(given testCaseWithNoErrors) {
				// setup
				retry := NewRetryResource()

				// when
				err := util_proto.FromYAML([]byte(given.retry), retry.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := retry.Validate()

				// then
				Expect(verr).To(BeNil())
			},
			Entry("all protocols configuration provided", testCaseWithNoErrors{
				retry: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                    http:
                        numRetries: 3
                        perTryTimeout: 200ms
                        backOff:
                            baseInterval: 30ms
                            maxInterval: 1.2s
                        retriableStatusCodes: [501, 502]
                    grpc:
                        numRetries: 3
                        perTryTimeout: 200ms
                        backOff:
                            baseInterval: 58ms
                            maxInterval: 1s
                        retryOn:
                        - cancelled
                        - unavailable
                    tcp:
                        maxConnectAttempts: 5
`,
			}),
		)
	})
})
