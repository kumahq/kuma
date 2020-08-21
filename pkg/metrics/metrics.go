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
}

type metrics struct {
	*prometheus.Registry
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

func NewMetrics() (Metrics, error) {
	registry := prometheus.NewRegistry()

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
		Registry:          registry,
		grpcServerMetrics: grpcServerMetrics,
		grpcClientMetrics: grpcClientMetrics,
	}
	return m, nil
}
