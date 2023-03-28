package server

import (
	"context"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/delta/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

type Server interface {
	mesh_proto.KDSSyncServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	deltaServer := delta.NewServer(context.Background(), config, callbacks)
	return &server{Server: deltaServer}
}

var _ Server = &server{}

type server struct {
	delta.Server
	mesh_proto.UnimplementedKDSSyncServiceServer
}

func (s *server) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	return s.Server.DeltaStreamHandler(stream, "")
}
