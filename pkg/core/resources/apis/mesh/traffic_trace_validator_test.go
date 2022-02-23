package mesh_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("TrafficTrace", func() {
	Describe("Validate()", func() {
		DescribeTable("should pass validation",
			func(trafficTraceYAML string) {
				// setup
				trafficTrace := NewTrafficTraceResource()

				// when
				err := util_proto.FromYAML([]byte(trafficTraceYAML), trafficTrace.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := trafficTrace.Validate()

				// then
				Expect(verr).ToNot(HaveOccurred())
			},
			Entry("full example", `
                selectors:
                - match:
                    region: eu
                conf:
                  backend: zipkin-eu`,
			),
			Entry("empty backend", `
                selectors:
                - match:
                    region: eu
                conf:
                  backend: # backend can be empty, default backend from mesh is chosen`,
			),
			Entry("empty conf", `
                selectors:
                - match:
                    region: eu`,
			),
		)

		type testCase struct {
			trafficTrace string
			expected     string
		}
		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// setup
				trafficTrace := NewTrafficTraceResource()

				// when
				err := util_proto.FromYAML([]byte(given.trafficTrace), trafficTrace.Spec)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				verr := trafficTrace.Validate()
				// and
				actual, err := yaml.Marshal(verr)

				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("empty spec", testCase{
				trafficTrace: ``,
				expected: `
                violations:
                - field: selectors
                  message: must have at least one element
`,
			}),
			Entry("selectors without tags", testCase{
				trafficTrace: `
                selectors:
                - match: {}
`,
				expected: `
                violations:
                - field: selectors[0].match
                  message: must have at least one tag
`,
			}),
			Entry("selectors with empty tags values", testCase{
				trafficTrace: `
                selectors:
                - match:
                    kuma.io/service:
                    region:
`,
				expected: `
                violations:
                - field: selectors[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: selectors[0].match["region"]
                  message: tag value must be non-empty
`,
			}),
			Entry("multiple selectors", testCase{
				trafficTrace: `
                selectors:
                - match:
                    kuma.io/service:
                    region:
                - match: {}
`,
				expected: `
                violations:
                - field: selectors[0].match["kuma.io/service"]
                  message: tag value must be non-empty
                - field: selectors[0].match["region"]
                  message: tag value must be non-empty
                - field: selectors[1].match
                  message: must have at least one tag
`,
			}),
		)
	})
})
