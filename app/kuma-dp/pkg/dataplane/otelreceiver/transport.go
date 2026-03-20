package otelreceiver

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/pkg/errors"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpcgzip "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/otelenv"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

type grpcConnEntry struct {
	key  grpcConnKey
	conn *grpc.ClientConn
}

type httpClientEntry struct {
	key    httpClientKey
	client *http.Client
}

type grpcConnKey struct {
	endpoint          string
	useTLS            bool
	certificate       string
	clientCertificate string
	clientKey         string
}

type httpClientKey struct {
	endpoint          string
	useTLS            bool
	certificate       string
	clientCertificate string
	clientKey         string
}

type senderFactory struct {
	grpcConns   []grpcConnEntry
	httpClients []httpClientEntry
	closeFns    []func()
}

func (f *senderFactory) close() {
	for _, closeFn := range f.closeFns {
		closeFn()
	}
}

func (f *senderFactory) newTraceExporter(runtime otelenv.SignalRuntime) (traceExportFunc, error) {
	if err := validateRuntime(runtime); err != nil {
		return unavailableTraceExporter(runtime), nil
	}

	switch runtime.Transport.Protocol {
	case core_xds.OtelProtocolHTTPProtobuf:
		client, err := f.httpClient(runtime.Transport)
		if err != nil {
			return nil, err
		}
		targetURL := otlpHTTPURL(runtime.Transport.Endpoint, runtime.HTTPPath, pointer.Deref(runtime.Transport.UseTLS))
		return func(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
			body, err := proto.Marshal(req)
			if err != nil {
				return nil, err
			}
			if err := postOTLPHTTP(ctx, client, targetURL, body, runtime.Transport.Headers, runtime.Transport.Compression, runtime.Transport.Timeout); err != nil {
				return nil, err
			}
			return &tracepb.ExportTraceServiceResponse{}, nil
		}, nil
	default:
		conn, err := f.grpcConn(runtime.Transport)
		if err != nil {
			return nil, err
		}
		client := tracepb.NewTraceServiceClient(conn)
		return func(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
			ctx, cancel := withExportTimeout(ctx, runtime.Transport.Timeout)
			defer cancel()
			ctx = withOutgoingHeaders(ctx, runtime.Transport.Headers)
			return client.Export(ctx, req, grpcCallOptions(runtime.Transport.Compression)...)
		}, nil
	}
}

func (f *senderFactory) newLogsExporter(runtime otelenv.SignalRuntime) (logsExportFunc, error) {
	if err := validateRuntime(runtime); err != nil {
		return unavailableLogsExporter(runtime), nil
	}

	switch runtime.Transport.Protocol {
	case core_xds.OtelProtocolHTTPProtobuf:
		client, err := f.httpClient(runtime.Transport)
		if err != nil {
			return nil, err
		}
		targetURL := otlpHTTPURL(runtime.Transport.Endpoint, runtime.HTTPPath, pointer.Deref(runtime.Transport.UseTLS))
		return func(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
			body, err := proto.Marshal(req)
			if err != nil {
				return nil, err
			}
			if err := postOTLPHTTP(ctx, client, targetURL, body, runtime.Transport.Headers, runtime.Transport.Compression, runtime.Transport.Timeout); err != nil {
				return nil, err
			}
			return &logspb.ExportLogsServiceResponse{}, nil
		}, nil
	default:
		conn, err := f.grpcConn(runtime.Transport)
		if err != nil {
			return nil, err
		}
		client := logspb.NewLogsServiceClient(conn)
		return func(ctx context.Context, req *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
			ctx, cancel := withExportTimeout(ctx, runtime.Transport.Timeout)
			defer cancel()
			ctx = withOutgoingHeaders(ctx, runtime.Transport.Headers)
			return client.Export(ctx, req, grpcCallOptions(runtime.Transport.Compression)...)
		}, nil
	}
}

func (f *senderFactory) newMetricsExporter(runtime otelenv.SignalRuntime) (metricsExportFunc, error) {
	if err := validateRuntime(runtime); err != nil {
		return unavailableMetricsExporter(runtime), nil
	}

	switch runtime.Transport.Protocol {
	case core_xds.OtelProtocolHTTPProtobuf:
		client, err := f.httpClient(runtime.Transport)
		if err != nil {
			return nil, err
		}
		targetURL := otlpHTTPURL(runtime.Transport.Endpoint, runtime.HTTPPath, pointer.Deref(runtime.Transport.UseTLS))
		return func(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
			body, err := proto.Marshal(req)
			if err != nil {
				return nil, err
			}
			if err := postOTLPHTTP(ctx, client, targetURL, body, runtime.Transport.Headers, runtime.Transport.Compression, runtime.Transport.Timeout); err != nil {
				return nil, err
			}
			return &metricspb.ExportMetricsServiceResponse{}, nil
		}, nil
	default:
		conn, err := f.grpcConn(runtime.Transport)
		if err != nil {
			return nil, err
		}
		client := metricspb.NewMetricsServiceClient(conn)
		return func(ctx context.Context, req *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
			ctx, cancel := withExportTimeout(ctx, runtime.Transport.Timeout)
			defer cancel()
			ctx = withOutgoingHeaders(ctx, runtime.Transport.Headers)
			return client.Export(ctx, req, grpcCallOptions(runtime.Transport.Compression)...)
		}, nil
	}
}

