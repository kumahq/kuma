package service

import (
	"context"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server/sotw/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"

	observability_v1alpha1 "github.com/kumahq/kuma/api/observability/v1alpha1"
	mads_v1alpha1 "github.com/kumahq/kuma/pkg/mads/v1alpha1"
)

type Server interface {
	observability_v1alpha1.MonitoringAssignmentDiscoveryServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks) Server {
	sotwServer := sotw.NewServer(context.Background(), config, callbacks)
	return &server{sotwServer}
}

var _ Server = &server{}

type server struct {
	sotw.Server
}

func (s *server) DeltaMonitoringAssignments(stream observability_v1alpha1.MonitoringAssignmentDiscoveryService_DeltaMonitoringAssignmentsServer) error {
	panic("not implemented") // we don't use delta on MADS for now
}

func (s *server) StreamMonitoringAssignments(stream observability_v1alpha1.MonitoringAssignmentDiscoveryService_StreamMonitoringAssignmentsServer) error {
	return s.StreamHandler(stream, mads_v1alpha1.MonitoringAssignmentType)
}

func (s *server) FetchMonitoringAssignments(ctx context.Context, request *envoy_api.DiscoveryRequest) (*envoy_api.DiscoveryResponse, error) {
	panic("not implemented") // we don't need Fetch operation on MADS for now
}
