package v3_test

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("Placeholder", func() {

	Describe("FormatHttpLogEntry() and FormatTcpLogEntry()", func() {
		type testCase struct {
			variable string
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// setup
				fragment := Placeholder(given.variable)

				// when
				actual, err := fragment.FormatHttpLogEntry(&accesslog_data.HTTPAccessLogEntry{})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))

				// when
				actual, err = fragment.FormatTcpLogEntry(&accesslog_data.TCPAccessLogEntry{})
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(given.expected))
			},
			Entry("KUMA_SOURCE_ADDRESS", testCase{
				variable: "KUMA_SOURCE_ADDRESS",
				expected: `%KUMA_SOURCE_ADDRESS%`, // placeholder must be rendered "as is"
			}),
			Entry("KUMA_SOURCE_ADDRESS_WITHOUT_PORT", testCase{
				variable: "KUMA_SOURCE_ADDRESS_WITHOUT_PORT",
				expected: `%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%`, // placeholder must be rendered "as is"
			}),
			Entry("KUMA_SOURCE_SERVICE", testCase{
				variable: "KUMA_SOURCE_SERVICE",
				expected: `%KUMA_SOURCE_SERVICE%`, // placeholder must be rendered "as is"
			}),
			Entry("KUMA_DESTINATION_SERVICE", testCase{
				variable: "KUMA_DESTINATION_SERVICE",
				expected: `%KUMA_DESTINATION_SERVICE%`, // placeholder must be rendered "as is"
			}),
		)
	})

	Describe("Interpolate()", func() {
		type testCase struct {
			variable string
			context  map[string]string
			expected string
		}

		DescribeTable("should replace placeholder with a text literal",
			func(given testCase) {
				// setup
				fragment := Placeholder(given.variable)

				// when
				actual, err := fragment.Interpolate(InterpolationVariables(given.context))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(actual).To(Equal(TextSpan(given.expected)))
			},
			Entry("`nil` context", testCase{
				variable: "KUMA_SOURCE_SERVICE",
				context:  nil,
				expected: ``,
			}),
			Entry("variable w/o a value in the context", testCase{
				variable: "KUMA_SOURCE_SERVICE",
				context: map[string]string{
					"KUMA_DESTINATION_SERVICE": "backend",
				},
				expected: ``,
			}),
			Entry("variable w/ a value in the context", testCase{
				variable: "KUMA_SOURCE_SERVICE",
				context: map[string]string{
					"KUMA_SOURCE_SERVICE": "web",
				},
				expected: `web`,
			}),
		)
	})

	Describe("String()", func() {
		type testCase struct {
			variable string
			expected string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := Placeholder(given.variable)

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("KUMA_SOURCE_ADDRESS", testCase{
				variable: "KUMA_SOURCE_ADDRESS",
				expected: `%KUMA_SOURCE_ADDRESS%`,
			}),
			Entry("KUMA_SOURCE_ADDRESS_WITHOUT_PORT", testCase{
				variable: "KUMA_SOURCE_ADDRESS_WITHOUT_PORT",
				expected: `%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%`,
			}),
			Entry("KUMA_SOURCE_SERVICE", testCase{
				variable: "KUMA_SOURCE_SERVICE",
				expected: `%KUMA_SOURCE_SERVICE%`,
			}),
			Entry("KUMA_DESTINATION_SERVICE", testCase{
				variable: "KUMA_DESTINATION_SERVICE",
				expected: `%KUMA_DESTINATION_SERVICE%`,
			}),
		)
	})
})
