package otelreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/pkg/errors"
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/v2/pkg/core"
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
}

type runningServer struct {
	server   *grpc.Server
	closeFns []func()
	backend  core_xds.OtelPipeBackend
}

var _ component.GracefulComponent = &Manager{}

// NewManager creates a unified Manager that registers all three OTel gRPC
// services (traces, logs, metrics) on each backend socket.
func NewManager() *Manager {
	return &Manager{
		newConfig: make(chan []core_xds.OtelPipeBackend, 1),
		running:   map[string]*runningServer{},
		done:      make(chan struct{}),
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
	return m.sendConfig(ctx, cfg.Backends)
}

// OnLogChange is the legacy configFetcher handler for /meshaccesslog.
// Kept for backward compat with older CPs that still send per-signal dynconf.
func (m *Manager) OnLogChange(ctx context.Context, r io.Reader) error {
	cfg := mal_dpapi.MeshAccessLogDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("meshaccesslog dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, cfg.Backends)
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
		if rs, ok := m.running[socketPath]; ok {
			if sameBackendConfig(rs.backend, b) {
				continue
			}
			stopRunningServer(rs)
			delete(m.running, socketPath)
		}

		rs, err := m.startBackend(socketPath, b)
		if err != nil {
			return errors.Wrapf(err, "failed to start OTel receiver on %s", socketPath)
		}
		m.running[socketPath] = rs
	}
	return nil
}

func (m *Manager) startBackend(socketPath string, backend core_xds.OtelPipeBackend) (*runningServer, error) {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "removing existing socket %s", socketPath)
	}
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, errors.Wrapf(err, "listening on %s", socketPath)
	}

	s := grpc.NewServer()
	var closeFns []func()
	cleanup := func() {
		_ = lis.Close()
		for _, fn := range closeFns {
			fn()
		}
	}

	traceRecv, traceClose, err := newTraceReceiver(backend.Endpoint, backend.UseHTTP, backend.Path)
	if err != nil {
		cleanup()
		return nil, errors.Wrap(err, "creating trace receiver")
	}
	tracepb.RegisterTraceServiceServer(s, traceRecv)
	closeFns = append(closeFns, traceClose)

	logsRecv, logsClose, err := newLogsReceiver(backend.Endpoint, backend.UseHTTP, backend.Path)
	if err != nil {
		cleanup()
		return nil, errors.Wrap(err, "creating logs receiver")
	}
	logspb.RegisterLogsServiceServer(s, logsRecv)
	closeFns = append(closeFns, logsClose)

	metricsRecv, metricsClose, err := newMetricsReceiver(backend.Endpoint, backend.UseHTTP, backend.Path)
	if err != nil {
		cleanup()
		return nil, errors.Wrap(err, "creating metrics receiver")
	}
	metricspb.RegisterMetricsServiceServer(s, metricsRecv)
	closeFns = append(closeFns, metricsClose)

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "OTel receiver stopped", "socketPath", socketPath)
		}
	}()

	logger.Info("started OTel receiver", "socketPath", socketPath, "endpoint", backend.Endpoint, "useHTTP", backend.UseHTTP, "path", backend.Path)
	return &runningServer{server: s, closeFns: closeFns, backend: backend}, nil
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

func sameBackendConfig(a, b core_xds.OtelPipeBackend) bool {
	return a.Endpoint == b.Endpoint &&
		a.UseHTTP == b.UseHTTP &&
		a.Path == b.Path
}
