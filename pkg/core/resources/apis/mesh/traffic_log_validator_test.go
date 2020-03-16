package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("TrafficLog", func() {
	Describe("Validate()", func() {
		type testCase struct {
			trafficLog string
			expected   string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				trafficLog := TrafficLogResource{}

				// when
				err := util_proto.FromYAML([]byte(given.trafficLog), &trafficLog.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := trafficLog.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				trafficLog: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
`,
			}),
			Entry("selectors without tags", testCase{
				trafficLog: `
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
`,
			}),
			Entry("selectors with empty tags values", testCase{
				trafficLog: `
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
`,
			}),
			Entry("multiple selectors", testCase{
				trafficLog: `
                sources:
                - match:
                    service:
                    region:
                - match: {}
                destinations:
                - match:
                    service:
                    region:
                - match: {}
`,
				expected: `
                violations:
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: sources[0].match["service"]
                  message: tag value must be non-empty
                - field: sources[1].match
                  message: must have at least one tag
                - field: sources[1].match
                  message: mandatory tag "service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "service"
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[0].match["service"]
                  message: tag value must be non-empty
                - field: destinations[1].match
                  message: must consist of exactly one tag "service"
                - field: destinations[1].match
                  message: mandatory tag "service" is missing
`,
			}),
		)
	})
})
