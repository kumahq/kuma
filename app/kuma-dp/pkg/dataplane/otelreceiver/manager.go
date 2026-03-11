package otelreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"slices"

	"github.com/pkg/errors"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/otelenv"
	"github.com/kumahq/kuma/v2/pkg/core"
	motb_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshopentelemetrybackend/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	mal_dpapi "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/dpapi"
	mt_dpapi "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/dpapi"
)

var logger = core.Log.WithName("otel-receiver")

// Manager listens on Unix sockets for OTel gRPC signals (traces, logs, metrics)
// from Envoy and kuma-dp, and forwards them to the real collector.
// Each backend gets one socket with all three gRPC services registered.
type Manager struct {
	newConfig chan []core_xds.OtelPipeBackend
	running   map[string]*runningServer // key: socketPath
	done      chan struct{}
	envConfig otelenv.Config
}

type runningServer struct {
	server   *grpc.Server
	closeFns []func()
	runtime  otelenv.BackendRuntime
}

var _ component.GracefulComponent = &Manager{}

// NewManager creates a unified Manager that registers all three OTel gRPC
// services (traces, logs, metrics) on each backend socket.
func NewManager(envConfig otelenv.Config) *Manager {
	return &Manager{
		newConfig: make(chan []core_xds.OtelPipeBackend, 1),
		running:   map[string]*runningServer{},
		done:      make(chan struct{}),
		envConfig: envConfig,
	}
}

func (m *Manager) NeedLeaderElection() bool { return false }

func (m *Manager) WaitForDone() { <-m.done }

func (m *Manager) Start(stop <-chan struct{}) error {
	defer func() {
		m.stopAll()
		close(m.done)
	}()
	for {
		select {
		case backends := <-m.newConfig:
			if err := m.reconcile(backends); err != nil {
				logger.Error(err, "failed to reconcile OTel receiver backends")
			}
		case <-stop:
			return nil
		}
	}
}

// OnOtelChange is the configFetcher handler for the unified /otel path.
func (m *Manager) OnOtelChange(ctx context.Context, r io.Reader) error {
	cfg := core_xds.OtelDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("otel dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, cfg.Backends)
}

// OnTraceChange is the legacy configFetcher handler for /meshtrace.
// Kept for backward compat with older CPs that still send per-signal dynconf.
func (m *Manager) OnTraceChange(ctx context.Context, r io.Reader) error {
	cfg := mt_dpapi.MeshTraceDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("meshtrace dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, markLegacySignal(cfg.Backends, core_xds.OtelSignalTraces))
}

// OnLogChange is the legacy configFetcher handler for /meshaccesslog.
// Kept for backward compat with older CPs that still send per-signal dynconf.
func (m *Manager) OnLogChange(ctx context.Context, r io.Reader) error {
	cfg := mal_dpapi.MeshAccessLogDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("meshaccesslog dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, markLegacySignal(cfg.Backends, core_xds.OtelSignalLogs))
}

func (m *Manager) sendConfig(ctx context.Context, backends []core_xds.OtelPipeBackend) error {
	select {
	case m.newConfig <- backends:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (m *Manager) reconcile(backends []core_xds.OtelPipeBackend) error {
	desired := map[string]core_xds.OtelPipeBackend{}
	for _, b := range backends {
		desired[b.SocketPath] = b
	}

	// stop removed backends
	for socketPath, rs := range m.running {
		if _, ok := desired[socketPath]; !ok {
			stopRunningServer(rs)
			delete(m.running, socketPath)
		}
	}

	// start new backends and restart updated backends
	for socketPath, b := range desired {
		runtime := m.envConfig.ResolveBackend(b)
		if rs, ok := m.running[socketPath]; ok {
			if sameBackendRuntime(rs.runtime, runtime) {
				continue
			}
			stopRunningServer(rs)
			delete(m.running, socketPath)
		}

		rs, err := m.startBackend(socketPath, b, runtime)
		if err != nil {
			return errors.Wrapf(err, "failed to start OTel receiver on %s", socketPath)
		}
		m.running[socketPath] = rs
	}
	return nil
}

func (m *Manager) startBackend(
	socketPath string,
	backend core_xds.OtelPipeBackend,
	runtime otelenv.BackendRuntime,
) (*runningServer, error) {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "removing existing socket %s", socketPath)
	}
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, errors.Wrapf(err, "listening on %s", socketPath)
	}

	senders := &senderFactory{}
	traceExporter, err := senders.newTraceExporter(runtime.Traces)
	if err != nil {
		senders.close()
		_ = lis.Close()
		return nil, err
	}
	logsExporter, err := senders.newLogsExporter(runtime.Logs)
	if err != nil {
		senders.close()
		_ = lis.Close()
		return nil, err
	}
	metricsExporter, err := senders.newMetricsExporter(runtime.Metrics)
	if err != nil {
		senders.close()
		_ = lis.Close()
		return nil, err
	}

	s := grpc.NewServer()
	tracepb.RegisterTraceServiceServer(s, newTraceReceiver(traceExporter))
	logspb.RegisterLogsServiceServer(s, newLogsReceiver(logsExporter))
	metricspb.RegisterMetricsServiceServer(s, newMetricsReceiver(metricsExporter))

	closeFns := append([]func(){func() { _ = lis.Close() }}, senders.closeFns...)

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "OTel receiver stopped", "socketPath", socketPath)
		}
	}()

	logger.Info(
		"started OTel receiver",
		"socketPath", socketPath,
		"clientLayout", backend.ClientLayout,
		"tracesEnabled", runtime.Traces.Enabled,
		"logsEnabled", runtime.Logs.Enabled,
		"metricsEnabled", runtime.Metrics.Enabled,
	)
	return &runningServer{server: s, closeFns: closeFns, runtime: runtime}, nil
}

