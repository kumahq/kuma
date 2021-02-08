package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
				route := NewTrafficRouteResource()

				// when
				err := util_proto.FromYAML([]byte(given.route), route.Spec)
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
                  message: must have split
                - field: conf.split
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
                  split:
                  - destination: {}
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
                - field: conf.split[0].destination
                  message: must have at least one tag
                - field: conf.split[0].destination
                  message: mandatory tag "kuma.io/service" is missing
`,
			}),
			Entry("selectors with empty tags values", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service:
                    region:
                destinations:
                - match:
                    kuma.io/service:
                    region:
                conf:
                  split:
                  - destination:
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
                - field: conf.split[0].destination["kuma.io/service"]
                  message: tag value must be non-empty
                - field: conf.split[0].destination["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				route: `
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
                  split:
                  - destination:
                      kuma.io/service:
                      region:
                  - destination: {}
`,
				expected: `
                violations:
                - field: sources[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: sources[0].match["region"]
                  message: tag value must be non-empty
                - field: sources[1].match
                  message: must have at least one tag
                - field: sources[1].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: destinations[0].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag "region" is not allowed
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
                - field: destinations[1].match
                  message: must consist of exactly one tag "kuma.io/service"
                - field: destinations[1].match
                  message: mandatory tag "kuma.io/service" is missing
                - field: conf.split[0].destination["kuma.io/service"]
                  message: tag value must be non-empty
                - field: conf.split[0].destination["region"]
                  message: tag value must be non-empty
                - field: conf.split[1].destination
                  message: must have at least one tag
                - field: conf.split[1].destination
                  message: mandatory tag "kuma.io/service" is missing
`,
			}),
			Entry("wrong ring hash function in the load balancer", testCase{
				route: `
                sources:
                - match:
                    kuma.io/service: '*'
                destinations:
                - match:
                    kuma.io/service: '*'
                conf:
                  split:
                  - destination:
                      kuma.io/service: 'backend'
                  loadBalancer:
                    ringHash:
                      hashFunction: 'INVALID_HASH_FUNCTION'
`,
				expected: `
                violations:
                - field: conf.loadBalancer.ringHash.hashFunction
                  message: must have a valid hash function
`,
			}),
		)
	})
})
