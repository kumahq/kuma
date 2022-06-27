package metrics

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rewriting the metrics URL", func() {
	type testCase struct {
		input         string
		adminPort     uint32
		expected      string
		queryModifier QueryParametersModifier
	}
	DescribeTable("should",
		func(given testCase) {
			u, err := url.Parse(given.input)
			Expect(err).ToNot(HaveOccurred())
			Expect(rewriteMetricsURL("/stats", given.adminPort, given.queryModifier, u)).Should(Equal(given.expected))
		},
		Entry("use the admin port", testCase{
			input:         "http://foo/bar",
			adminPort:     99,
			expected:      "http://127.0.0.1:99/stats?format=prometheus",
			queryModifier: AddPrometheusFormat,
		}),
		Entry("preserve query parameters", testCase{
			input:         "http://foo/bar?one=two&three=four&filter=test_.*&usedonly",
			adminPort:     80,
			expected:      "http://127.0.0.1:80/stats?filter=test_.%2A&format=prometheus&one=two&three=four&usedonly=",
			queryModifier: AddPrometheusFormat,
		}),
		Entry("remove query parameters", testCase{
			input:         "http://foo/bar?one=two&three=four",
			adminPort:     80,
			expected:      "http://127.0.0.1:80/stats",
			queryModifier: RemoveQueryParameters,
		}),
	)
})