func (m *Manager) stopAll() {
	for socketPath, runningServer := range m.running {
		stopRunningServer(runningServer)
		delete(m.running, socketPath)
	}
}

func stopRunningServer(rs *runningServer) {
	rs.server.GracefulStop()
	for _, fn := range rs.closeFns {
		fn()
	}
}

func sameBackendRuntime(a, b otelenv.BackendRuntime) bool {
	return sameSignalRuntime(a.Traces, b.Traces) &&
		sameSignalRuntime(a.Logs, b.Logs) &&
		sameSignalRuntime(a.Metrics, b.Metrics)
}

func sameSignalRuntime(a, b otelenv.SignalRuntime) bool {
	return a.Enabled == b.Enabled &&
		slices.Equal(a.BlockedReasons, b.BlockedReasons) &&
		a.HTTPPath == b.HTTPPath &&
		sameTransport(a.Transport, b.Transport)
}

func sameTransport(a, b otelenv.ExporterTransport) bool {
	return a.Protocol == b.Protocol &&
		a.Endpoint == b.Endpoint &&
		sameBoolPtr(a.UseTLS, b.UseTLS) &&
		maps.Equal(a.Headers, b.Headers) &&
		a.Compression == b.Compression &&
		a.Timeout == b.Timeout &&
		a.Certificate == b.Certificate &&
		a.ClientCertificate == b.ClientCertificate &&
		a.ClientKey == b.ClientKey
}

func sameBoolPtr(a, b *bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func markLegacySignal(backends []core_xds.OtelPipeBackend, signal core_xds.OtelSignal) []core_xds.OtelPipeBackend {
	result := make([]core_xds.OtelPipeBackend, len(backends))
	for i, backend := range backends {
		result[i] = backend
		result[i].EnvPolicy = core_xds.OtelResolvedEnvPolicy{
			Mode:                 motb_api.EnvModeDisabled,
			Precedence:           motb_api.EnvPrecedenceExplicitFirst,
			AllowSignalOverrides: false,
		}
		result[i].Traces = nil
		result[i].Logs = nil
		result[i].Metrics = nil

		plan := core_xds.OtelSignalRuntimePlan{
			Enabled:        true,
			BlockedReasons: []string{core_xds.OtelBlockedReasonEnvDisabledByPlatform},
		}
		switch signal {
		case core_xds.OtelSignalTraces:
			result[i].Traces = &plan
		case core_xds.OtelSignalLogs:
			result[i].Logs = &plan
		case core_xds.OtelSignalMetrics:
			result[i].Metrics = &plan
		}
	}
	return result
}
