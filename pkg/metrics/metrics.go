package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	config_core "github.com/kumahq/kuma/pkg/config/core"
)

type RegistererGatherer interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type Metrics interface {
	RegistererGatherer

	RegisterGRPC(*grpc.Server)
	GRPCServerInterceptors() []grpc.ServerOption
	GRPCClientInterceptors() []grpc.DialOption
	BulkRegister(...prometheus.Collector) error
}

type metrics struct {
	prometheus.Registerer
	prometheus.Gatherer
	grpcServerMetrics *grpc_prometheus.ServerMetrics
	grpcClientMetrics *grpc_prometheus.ClientMetrics
}

func (m *metrics) RegisterGRPC(server *grpc.Server) {
	m.grpcServerMetrics.InitializeMetrics(server)
}

func (m *metrics) GRPCServerInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StreamInterceptor(m.grpcServerMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(m.grpcServerMetrics.UnaryServerInterceptor()),
	}
}

func (m *metrics) GRPCClientInterceptors() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStreamInterceptor(m.grpcClientMetrics.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(m.grpcClientMetrics.UnaryClientInterceptor()),
	}
}

var _ Metrics = &metrics{}

func NewMetrics(zone string) (Metrics, error) {
	return NewMetricsOfRegistererGatherer(zone, prometheus.NewRegistry())
}

func NewMetricsOfRegistererGatherer(zone string, registererGatherer RegistererGatherer) (Metrics, error) {
	wrappedRegisterer := prometheus.WrapRegistererWith(map[string]string{
		"zone": zone,
	}, registererGatherer)

	grpcServerMetrics := grpc_prometheus.NewServerMetrics()
	if err := wrappedRegisterer.Register(grpcServerMetrics); err != nil {
		return nil, err
	}
	grpcServerMetrics.EnableHandlingTimeHistogram()

	grpcClientMetrics := grpc_prometheus.NewClientMetrics()
	if err := wrappedRegisterer.Register(grpcClientMetrics); err != nil {
		return nil, err
	}
	grpcClientMetrics.EnableClientHandlingTimeHistogram()

	m := &metrics{
		Registerer:        wrappedRegisterer,
		Gatherer:          registererGatherer,
		grpcServerMetrics: grpcServerMetrics,
		grpcClientMetrics: grpcClientMetrics,
	}
	return m, nil
}

func (m *metrics) BulkRegister(cs ...prometheus.Collector) error {
	for _, c := range cs {
		if err := m.Register(c); err != nil {
			return err
		}
	}
	return nil
}

func ZoneNameOrMode(mode config_core.CpMode, name string) string {
	switch mode {
	case config_core.Zone:
		if name == "" {
			return "Zone"
		}
		return name
	case config_core.Global:
		return "Global"
	default:
		return ""
	}
}
