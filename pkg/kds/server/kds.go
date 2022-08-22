package server

import (
	"context"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/sotw/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Server interface {
	mesh_proto.KumaDiscoveryServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	sotwServer := sotw.NewServer(context.Background(), config, callbacks)
	return &server{Server: sotwServer}
}

var _ Server = &server{}

type server struct {
	sotw.Server
	mesh_proto.UnimplementedKumaDiscoveryServiceServer
}

func (s *server) StreamKumaResources(stream mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer) error {
	return s.Server.StreamHandler(stream, "")
}
