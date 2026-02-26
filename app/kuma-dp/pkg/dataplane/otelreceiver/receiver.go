package otelreceiver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

// traceReceiver proxies ExportTraceServiceRequest from Envoy directly to the collector.
type traceReceiver struct {
	tracepb.UnimplementedTraceServiceServer
	client tracepb.TraceServiceClient
}

func newTraceReceiver(endpoint string) (*traceReceiver, func(), error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &traceReceiver{client: tracepb.NewTraceServiceClient(conn)}, func() { _ = conn.Close() }, nil
}

func (r *traceReceiver) Export(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	return r.client.Export(ctx, req)
}

// logsReceiver proxies ExportLogsServiceRequest from Envoy directly to the collector.
type logsReceiver struct {
	logspb.UnimplementedLogsServiceServer
	client logspb.LogsServiceClient
}

func newLogsReceiver(endpoint string) (*logsReceiver, func(), error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return &logsReceiver{client: logspb.NewLogsServiceClient(conn)}, func() { _ = conn.Close() }, nil
}

func (r *logsReceiver) Export(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
	return r.client.Export(ctx, req)
}
