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
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

var logger = core.Log.WithName("otel-receiver")

// Manager listens on Unix sockets for OTel gRPC signals (traces, logs, metrics)
// from Envoy and kuma-dp, and forwards them to the real collector.
// Each backend gets one socket with all three gRPC services registered.
type Manager struct {
	newConfig   chan []core_xds.OtelPipeBackend
	running     map[string]*runningServer // key: socketPath
	done        chan struct{}
	envConfig   otelenv.Config
	onReconcile func([]core_xds.OtelPipeBackend)
}

type runningServer struct {
	server   *grpc.Server
	closeFns []func()
	runtime  otelenv.BackendRuntime
	stopped  chan struct{} // closed when the Serve goroutine exits
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

// SetOnReconcile sets a callback that fires after each reconcile with the
// current backends list. Used by run.go to bridge OTEL metric export targets
// to meshmetrics.Manager.
func (m *Manager) SetOnReconcile(fn func([]core_xds.OtelPipeBackend)) {
	m.onReconcile = fn
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

func (m *Manager) sendConfig(_ context.Context, backends []core_xds.OtelPipeBackend) error {
	select {
	case m.newConfig <- backends:
	default:
		// Drop stale value, push new one.
		select {
		case <-m.newConfig:
		default:
		}
		m.newConfig <- backends
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

	// start new backends and restart updated or dead backends
	for socketPath, b := range desired {
		runtime := m.envConfig.ResolveBackend(b)
		if rs, ok := m.running[socketPath]; ok {
			if !rs.isStopped() && sameBackendRuntime(rs.runtime, runtime) {
				continue
			}
			stopRunningServer(rs)
			delete(m.running, socketPath)
		}

		rs, err := m.startBackend(socketPath, runtime)
		if err != nil {
			// Notify about successfully started backends before returning
			// so meshmetrics can begin exporting to them.
			m.notifyReconcile(backends)
			return errors.Wrapf(err, "failed to start OTel receiver on %s", socketPath)
		}
		m.running[socketPath] = rs
	}
	m.notifyReconcile(backends)
	return nil
}

func (m *Manager) startBackend(
	socketPath string,
	runtime otelenv.BackendRuntime,
) (*runningServer, error) {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "removing existing socket %s", socketPath)
	}
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "unix", socketPath)
	if err != nil {
		return nil, errors.Wrapf(err, "listening on %s", socketPath)
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = lis.Close()
		return nil, errors.Wrapf(err, "chmod socket %s", socketPath)
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

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "OTel receiver stopped", "socketPath", socketPath)
		}
	}()

	logger.Info(
		"started OTel receiver",
		"socketPath", socketPath,
		"tracesEnabled", runtime.Traces.Enabled,
		"logsEnabled", runtime.Logs.Enabled,
		"metricsEnabled", runtime.Metrics.Enabled,
	)
	return &runningServer{server: s, closeFns: closeFns, runtime: runtime, stopped: stopped}, nil
}

func (m *Manager) notifyReconcile(backends []core_xds.OtelPipeBackend) {
	if m.onReconcile == nil || len(m.running) == 0 {
		return
	}
	// Filter to backends that are actually running so meshmetrics
	// doesn't try to dial sockets that were never started.
	var running []core_xds.OtelPipeBackend
	for _, b := range backends {
		if _, ok := m.running[b.SocketPath]; ok {
			running = append(running, b)
		}
	}
	if len(running) > 0 {
		m.onReconcile(running)
	}
}

func (m *Manager) stopAll() {
	for socketPath, runningServer := range m.running {
		stopRunningServer(runningServer)
		delete(m.running, socketPath)
	}
}

func (rs *runningServer) isStopped() bool {
	select {
	case <-rs.stopped:
		return true
	default:
		return false
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
