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
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	mal_dpapi "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/dpapi"
	mt_dpapi "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/dpapi"
)

var logger = core.Log.WithName("otel-receiver")

// registerFn is the function used to register a gRPC service on a server.
type registerFn func(s *grpc.Server, endpoint string) (func(), error)

// Manager listens on Unix sockets for OTel gRPC signals from Envoy and forwards them to the real collector.
// One instance handles one signal type (traces OR logs).
type Manager struct {
	registerService registerFn
	newConfig       chan []mt_dpapi.OtelBackendConfig // reused for both trace and log configs (identical shape)
	running         map[string]*runningServer         // key: socketPath
	done            chan struct{}
}

type runningServer struct {
	server      *grpc.Server
	closeClient func()
}

var _ component.GracefulComponent = &Manager{}

// NewTraceManager creates a Manager for MeshTrace OTel backends.
func NewTraceManager() *Manager {
	return &Manager{
		registerService: func(s *grpc.Server, endpoint string) (func(), error) {
			recv, closeClient, err := newTraceReceiver(endpoint)
			if err != nil {
				return nil, err
			}
			tracepb.RegisterTraceServiceServer(s, recv)
			return closeClient, nil
		},
		newConfig: make(chan []mt_dpapi.OtelBackendConfig, 1),
		running:   map[string]*runningServer{},
		done:      make(chan struct{}),
	}
}

// NewLogManager creates a Manager for MeshAccessLog OTel backends.
func NewLogManager() *Manager {
	return &Manager{
		registerService: func(s *grpc.Server, endpoint string) (func(), error) {
			recv, closeClient, err := newLogsReceiver(endpoint)
			if err != nil {
				return nil, err
			}
			logspb.RegisterLogsServiceServer(s, recv)
			return closeClient, nil
		},
		newConfig: make(chan []mt_dpapi.OtelBackendConfig, 1),
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

// OnTraceChange is the configFetcher handler for the /meshtrace path.
func (m *Manager) OnTraceChange(ctx context.Context, r io.Reader) error {
	cfg := mt_dpapi.MeshTraceDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("meshtrace dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, toGenericBackends(cfg.Backends))
}

// OnLogChange is the configFetcher handler for the /meshaccesslog path.
func (m *Manager) OnLogChange(ctx context.Context, r io.Reader) error {
	cfg := mal_dpapi.MeshAccessLogDpConfig{}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return fmt.Errorf("meshaccesslog dp config decode error: %w", err)
	}
	return m.sendConfig(ctx, toGenericBackendsFromMAL(cfg.Backends))
}

func (m *Manager) sendConfig(ctx context.Context, backends []mt_dpapi.OtelBackendConfig) error {
	select {
	case m.newConfig <- backends:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (m *Manager) reconcile(backends []mt_dpapi.OtelBackendConfig) error {
	desired := map[string]mt_dpapi.OtelBackendConfig{}
	for _, b := range backends {
		desired[b.SocketPath] = b
	}

	// stop removed backends
	for socketPath, rs := range m.running {
		if _, ok := desired[socketPath]; !ok {
			rs.server.GracefulStop()
			rs.closeClient()
			delete(m.running, socketPath)
		}
	}

	// start new backends
	for socketPath, b := range desired {
		if _, ok := m.running[socketPath]; ok {
			continue // already running; reconfiguration not yet supported
		}
		rs, err := m.startBackend(socketPath, b.Endpoint)
		if err != nil {
			return errors.Wrapf(err, "failed to start OTel receiver on %s", socketPath)
		}
		m.running[socketPath] = rs
	}
	return nil
}

func (m *Manager) startBackend(socketPath, endpoint string) (*runningServer, error) {
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "removing existing socket %s", socketPath)
	}
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, errors.Wrapf(err, "listening on %s", socketPath)
	}

	s := grpc.NewServer()
	closeClient, err := m.registerService(s, endpoint)
	if err != nil {
		_ = lis.Close()
		return nil, err
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "OTel receiver stopped", "socketPath", socketPath)
		}
	}()

	logger.Info("started OTel receiver", "socketPath", socketPath, "endpoint", endpoint)
	return &runningServer{server: s, closeClient: closeClient}, nil
}

func (m *Manager) stopAll() {
	for socketPath, rs := range m.running {
		rs.server.GracefulStop()
		rs.closeClient()
		delete(m.running, socketPath)
	}
}

// toGenericBackends converts meshtrace dpapi backends to the internal format.
func toGenericBackends(in []mt_dpapi.OtelBackendConfig) []mt_dpapi.OtelBackendConfig {
	return in
}

// toGenericBackendsFromMAL converts meshaccesslog dpapi backends to the same internal format.
func toGenericBackendsFromMAL(in []mal_dpapi.OtelBackendConfig) []mt_dpapi.OtelBackendConfig {
	out := make([]mt_dpapi.OtelBackendConfig, len(in))
	for i, b := range in {
		out[i] = mt_dpapi.OtelBackendConfig{
			SocketPath: b.SocketPath,
			Endpoint:   b.Endpoint,
			UseHTTP:    b.UseHTTP,
			Path:       b.Path,
		}
	}
	return out
}
