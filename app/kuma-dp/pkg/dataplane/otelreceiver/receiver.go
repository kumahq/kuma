package otelreceiver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// traceReceiver proxies ExportTraceServiceRequest from Envoy directly to the collector.
type traceReceiver struct {
	tracepb.UnimplementedTraceServiceServer
	exportFn func(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error)
}

func newTraceReceiver(conn *grpc.ClientConn, httpClient *http.Client, endpoint, basePath string, useHTTP bool) *traceReceiver {
	if useHTTP {
		targetURL := otlpHTTPURL(endpoint, basePath, "traces")
		return &traceReceiver{
			exportFn: func(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
				body, err := proto.Marshal(req)
				if err != nil {
					return nil, err
				}
				if err := postOTLPHTTP(ctx, httpClient, targetURL, body); err != nil {
					return nil, err
				}
				return &tracepb.ExportTraceServiceResponse{}, nil
			},
		}
	}
	client := tracepb.NewTraceServiceClient(conn)
	return &traceReceiver{
		exportFn: func(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
			return client.Export(ctx, req)
		},
	}
}

func (r *traceReceiver) Export(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	return r.exportFn(ctx, req)
}

// logsReceiver proxies ExportLogsServiceRequest from Envoy directly to the collector.
type logsReceiver struct {
	logspb.UnimplementedLogsServiceServer
	exportFn func(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error)
}

func newLogsReceiver(conn *grpc.ClientConn, httpClient *http.Client, endpoint, basePath string, useHTTP bool) *logsReceiver {
	if useHTTP {
		targetURL := otlpHTTPURL(endpoint, basePath, "logs")
		return &logsReceiver{
			exportFn: func(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
				body, err := proto.Marshal(req)
				if err != nil {
					return nil, err
				}
				if err := postOTLPHTTP(ctx, httpClient, targetURL, body); err != nil {
					return nil, err
				}
				return &logspb.ExportLogsServiceResponse{}, nil
			},
		}
	}
	client := logspb.NewLogsServiceClient(conn)
	return &logsReceiver{
		exportFn: func(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
			return client.Export(ctx, req)
		},
	}
}

func (r *logsReceiver) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	return r.exportFn(ctx, req)
}

// metricsReceiver proxies ExportMetricsServiceRequest from kuma-dp SDK exporter to the collector.
type metricsReceiver struct {
	metricspb.UnimplementedMetricsServiceServer
	exportFn func(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error)
}

func newMetricsReceiver(conn *grpc.ClientConn, httpClient *http.Client, endpoint, basePath string, useHTTP bool) *metricsReceiver {
	if useHTTP {
		targetURL := otlpHTTPURL(endpoint, basePath, "metrics")
		return &metricsReceiver{
			exportFn: func(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
				body, err := proto.Marshal(req)
				if err != nil {
					return nil, err
				}
				if err := postOTLPHTTP(ctx, httpClient, targetURL, body); err != nil {
					return nil, err
				}
				return &metricspb.ExportMetricsServiceResponse{}, nil
			},
		}
	}
	client := metricspb.NewMetricsServiceClient(conn)
	return &metricsReceiver{
		exportFn: func(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
			return client.Export(ctx, req)
		},
	}
}

func (r *metricsReceiver) Export(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
	return r.exportFn(ctx, req)
}

func otlpHTTPURL(endpoint, basePath, signal string) string {
	targetURL := url.URL{
		Scheme: "http",
		Host:   endpoint,
		Path:   path.Join("/", basePath, "v1", signal),
	}
	return targetURL.String()
}

func postOTLPHTTP(
	ctx context.Context,
	client *http.Client,
	targetURL string,
	body []byte,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")

	resp, err := client.Do(req) //nolint:gosec // URL is built from internal config, not user input
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if len(responseBody) == 0 {
		return fmt.Errorf("collector returned status %d", resp.StatusCode)
	}
	return fmt.Errorf("collector returned status %d: %s", resp.StatusCode, string(responseBody))
}
