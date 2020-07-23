package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
				healthCheck := HealthCheckResource{}

				// when
				err := util_proto.FromYAML([]byte(given.healthCheck), &healthCheck.Spec)
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
                  message: mandatory tag "service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "service"
                - field: destinations[0].match
                  message: mandatory tag "service" is missing
                - field: conf
                  message: has to be defined
`,
			}),
			Entry("selectors with empty tags values", testCase{
				healthCheck: `
                sources:
                - match:
                    service:
                    region:
                destinations:
                - match:
                    service:
                    region:
`,
				expected: `
                violations:
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: sources[0].match["service"]
                  message: tag value must be non-empty
                - field: destinations[0].match
                  message: must consist of exactly one tag "service"
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[0].match["service"]
                  message: tag value must be non-empty
                - field: conf
                  message: has to be defined`,
			}),
			Entry("invalid active checks conf", testCase{
				healthCheck: `
                sources:
                - match:
                    service: web
                    region: eu
                destinations:
                - match:
                    service: backend
                conf:
                  interval: 0s
                  timeout: 0s
                  unhealthyThreshold: 0
                  healthyThreshold: 0
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
`,
			}),
		)
	})
})