func (f *senderFactory) grpcConn(transport otelenv.ExporterTransport) (*grpc.ClientConn, error) {
	key := grpcConnKey{
		endpoint:          transport.Endpoint,
		useTLS:            pointer.Deref(transport.UseTLS),
		certificate:       transport.Certificate,
		clientCertificate: transport.ClientCertificate,
		clientKey:         transport.ClientKey,
	}
	for _, entry := range f.grpcConns {
		if entry.key == key {
			return entry.conn, nil
		}
	}

	dialOpts := []grpc.DialOption{}
	if pointer.Deref(transport.UseTLS) {
		tlsConfig, err := newTLSConfig(transport)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(transport.Endpoint, dialOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "creating gRPC client")
	}
	f.grpcConns = append(f.grpcConns, grpcConnEntry{
		key:  key,
		conn: conn,
	})
	f.closeFns = append(f.closeFns, func() { _ = conn.Close() })
	return conn, nil
}

func (f *senderFactory) httpClient(transport otelenv.ExporterTransport) (*http.Client, error) {
	key := httpClientKey{
		endpoint:          transport.Endpoint,
		useTLS:            pointer.Deref(transport.UseTLS),
		certificate:       transport.Certificate,
		clientCertificate: transport.ClientCertificate,
		clientKey:         transport.ClientKey,
	}
	for _, entry := range f.httpClients {
		if entry.key == key {
			return entry.client, nil
		}
	}

	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	if pointer.Deref(transport.UseTLS) {
		tlsConfig, err := newTLSConfig(transport)
		if err != nil {
			return nil, err
		}
		httpTransport.TLSClientConfig = tlsConfig
	}

	client := &http.Client{
		Transport: httpTransport,
	}
	f.httpClients = append(f.httpClients, httpClientEntry{
		key:    key,
		client: client,
	})
	f.closeFns = append(f.closeFns, httpTransport.CloseIdleConnections)
	return client, nil
}

func newTLSConfig(transport otelenv.ExporterTransport) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if transport.Certificate != "" {
		certificateBytes, err := os.ReadFile(transport.Certificate)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read certificate %s", transport.Certificate)
		}
		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM(certificateBytes); !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		tlsConfig.RootCAs = rootCAs
	}

	if transport.ClientCertificate != "" || transport.ClientKey != "" {
		if transport.ClientCertificate == "" || transport.ClientKey == "" {
			return nil, errors.New("both client certificate and client key must be provided")
		}
		certificate, err := tls.LoadX509KeyPair(transport.ClientCertificate, transport.ClientKey)
		if err != nil {
			return nil, errors.Wrap(err, "could not create key pair from client cert and client key")
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}

	return tlsConfig, nil
}

func validateRuntime(runtime otelenv.SignalRuntime) error {
	if !runtime.Enabled {
		return errors.New("signal is not enabled")
	}
	if runtime.Transport.Protocol == "" {
		return errors.New("transport protocol is not configured")
	}
	if runtime.Transport.Endpoint == "" {
		return errors.New("transport endpoint is not configured")
	}
	if runtime.Transport.Protocol == core_xds.OtelProtocolHTTPProtobuf && runtime.HTTPPath == "" {
		return errors.New("http path is not configured")
	}
	return nil
}

func unavailableTraceExporter(runtime otelenv.SignalRuntime) traceExportFunc {
	return func(context.Context, *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
		return nil, signalUnavailableError("traces", runtime)
	}
}

func unavailableLogsExporter(runtime otelenv.SignalRuntime) logsExportFunc {
	return func(context.Context, *logspb.ExportLogsServiceRequest) (*logspb.ExportLogsServiceResponse, error) {
		return nil, signalUnavailableError("logs", runtime)
	}
}

func unavailableMetricsExporter(runtime otelenv.SignalRuntime) metricsExportFunc {
	return func(context.Context, *metricspb.ExportMetricsServiceRequest) (*metricspb.ExportMetricsServiceResponse, error) {
		return nil, signalUnavailableError("metrics", runtime)
	}
}

func signalUnavailableError(signal string, runtime otelenv.SignalRuntime) error {
	reason := "signal is not enabled"
	switch {
	case len(runtime.BlockedReasons) > 0:
		reason = "blocked: " + strings.Join(runtime.BlockedReasons, ", ")
	case runtime.Transport.Protocol == "":
		reason = "transport protocol is not configured"
	case runtime.Transport.Endpoint == "":
		reason = "transport endpoint is not configured"
	case runtime.HTTPPath == "" && runtime.Transport.Protocol == core_xds.OtelProtocolHTTPProtobuf:
		reason = "HTTP path is not configured"
	}
	return status.Errorf(codes.FailedPrecondition, "OTEL %s exporter is unavailable: %s", signal, reason)
}

func withExportTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func withOutgoingHeaders(ctx context.Context, headers map[string]string) context.Context {
	if len(headers) == 0 {
		return ctx
	}

	keys := slices.Sorted(maps.Keys(headers))
	pairs := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		pairs = append(pairs, key, headers[key])
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func grpcCallOptions(compression string) []grpc.CallOption {
	if compression == "gzip" {
		return []grpc.CallOption{grpc.UseCompressor(grpcgzip.Name)}
	}
	return nil
}

func postOTLPHTTP(
	ctx context.Context,
	client *http.Client,
	targetURL string,
	body []byte,
	headers map[string]string,
	compression string,
	timeout time.Duration,
) error {
	ctx, cancel := withExportTimeout(ctx, timeout)
	defer cancel()

	requestBody := body
	if compression == "gzip" {
		compressedBody, err := gzipBody(body)
		if err != nil {
			return err
		}
		requestBody = compressedBody
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	if compression == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := client.Do(req)
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
		return errors.Errorf("collector returned status %d", resp.StatusCode)
	}
	return errors.Errorf("collector returned status %d: %s", resp.StatusCode, string(responseBody))
}

func gzipBody(body []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(body); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
