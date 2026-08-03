package otelreceiver

import (
	"context"
	"net/url"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

type traceExportFunc func(
	ctx context.Context,
	req *tracepb.ExportTraceServiceRequest,
) (*tracepb.ExportTraceServiceResponse, error)

// traceReceiver proxies ExportTraceServiceRequest from Envoy directly to the collector.
type traceReceiver struct {
	tracepb.UnimplementedTraceServiceServer
	exportFn traceExportFunc
}

func newTraceReceiver(exportFn traceExportFunc) *traceReceiver {
	return &traceReceiver{exportFn: exportFn}
}

func (r *traceReceiver) Export(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	return r.exportFn(ctx, req)
}

type logsExportFunc func(
	ctx context.Context,
	req *logspb.ExportLogsServiceRequest,
) (*logspb.ExportLogsServiceResponse, error)

// logsReceiver proxies ExportLogsServiceRequest from Envoy directly to the collector.
type logsReceiver struct {
	logspb.UnimplementedLogsServiceServer
	exportFn logsExportFunc
}

func newLogsReceiver(exportFn logsExportFunc) *logsReceiver {
	return &logsReceiver{exportFn: exportFn}
}

func (r *logsReceiver) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	return r.exportFn(ctx, req)
}

type metricsExportFunc func(
	ctx context.Context,
	req *metricspb.ExportMetricsServiceRequest,
) (*metricspb.ExportMetricsServiceResponse, error)

// metricsReceiver proxies ExportMetricsServiceRequest from kuma-dp SDK exporter to the collector.
type metricsReceiver struct {
	metricspb.UnimplementedMetricsServiceServer
	exportFn metricsExportFunc
}

func newMetricsReceiver(exportFn metricsExportFunc) *metricsReceiver {
	return &metricsReceiver{exportFn: exportFn}
}

func (r *metricsReceiver) Export(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
	return r.exportFn(ctx, req)
}

func otlpHTTPURL(endpoint, targetPath string, useHTTPS bool) string {
	scheme := "http"
	if useHTTPS {
		scheme = "https"
	}
	targetURL := url.URL{
		Scheme: scheme,
		Host:   endpoint,
		Path:   targetPath,
	}
	return targetURL.String()
}
