package server

import (
	"context"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server/sotw/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Server interface {
	mesh_proto.KumaDiscoveryServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	sotwServer := sotw.NewServer(context.Background(), config, callbacks)
	return &server{sotwServer}
}

var _ Server = &server{}

type server struct {
	sotw.Server
}

func (s *server) DeltaKumaResources(resourcesServer mesh_proto.KumaDiscoveryService_DeltaKumaResourcesServer) error {
	panic("not supported") // we don't use Delta XDS in KDS
}

func (s *server) StreamKumaResources(stream mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer) error {
	return s.StreamHandler(stream, "")
}

func (s *server) FetchKumaResources(ctx context.Context, request *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	panic("not supported") // we don't need to support Fetch
}
