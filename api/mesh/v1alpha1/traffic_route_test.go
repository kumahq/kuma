package v1alpha1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/api/mesh/v1alpha1"

	util_proto "github.com/Kong/kuma/api/internal/util/proto"
)

var _ = Describe("TrafficRoute", func() {

	Context("valid configurations", func() {
		type testCase struct {
			input string
		}

		DescribeTable("Validate() should return a nil",
			func(given testCase) {
				// setup
				route := &TrafficRoute{}

				// when
				err := util_proto.FromYAML([]byte(given.input), route)
				// then
				Expect(err).ToNot(HaveOccurred())

				// expect
				Expect(route.Validate()).To(Succeed())
			},
			Entry("conf with 1 destination but no weight", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                - destination:
                    service: backend
`,
			}),
			Entry("conf with 2 destinations and weights", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                - destination:
                    service: backend
                    version: v1
                - weight: 99
                  destination:
                    service: backend
                    version: v2
`,
			}),
		)
	})

	Context("invalid configurations", func() {
		type testCase struct {
			input       string
			expectedErr interface{}
		}

		DescribeTable("Validate() should return an error",
			func(given testCase) {
				// setup
				route := &TrafficRoute{}

				// when
				err := util_proto.FromYAML([]byte(given.input), route)
				// then
				Expect(err).ToNot(HaveOccurred())

				// when
				err = route.Validate()
				// then
				Expect(err).To(MatchError(given.expectedErr))
			},
			Entry("0 sources", testCase{
				input:       ``,
				expectedErr: `invalid TrafficRoute.Sources: value must contain at least 1 item(s)`,
			}),
			Entry("0 destinations", testCase{
				input: `
                sources:
                - match:
                    service: web
`,
				expectedErr: `invalid TrafficRoute.Destinations: value must contain at least 1 item(s)`,
			}),
			Entry("no conf", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
`,
				expectedErr: `invalid TrafficRoute.Conf: value must contain at least 1 item(s)`,
			}),
			Entry("conf with 1 weighted destination but not selector", testCase{
				input: `
                sources:
                - match:
                    service: web
                destinations:
                - match:
                    service: backend
                conf:
                - {}
`,
				expectedErr: `invalid TrafficRoute.Conf[0]: embedded message failed validation | caused by: invalid TrafficRoute_WeightedDestination.Destination: value must contain at least 1 pair(s)`,
			}),
		)
	})
})
