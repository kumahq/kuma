package service

import (
	"context"

	envoy_sd "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/rest/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

type Server interface {
	observability_v1.MonitoringAssignmentDiscoveryServiceServer
}

func NewServer(cache envoy_cache.Cache, callbacks envoy_server.Callbacks) Server {
	restServer := rest.NewServer(cache, callbacks)
	return &server{rest: restServer}
}

var _ Server = &server{}

type server struct {
	rest rest.Server
	observability_v1.UnimplementedMonitoringAssignmentDiscoveryServiceServer
}

func (s *server) FetchMonitoringAssignments(ctx context.Context, request *envoy_sd.DiscoveryRequest) (*envoy_sd.DiscoveryResponse, error) {
	return s.rest.Fetch(ctx, request)
}
