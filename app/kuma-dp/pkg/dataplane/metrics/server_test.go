package metrics

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/common/expfmt"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
)

var (
	regex         = "abc"
	includeUnused = true
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
		Entry("add usedonly and filter parameters", testCase{
			address:   "127.0.0.1",
			input:     "http://foo/bar?one=two&three=four",
			adminPort: 80,
			expected:  "http://127.0.0.1:80/stats?filter=abc&one=two&three=four&usedonly=",
			queryModifier: AddSidecarParameters(&v1alpha1.Sidecar{
				Regex:         &regex,
				IncludeUnused: &includeUnused,
			}),
		}),
		Entry("add default usedonly parameter", testCase{
			address:       "127.0.0.1",
			input:         "http://foo/bar?one=two&three=four",
			adminPort:     80,
			expected:      "http://127.0.0.1:80/stats?filter=&one=two&three=four&usedonly=",
			queryModifier: AddSidecarParameters(nil),
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
		contentTypes <- FmtOpenMetrics_0_0_1
		contentTypes <- expfmt.Format("")
		contentTypes <- expfmt.NewFormat(expfmt.TypeTextPlain)
		close(contentTypes)
		reqHeader.Add("Accept", "application/openmetrics-text;version=1.0.0,application/openmetrics-text;version=0.0.1;q=0.75,text/plain;version=0.0.4;q=0.5,*/*;q=0.1")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(FmtOpenMetrics_0_0_1))
	})

	It("should negotiate content-type based on Accept header", func() {
		contentTypes := make(chan expfmt.Format, 1)
		contentTypes <- expfmt.Format("invalid_content_type")
		close(contentTypes)
		reqHeader.Add("Accept", "application/openmetrics-text;version=1.0.0,application/openmetrics-text;version=0.0.1;q=0.75,text/plain;version=0.0.4;q=0.5,*/*;q=0.1")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(expfmt.Negotiate(reqHeader)))
	})

	It("should negotiate content-type based on Accept header", func() {
		contentTypes := make(chan expfmt.Format, 1)
		contentTypes <- expfmt.Format("invalid_content_type")
		close(contentTypes)
		reqHeader.Add("Accept", "*/*")

		actualContentType := selectContentType(contentTypes, reqHeader)
		Expect(actualContentType).To(Equal(expfmt.Negotiate(reqHeader)))
	})
})

var _ = Describe("Response Format", func() {
	type testCase struct {
		contentType    string
		expectedFormat expfmt.Format
	}
	DescribeTable("should",
		func(given testCase) {
			h := make(http.Header)
			h.Set(hdrContentType, given.contentType)
			Expect(responseFormat(h)).To(Equal(given.expectedFormat))
		},
		Entry("return FmtProtoDelim for a 'delimited protobuf content type' response", testCase{
			contentType:    "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited",
			expectedFormat: expfmt.NewFormat(expfmt.TypeProtoDelim),
		}),
		Entry("return expfmt.NewFormat(expfmt.TypeUnknown) for a 'text protobuf content type' response", testCase{
			contentType:    "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=text",
			expectedFormat: expfmt.NewFormat(expfmt.TypeUnknown),
		}),
		Entry("return FmtText for a 'text plain content type' response", testCase{
			contentType:    "text/plain; charset=UTF-8",
			expectedFormat: expfmt.NewFormat(expfmt.TypeTextPlain),
		}),
		Entry("return FmtOpenMetrics_1_0_0 for a 'openmetrics v1.0.0 content type' response", testCase{
			contentType:    "application/openmetrics-text; version=1.0.0",
			expectedFormat: FmtOpenMetrics_1_0_0,
		}),
		Entry("return FmtOpenMetrics_0_0_1 for a 'openmetrics v0.0.1 content type' response", testCase{
			contentType:    "application/openmetrics-text; version=0.0.1",
			expectedFormat: FmtOpenMetrics_0_0_1,
		}),
		Entry("return expfmt.NewFormat(expfmt.TypeUnknown) for a 'invalid content type' response", testCase{
			contentType:    "application/invalid",
			expectedFormat: expfmt.NewFormat(expfmt.TypeUnknown),
		}),
		Entry("return FmtOpenMetrics_0_0_1 for a 'openmetrics content type with no version param' response", testCase{
			contentType:    "application/openmetrics-text",
			expectedFormat: FmtOpenMetrics_0_0_1,
		}),
		Entry("return expfmt.NewFormat(expfmt.TypeUnknown) for a 'openmetrics content type with unsupported version param' response", testCase{
			contentType:    "application/openmetrics-text; version=2.0.0",
			expectedFormat: expfmt.NewFormat(expfmt.TypeUnknown),
		}),
	)
})

var _ = Describe("Process Metrics", func() {
	type testCase struct {
		input       []string // input files containing metrics
		contentType expfmt.Format
		expected    string // expected output file
	}
	DescribeTable("should",
		func(given testCase) {
			inputMetrics := make(chan []byte, len(given.input))
			for _, input := range given.input {
				fo, err := os.Open(path.Join("testdata", input))
				Expect(err).ToNot(HaveOccurred())
				byteData, err := io.ReadAll(fo)
				Expect(err).ToNot(HaveOccurred())
				inputMetrics <- byteData
			}
			close(inputMetrics)

			fo, err := os.Open(path.Join("testdata", given.expected))
			Expect(err).ToNot(HaveOccurred())
			expected, err := io.ReadAll(fo)
			Expect(err).ToNot(HaveOccurred())

			actual := processMetrics(inputMetrics, given.contentType)
			Expect(string(actual)).To(Equal(string(expected)))
		},
		Entry("return OpenMetrics compliant metrics", testCase{
			input:       []string{"openmetrics_0_1_1.in", "counter.out"},
			contentType: FmtOpenMetrics_0_0_1,
			expected:    "openmetrics_0_0_1-counter.out",
		}),
		Entry("handle multiple # EOF", testCase{
			input:       []string{"openmetrics_0_1_1.in", "openmetrics_0_1_1.in", "counter.out"},
			contentType: FmtOpenMetrics_0_0_1,
			expected:    "multi-openmetrics-counter.out",
		}),
		Entry("return Prometheus text compliant metrics", testCase{
			input:       []string{"prom-text.in", "counter.out"},
			contentType: expfmt.NewFormat(expfmt.TypeTextPlain),
			expected:    "prom-text-counter.out",
		}),
	)
})
