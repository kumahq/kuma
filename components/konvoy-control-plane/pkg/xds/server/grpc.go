package server

import (
	"fmt"
	"net"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
)

const grpcMaxConcurrentStreams = 1000000

var (
	grpcServerLog = core.Log.WithName("xds-server").WithName("grpc")
)

type grpcServer struct {
	server envoy_xds.Server
	port   int
}

// Make sure that grpcServer implements all relevant interfaces
var (
	_ core_runtime.Component = &grpcServer{}
)

func (s *grpcServer) Start(stop <-chan struct{}) error {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	// register services
	envoy_discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, s.server)
	envoy.RegisterEndpointDiscoveryServiceServer(grpcServer, s.server)
	envoy.RegisterClusterDiscoveryServiceServer(grpcServer, s.server)
	envoy.RegisterRouteDiscoveryServiceServer(grpcServer, s.server)
	envoy.RegisterListenerDiscoveryServiceServer(grpcServer, s.server)
	envoy_discovery.RegisterSecretDiscoveryServiceServer(grpcServer, s.server)

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err = grpcServer.Serve(lis); err != nil {
			grpcServerLog.Error(err, "terminated with an error")
			errChan <- err
		} else {
			grpcServerLog.Info("terminated normally")
		}
	}()
	grpcServerLog.Info("starting", "port", s.port)

	select {
	case <-stop:
		grpcServerLog.Info("stopping gracefully")
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}
