package otelreceiver

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/otelenv"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

type testTraceCollector struct {
	tracepb.UnimplementedTraceServiceServer
	mu       sync.Mutex
	requests []*tracepb.ExportTraceServiceRequest
	headers  metadata.MD
}

func (t *testTraceCollector) Export(
	ctx context.Context,
	req *tracepb.ExportTraceServiceRequest,
) (*tracepb.ExportTraceServiceResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests = append(t.requests, req)
	t.headers, _ = metadata.FromIncomingContext(ctx)
	return &tracepb.ExportTraceServiceResponse{}, nil
}

func (t *testTraceCollector) latestRequest() *tracepb.ExportTraceServiceRequest {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.requests) == 0 {
		return nil
	}
	return t.requests[len(t.requests)-1]
}

func (t *testTraceCollector) latestHeaders() metadata.MD {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.headers
}

type testLogsCollector struct {
	logspb.UnimplementedLogsServiceServer
	mu       sync.Mutex
	requests []*logspb.ExportLogsServiceRequest
}

func (t *testLogsCollector) Export(
	_ context.Context,
	req *logspb.ExportLogsServiceRequest,
) (*logspb.ExportLogsServiceResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests = append(t.requests, req)
	return &logspb.ExportLogsServiceResponse{}, nil
}

func (t *testLogsCollector) latestRequest() *logspb.ExportLogsServiceRequest {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.requests) == 0 {
		return nil
	}
	return t.requests[len(t.requests)-1]
}

var _ = Describe("senderFactory", func() {
	It("should forward traces over gRPC with outgoing headers", func() {
		collector := &testTraceCollector{}
		server := grpc.NewServer()
		tracepb.RegisterTraceServiceServer(server, collector)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer listener.Close()

		go func() {
			defer GinkgoRecover()
			_ = server.Serve(listener)
		}()
		defer server.Stop()

		factory := &senderFactory{}
		DeferCleanup(factory.close)

		exporter, err := factory.newTraceExporter(otelenv.SignalRuntime{
			Enabled: true,
			Transport: otelenv.ExporterTransport{
				Protocol: core_xds.OtelProtocolGRPC,
				Endpoint: listener.Addr().String(),
				Headers:  map[string]string{"authorization": "token"},
				Timeout:  time.Second,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		receiver := newTraceReceiver(exporter)
		req := &tracepb.ExportTraceServiceRequest{
			ResourceSpans: []*tracev1.ResourceSpans{
				{ScopeSpans: []*tracev1.ScopeSpans{
					{Spans: []*tracev1.Span{{Name: "trace-over-grpc"}}},
				}},
			},
		}

		_, err = receiver.Export(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		Expect(proto.Equal(req, collector.latestRequest())).To(BeTrue())
		Expect(collector.latestHeaders().Get("authorization")).To(ConsistOf("token"))
	})

	It("should forward logs over HTTP with the resolved path and gzip body", func() {
		var gotPath string
		var gotAuthorization string
		var gotContentType string
		var gotContentEncoding string
		var gotReq logspb.ExportLogsServiceRequest

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotAuthorization = r.Header.Get("Authorization")
			gotContentType = r.Header.Get("Content-Type")
			gotContentEncoding = r.Header.Get("Content-Encoding")

			body, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			if gotContentEncoding == "gzip" {
				reader, err := gzip.NewReader(bytes.NewReader(body))
				Expect(err).NotTo(HaveOccurred())
				defer reader.Close()
				body, err = io.ReadAll(reader)
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(proto.Unmarshal(body, &gotReq)).To(Succeed())
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		factory := &senderFactory{}
		DeferCleanup(factory.close)

		exporter, err := factory.newLogsExporter(otelenv.SignalRuntime{
			Enabled:  true,
			HTTPPath: "/collector/v1/logs",
			Transport: otelenv.ExporterTransport{
				Protocol:    core_xds.OtelProtocolHTTPProtobuf,
				Endpoint:    strings.TrimPrefix(server.URL, "http://"),
				Headers:     map[string]string{"Authorization": "Bearer token"},
				Compression: "gzip",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		receiver := newLogsReceiver(exporter)
		req := &logspb.ExportLogsServiceRequest{
			ResourceLogs: []*logsv1.ResourceLogs{
				{ScopeLogs: []*logsv1.ScopeLogs{
					{LogRecords: []*logsv1.LogRecord{
						{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "log-over-http"}}},
					}},
				}},
			},
		}

		_, err = receiver.Export(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		Expect(gotPath).To(Equal("/collector/v1/logs"))
		Expect(gotAuthorization).To(Equal("Bearer token"))
		Expect(gotContentType).To(Equal("application/x-protobuf"))
		Expect(gotContentEncoding).To(Equal("gzip"))
		Expect(proto.Equal(req, &gotReq)).To(BeTrue())
	})

	It("should return a failed precondition when metrics runtime is blocked", func() {
		factory := &senderFactory{}
		exporter, err := factory.newMetricsExporter(otelenv.SignalRuntime{
			Enabled:        true,
			BlockedReasons: []string{core_xds.OtelBlockedReasonRequiredEnvMissing},
		})
		Expect(err).NotTo(HaveOccurred())

		receiver := newMetricsReceiver(exporter)
		_, err = receiver.Export(context.Background(), &metricspb.ExportMetricsServiceRequest{
			ResourceMetrics: []*metricsv1.ResourceMetrics{},
		})
		Expect(status.Code(err)).To(Equal(codes.FailedPrecondition))
		Expect(err.Error()).To(ContainSubstring(core_xds.OtelBlockedReasonRequiredEnvMissing))
	})

	It("should propagate HTTP collector errors", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("collector failed"))
		}))
		defer server.Close()

		factory := &senderFactory{}
		DeferCleanup(factory.close)

		exporter, err := factory.newTraceExporter(otelenv.SignalRuntime{
			Enabled:  true,
			HTTPPath: "/v1/traces",
			Transport: otelenv.ExporterTransport{
				Protocol: core_xds.OtelProtocolHTTPProtobuf,
				Endpoint: strings.TrimPrefix(server.URL, "http://"),
			},
		})
		Expect(err).NotTo(HaveOccurred())

		receiver := newTraceReceiver(exporter)
		_, err = receiver.Export(context.Background(), &tracepb.ExportTraceServiceRequest{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("502"))
		Expect(err.Error()).To(ContainSubstring("collector failed"))
	})
})

var _ = Describe("otlpHTTPURL", func() {
	It("should handle IPv6 endpoints", func() {
		url := otlpHTTPURL("[2001:db8::1]:4318", "/collector/v1/logs", false)
		Expect(url).To(Equal("http://[2001:db8::1]:4318/collector/v1/logs"))
	})

	It("should generate HTTPS URLs when requested", func() {
		url := otlpHTTPURL("collector.example:443", "/otlp/v1/metrics", true)
		Expect(url).To(Equal("https://collector.example:443/otlp/v1/metrics"))
	})
})
