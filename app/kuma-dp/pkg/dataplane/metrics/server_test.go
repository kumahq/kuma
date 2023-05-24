package metrics

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"
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

var _ = Describe("Select Content Type", func() {
	var reqHeader http.Header
	BeforeEach(func() {
		reqHeader = make(http.Header)
	})

	It("should honor app content-type", func() {
		contentTypes := make(chan expfmt.Format, 3)
		contentTypes <- expfmt.FmtOpenMetrics_0_0_1
		contentTypes <- expfmt.Format("")
		contentTypes <- expfmt.FmtText
		close(contentTypes)
		reqHeader.Add("Accept", "application/openmetrics-text;version=1.0.0,application/openmetrics-text;version=0.0.1;q=0.75,text/plain;version=0.0.4;q=0.5,*/*;q=0.1")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(expfmt.FmtOpenMetrics_0_0_1))
	})

	It("should honor max supported accept type when app returns invalid content-type", func() {
		contentTypes := make(chan expfmt.Format, 1)
		contentTypes <- expfmt.Format("invalid_content_type")
		close(contentTypes)
		reqHeader.Add("Accept", "application/openmetrics-text;version=1.0.0,application/openmetrics-text;version=0.0.1;q=0.75,text/plain;version=0.0.4;q=0.5,*/*;q=0.1")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(expfmt.FmtOpenMetrics_1_0_0))
	})

	It("should use highest priority content-type available", func() {
		contentTypes := make(chan expfmt.Format, 1)
		contentTypes <- expfmt.Format("invalid_content_type")
		close(contentTypes)
		reqHeader.Add("Accept", "*/*")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(expfmt.FmtText))
	})
})
