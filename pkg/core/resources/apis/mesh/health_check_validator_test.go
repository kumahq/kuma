package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("HealthCheck", func() {
	Describe("Validate()", func() {
		type testCase struct {
			healthCheck string
			expected    string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				healthCheck := NewHealthCheckResource()

				// when
				err := util_proto.FromYAML([]byte(given.healthCheck), healthCheck.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := healthCheck.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				healthCheck: ``,
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
				healthCheck: `
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
				healthCheck: `
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
			Entry("invalid active checks conf", testCase{
				healthCheck: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  interval: 0s
                  timeout: 0s
                  unhealthyThreshold: 0
                  healthyThreshold: 0
                  initialJitter: 0s
                  intervalJitter: 0s
                  healthyPanicThreshold: 101
                  noTrafficInterval: 0s
                  http:
                    path: ""
                    requestHeadersToAdd:
                    - header:
                        value: foo
                    - append: false
                    expectedStatuses:
                    - 99
                    - 600
`,
				expected: `
                violations:
                - field: conf.interval
                  message: must have a positive value
                - field: conf.timeout
                  message: must have a positive value
                - field: conf.unhealthyThreshold
                  message: must have a positive value
                - field: conf.healthyThreshold
                  message: must have a positive value
                - field: conf.initialJitter
                  message: must have a positive value
                - field: conf.intervalJitter
                  message: must have a positive value
                - field: conf.noTrafficInterval
                  message: must have a positive value
                - field: conf.healthyPanicThreshold
                  message: must be in range [0.0 - 100.0]
                - field: conf.http.path
                  message: has to be defined and cannot be empty
                - field: conf.http.expectedStatuses[0]
                  message: must be in range [100, 600)
                - field: conf.http.expectedStatuses[1]
                  message: must be in range [100, 600)
                - field: conf.http.requestHeadersToAdd[0].header.key
                  message: cannot be empty
                - field: conf.http.requestHeadersToAdd[1].header
                  message: has to be defined
`,
			}),
			Entry("invalid active checks http configuration", testCase{
				healthCheck: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  interval: 3s
                  timeout: 10s
                  unhealthyThreshold: 3
                  healthyThreshold: 1
                  http: {}
`,
				expected: `
                violations:
                - field: conf.http.path
                  message: has to be defined and cannot be empty
`,
			}),
			Entry("http and tcp configuration", testCase{
				healthCheck: `
                sources:
                - match:
                    kuma.io/service: web
                    region: eu
                destinations:
                - match:
                    kuma.io/service: backend
                conf:
                  interval: 3s
                  timeout: 10s
                  unhealthyThreshold: 3
                  healthyThreshold: 1
                  http: {}
                  tcp: {}
`,
				expected: `
                violations:
                - field: conf.http.path
                  message: has to be defined and cannot be empty
                - field: conf
                  message: http and tcp cannot be defined at the same time
`,
			}),
		)
	})
})
