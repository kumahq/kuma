package otelreceiver

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type testTraceCollector struct {
	tracepb.UnimplementedTraceServiceServer
	mu       sync.Mutex
	requests []*tracepb.ExportTraceServiceRequest
}

func (t *testTraceCollector) Export(
	_ context.Context,
	req *tracepb.ExportTraceServiceRequest,
) (*tracepb.ExportTraceServiceResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests = append(t.requests, req)
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

var _ = Describe("Trace receiver", func() {
	It("should forward over gRPC", func() {
		collector := &testTraceCollector{}
		server := grpc.NewServer()
		tracepb.RegisterTraceServiceServer(server, collector)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer listener.Close()

		go func() { defer GinkgoRecover(); _ = server.Serve(listener) }()
		defer server.Stop()

		conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = conn.Close() }()

		receiver := newTraceReceiver(conn, nil, listener.Addr().String(), "", false)

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
	})

	DescribeTable("should forward over HTTP with correct path",
		func(basePath, expectedPath string) {
			var gotPath string
			var gotMethod string
			var gotContentType string
			var gotReq tracepb.ExportTraceServiceRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotMethod = r.Method
				gotContentType = r.Header.Get("Content-Type")

				body, err := io.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(proto.Unmarshal(body, &gotReq)).To(Succeed())
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			endpoint := strings.TrimPrefix(server.URL, "http://")
			httpClient := &http.Client{}
			receiver := newTraceReceiver(nil, httpClient, endpoint, basePath, true)

			req := &tracepb.ExportTraceServiceRequest{
				ResourceSpans: []*tracev1.ResourceSpans{
					{ScopeSpans: []*tracev1.ScopeSpans{
						{Spans: []*tracev1.Span{{Name: "trace-over-http"}}},
					}},
				},
			}
			_, err := receiver.Export(context.Background(), req)
			Expect(err).NotTo(HaveOccurred())

			Expect(gotPath).To(Equal(expectedPath))
			Expect(gotMethod).To(Equal(http.MethodPost))
			Expect(gotContentType).To(Equal("application/x-protobuf"))
			Expect(proto.Equal(req, &gotReq)).To(BeTrue())
		},
		Entry("empty base path", "", "/v1/traces"),
		Entry("slash base path", "/", "/v1/traces"),
		Entry("named base path", "/collector", "/collector/v1/traces"),
		Entry("named base path without leading slash", "collector", "/collector/v1/traces"),
		Entry("named base path with trailing slash", "/collector/", "/collector/v1/traces"),
	)

	It("should propagate HTTP error", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("collector failed"))
		}))
		defer server.Close()

		endpoint := strings.TrimPrefix(server.URL, "http://")
		httpClient := &http.Client{}
		receiver := newTraceReceiver(nil, httpClient, endpoint, "", true)

		_, err := receiver.Export(context.Background(), &tracepb.ExportTraceServiceRequest{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("502"))
		Expect(err.Error()).To(ContainSubstring("collector failed"))
	})

	It("should propagate context cancellation", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer server.Close()

		endpoint := strings.TrimPrefix(server.URL, "http://")
		httpClient := &http.Client{}
		receiver := newTraceReceiver(nil, httpClient, endpoint, "", true)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := receiver.Export(ctx, &tracepb.ExportTraceServiceRequest{})
		Expect(errors.Is(err, context.Canceled)).To(BeTrue())
	})
})

var _ = Describe("Logs receiver", func() {
	It("should forward over gRPC", func() {
		collector := &testLogsCollector{}
		server := grpc.NewServer()
		logspb.RegisterLogsServiceServer(server, collector)

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		defer listener.Close()

		go func() { defer GinkgoRecover(); _ = server.Serve(listener) }()
		defer server.Stop()

		conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = conn.Close() }()

		receiver := newLogsReceiver(conn, nil, listener.Addr().String(), "", false)

		req := &logspb.ExportLogsServiceRequest{
			ResourceLogs: []*logsv1.ResourceLogs{
				{ScopeLogs: []*logsv1.ScopeLogs{
					{LogRecords: []*logsv1.LogRecord{
						{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "log-over-grpc"}}},
					}},
				}},
			},
		}
		_, err = receiver.Export(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())
		Expect(proto.Equal(req, collector.latestRequest())).To(BeTrue())
	})

	It("should forward over HTTP", func() {
		var gotPath string
		var gotReq logspb.ExportLogsServiceRequest

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			body, err := io.ReadAll(r.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(proto.Unmarshal(body, &gotReq)).To(Succeed())
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		endpoint := strings.TrimPrefix(server.URL, "http://")
		httpClient := &http.Client{}
		receiver := newLogsReceiver(nil, httpClient, endpoint, "/otel", true)

		req := &logspb.ExportLogsServiceRequest{
			ResourceLogs: []*logsv1.ResourceLogs{
				{ScopeLogs: []*logsv1.ScopeLogs{
					{LogRecords: []*logsv1.LogRecord{
						{Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "log-over-http"}}},
					}},
				}},
			},
		}
		_, err := receiver.Export(context.Background(), req)
		Expect(err).NotTo(HaveOccurred())

		Expect(gotPath).To(Equal("/otel/v1/logs"))
		Expect(proto.Equal(req, &gotReq)).To(BeTrue())
	})

	It("should propagate HTTP error", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("collector failed"))
		}))
		defer server.Close()

		endpoint := strings.TrimPrefix(server.URL, "http://")
		httpClient := &http.Client{}
		receiver := newLogsReceiver(nil, httpClient, endpoint, "", true)

		_, err := receiver.Export(context.Background(), &logspb.ExportLogsServiceRequest{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("502"))
		Expect(err.Error()).To(ContainSubstring("collector failed"))
	})
})

var _ = Describe("otlpHTTPURL", func() {
	It("should handle IPv6 endpoint", func() {
		url := otlpHTTPURL("[2001:db8::1]:4318", "/collector", "logs")
		Expect(url).To(Equal("http://[2001:db8::1]:4318/collector/v1/logs"))
	})
})
