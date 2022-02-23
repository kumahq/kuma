package metrics

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rewriting the metrics URL", func() {
	type testCase struct {
		input     string
		adminPort uint32
		expected  string
	}
	DescribeTable("should",
		func(given testCase) {
			u, err := url.Parse(given.input)
			Expect(err).ToNot(HaveOccurred())

			Expect(rewriteMetricsURL(given.adminPort, u)).Should(Equal(given.expected))
		},
		Entry("use the admin port", testCase{
			input:     "http://foo/bar",
			adminPort: 99,
			expected:  "http://127.0.0.1:99/stats/prometheus",
		}),
		Entry("preserve query parameters", testCase{
			input:     "http://foo/bar?one=two&three=four",
			adminPort: 80,
			expected:  "http://127.0.0.1:80/stats/prometheus?one=two&three=four",
		}),
	)
})
