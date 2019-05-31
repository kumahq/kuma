package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

const grpcMaxConcurrentStreams = 1000000

var (
	grpcServerLog = ctrl.Log.WithName("xds-server").WithName("grpc")
)

type grpcServer struct {
	server xds.Server
	port   int
}

// Make sure that grpcServer implements all relevant interfaces
var (
	_ manager.Runnable               = &grpcServer{}
	_ manager.LeaderElectionRunnable = &grpcServer{}
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
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, s.server)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, s.server)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, s.server)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, s.server)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, s.server)
	discovery.RegisterSecretDiscoveryServiceServer(grpcServer, s.server)

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err = grpcServer.Serve(lis); err != nil {
			grpcServerLog.Error(err, "terminated with an error")
			errChan <- err
		}
	}()
	grpcServerLog.Info("starting", "port", s.port)

	select {
	case <-stop:
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *grpcServer) NeedLeaderElection() bool {
	return false
}
