package server

import (
	"context"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/delta/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// Server is common for global and zone
type Server interface {
	ZoneToGlobal(stream.DeltaStream) error
	mesh_proto.KDSSyncServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	deltaServer := delta.NewServer(context.Background(), config, callbacks, delta.WithDistinctResourceTypes(1000))
	return &server{Server: deltaServer}
}

var _ Server = &server{}

type server struct {
	delta.Server
	mesh_proto.UnimplementedKDSSyncServiceServer
}

func (s *server) GlobalToZoneSync(stream mesh_proto.KDSSyncService_GlobalToZoneSyncServer) error {
	errorStream := NewErrorRecorderStream(stream)
	err := s.Server.DeltaStreamHandler(errorStream, "")
	if err == nil {
		err = errorStream.Err()
	}
	return err
}

// Delta xDS server expects `KDSSyncService_ZoneToGlobalSyncServer` to have Send(*v3.DeltaDiscoveryResponse)
// and Recv() (*v3.DeltaDiscoveryRequest, error) but proto has different definition to make it works for
// synchronization from Zone to Global.
func (s *server) ZoneToGlobalSync(stream mesh_proto.KDSSyncService_ZoneToGlobalSyncServer) error {
	panic("not implemented")
}

// ZoneToGlobal is the custom implementation for `ZoneToGlobalSync` to support running delta server
// on zone while kds.proto has different definition of `KDSSyncService_ZoneToGlobalSyncServer` then
// expected by delta xDS server.
func (s *server) ZoneToGlobal(stream stream.DeltaStream) error {
	errorStream := NewErrorRecorderStream(stream)
	err := s.Server.DeltaStreamHandler(errorStream, "")
	if err == nil {
		err = errorStream.Err()
	}
	return err
}
