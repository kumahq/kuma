package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("TrafficRoute", func() {
	Describe("Validate()", func() {
		type testCase struct {
			route    string
			expected string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				route := TrafficRouteResource{}

				// when
				err := util_proto.FromYAML([]byte(given.route), &route.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := route.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				route: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
                - field: conf
                  message: must have at least one element
`,
			}),
			Entry("selectors without tags", testCase{
				route: `
                sources:
                - match: {}
                destinations:
                - match: {}
                conf:
                - destination: {}
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
                - field: conf[0].destination
                  message: must have at least one tag
                - field: conf[0].destination
                  message: mandatory tag "service" is missing
`,
			}),
			Entry("selectors with empty tags values", testCase{
				route: `
                sources:
                - match:
                    service:
                    region:
                destinations:
                - match:
                    service:
                    region:
                conf:
                - destination:
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
                - field: conf[0].destination["region"]
                  message: tag value must be non-empty
                - field: conf[0].destination["service"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				route: `
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
                conf:
                - destination:
                    service:
                    region:
                - destination: {}
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
                - field: conf[0].destination["region"]
                  message: tag value must be non-empty
                - field: conf[0].destination["service"]
                  message: tag value must be non-empty
                - field: conf[1].destination
                  message: must have at least one tag
                - field: conf[1].destination
                  message: mandatory tag "service" is missing
`,
			}),
		)
	})
})
