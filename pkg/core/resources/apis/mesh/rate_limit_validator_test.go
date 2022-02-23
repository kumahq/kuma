package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("RateLimit", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(rateLimitYAML string) {
				// setup
				rateLimit := NewRateLimitResource()

				// when
				err := util_proto.FromYAML([]byte(rateLimitYAML), rateLimit.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := rateLimit.Validate()
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
                  http:
                    requests: 10
                    interval: 10s
                    onRateLimit:
                      status: 423
                      headers:
                        - key: "x-mesh-rate-limit"
                          value: "true"
                          append: false
                        - key: "x-kuma-rate-limit"
                          value: "true"
                          append: true`),
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
                  http:
                    requests: 10
                    interval: 10s
                    onRateLimit:
                      status: 423
                      headers:
                        - key: "x-kuma-rate-limit"
                          value: "true"
                          append: true`),
		)

		type testCase struct {
			ratelimit string
			expected  string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				ratelimit := NewRateLimitResource()

				// when
				err := util_proto.FromYAML([]byte(given.ratelimit), ratelimit.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := ratelimit.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				ratelimit: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
                - field: conf
                  message: must have conf
`,
			}),
			Entry("selectors without tags", testCase{
				ratelimit: `
                sources:
                - match: {}
                destinations:
                - match: {}
                conf:
                  http:
                    requests: 10
                    interval: 10s
`,
				expected: `
                violations:
                - field: sources[0].match
                  message: must have at least one tag
                - field: destinations[0].match
                  message: must have at least one tag
`,
			}),
			Entry("selectors with empty tags values", testCase{
				ratelimit: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                destinations:
                - match:
                    kuma.io/service:
                    region:
                conf:
                  http:
                    requests: 10
                    interval: 10s
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				ratelimit: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
                destinations:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
                conf:
                  http:
                    requests: 10
                    interval: 10s
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: sources[1].match
                  message: must have at least one tag
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[1].match
                  message: must have at least one tag
`,
			}),
			Entry("http", testCase{
				ratelimit: `
                sources:
                - match:
                    kuma.io/service: '*' 
                destinations:
                - match:
                    kuma.io/service: '*'
                conf:
                  http:
                    onRateLimit:
                      headers:
                        - key: ""
                          value: ""
`,
				expected: `
                violations:
                - field: conf.http.requests
                  message: requests must be set
                - field: conf.http.interval
                  message: interval must be set
                - field: conf.http.onRateLimit.header["0"]
                  message: key must be set
                - field: conf.http.onRateLimit.header["0"]
                  message: value must be set
`,
			}),
		)
	})
})
