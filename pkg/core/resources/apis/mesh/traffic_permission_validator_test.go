package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("TrafficPermission", func() {
	Describe("Validate()", func() {
		type testCase struct {
			permission string
			expected   string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				permission := NewTrafficPermissionResource()

				// when
				err := util_proto.FromYAML([]byte(given.permission), permission.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := permission.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				permission: ``,
				expected: `
                violations:
                - field: sources
                  message: must have at least one element
                - field: destinations
                  message: must have at least one element
`,
			}),
			Entry("selectors without tags", testCase{
				permission: `
                sources:
                - match: {}
                destinations:
                - match: {}
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
				permission: `
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
                - field: destinations[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: destinations[0].match["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				permission: `
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
		)
	})
})
