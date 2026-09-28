package metrics

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prom_client "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshmetric_plugin "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/plugin/v1alpha1"
)

var (
	includeUnused = true
	excludeUnused = false
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
		Entry("not add usedonly parameter when unused metrics are included", testCase{
			address:   "127.0.0.1",
			input:     "http://foo/bar?one=two&three=four",
			adminPort: 80,
			expected:  "http://127.0.0.1:80/stats?one=two&three=four",
			queryModifier: AddSidecarParameters(&v1alpha1.Sidecar{
				IncludeUnused: &includeUnused,
			}),
		}),
		Entry("drop usedonly parameter passed by the scraper when unused metrics are included", testCase{
			address:   "127.0.0.1",
			input:     "http://foo/bar?one=two&usedonly",
			adminPort: 80,
			expected:  "http://127.0.0.1:80/stats?one=two",
			queryModifier: AddSidecarParameters(&v1alpha1.Sidecar{
				IncludeUnused: &includeUnused,
			}),
		}),
		Entry("add usedonly parameter when unused metrics are excluded", testCase{
			address:   "127.0.0.1",
			input:     "http://foo/bar?one=two&three=four",
			adminPort: 80,
			expected:  "http://127.0.0.1:80/stats?one=two&three=four&usedonly=",
			queryModifier: AddSidecarParameters(&v1alpha1.Sidecar{
				IncludeUnused: &excludeUnused,
			}),
		}),
		Entry("add default usedonly parameter", testCase{
			address:       "127.0.0.1",
			input:         "http://foo/bar?one=two&three=four",
			adminPort:     80,
			expected:      "http://127.0.0.1:80/stats?one=two&three=four&usedonly=",
			queryModifier: AddSidecarParameters(nil),
		}),
	)
})

var _ = Describe("MeshMetric Prometheus endpoint", func() {
	var hits atomic.Int64
	var body string

	BeforeEach(func() {
		hits.Store(0)
		app := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			w.Header().Set(hdrContentType, "text/plain; charset=UTF-8")
			_, _ = w.Write([]byte("# TYPE test_app_requests counter\ntest_app_requests 1\n"))
		}))
		DeferCleanup(app.Close)
		host, portStr, err := net.SplitHostPort(app.Listener.Addr().String())
		Expect(err).ToNot(HaveOccurred())
		port, err := strconv.ParseUint(portStr, 10, 32)
		Expect(err).ToNot(HaveOccurred())

		producer := NewAggregatedMetricsProducer([]ApplicationToScrape{{
			Name:              "app",
			Address:           host,
			Port:              uint32(port),
			Path:              "/metrics",
			QueryModifier:     RemoveQueryParameters,
			MeshMetricMutator: AggregatedOtelMutator(),
		}}, false, "dev")

		// not GinkgoT().TempDir(), unix socket paths are capped at ~104 chars on macOS
		dir, err := os.MkdirTemp("", "hijacker")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(os.RemoveAll, dir)
		socketPath := filepath.Join(dir, "metrics.sock")

		stop := make(chan struct{})
		hijacker := New(socketPath, nil, false, producer)
		go func() {
			defer GinkgoRecover()
			Expect(hijacker.Start(stop)).To(Succeed())
		}()
		DeferCleanup(func() { close(stop) })
		Eventually(func() error {
			_, err := os.Stat(socketPath)
			return err
		}).Should(Succeed())

		client := createHTTPClientForUDS(socketPath)
		resp, err := client.Get("http://localhost" + meshmetric_plugin.PrometheusDataplaneStatsPath)
		Expect(err).ToNot(HaveOccurred())
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		body = string(b)
	})

	It("should serve application and kuma-dp metrics", func() {
		Expect(body).To(ContainSubstring("test_app_requests"))
		Expect(body).To(ContainSubstring("go_goroutines"))
		Expect(hits.Load()).To(BeEquivalentTo(1))
	})

	It("should not scrape applications when the default gatherer is gathered", func() {
		hits.Store(0)

		_, err := prom_client.DefaultGatherer.Gather()

		Expect(err).ToNot(HaveOccurred())
		Expect(hits.Load()).To(BeZero())
	})
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
