package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type Metrics interface {
	prometheus.Registerer
	prometheus.Gatherer

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
	registry := prometheus.NewRegistry()
	registerer := prometheus.WrapRegistererWith(map[string]string{
		"zone": zone,
	}, registry)

	grpcServerMetrics := grpc_prometheus.NewServerMetrics()
	if err := registry.Register(grpcServerMetrics); err != nil {
		return nil, err
	}
	grpcServerMetrics.EnableHandlingTimeHistogram()

	grpcClientMetrics := grpc_prometheus.NewClientMetrics()
	if err := registry.Register(grpcClientMetrics); err != nil {
		return nil, err
	}
	grpcClientMetrics.EnableClientHandlingTimeHistogram()

	m := &metrics{
		Registerer:        registerer,
		Gatherer:          registry,
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
