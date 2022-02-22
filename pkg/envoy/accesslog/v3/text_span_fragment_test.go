package v3_test

import (
	accesslog_data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/envoy/accesslog/v3"
)

var _ = Describe("TextSpan", func() {

	Describe("FormatHttpLogEntry() and FormatTcpLogEntry()", func() {
		type testCase struct {
			text     string
			expected string
		}

		DescribeTable("should format properly",
			func(given testCase) {
				// setup
				fragment := TextSpan(given.text)

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
			Entry("", testCase{
				text:     "",
				expected: ``,
			}),
			Entry("plain text", testCase{
				text:     "plain text",
				expected: `plain text`,
			}),
		)
	})

	Describe("String()", func() {
		type testCase struct {
			text     string
			expected string
		}

		DescribeTable("should return correct canonical representation",
			func(given testCase) {
				// setup
				fragment := TextSpan(given.text)

				// when
				actual := fragment.String()
				// then
				Expect(actual).To(Equal(given.expected))

			},
			Entry("", testCase{
				text:     "",
				expected: ``,
			}),
			Entry("plain text", testCase{
				text:     "plain text",
				expected: `plain text`,
			}),
		)
	})
})
