package metrics

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rewriting the metrics URL", func() {
	type testCase struct {
		input         string
		address       string
		adminPort     uint32
		expected      string
		queryModifier QueryParametersModifier
	}
	DescribeTable("should",
		func(given testCase) {
			u, err := url.Parse(given.input)
			Expect(err).ToNot(HaveOccurred())
			Expect(rewriteMetricsURL(given.address, given.adminPort, "/stats", given.queryModifier, u)).Should(Equal(given.expected))
		},
		Entry("use the admin port", testCase{
			address:       "1.2.3.4",
			input:         "http://foo/bar",
			adminPort:     99,
			expected:      "http://1.2.3.4:99/stats?format=prometheus&text_readouts=",
			queryModifier: AddPrometheusFormat,
		}),
		Entry("preserve query parameters", testCase{
			address:       "1.2.3.4",
			input:         "http://foo/bar?one=two&three=four&filter=test_.*&usedonly",
			adminPort:     80,
			expected:      "http://1.2.3.4:80/stats?filter=test_.%2A&format=prometheus&one=two&text_readouts=&three=four&usedonly=",
			queryModifier: AddPrometheusFormat,
		}),
		Entry("remove query parameters", testCase{
			address:       "127.0.0.1",
			input:         "http://foo/bar?one=two&three=four",
			adminPort:     80,
			expected:      "http://127.0.0.1:80/stats",
			queryModifier: RemoveQueryParameters,
		}),
	)
})
