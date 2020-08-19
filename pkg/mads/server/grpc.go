package server

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"

	observability_proto "github.com/kumahq/kuma/api/observability/v1alpha1"

	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

const grpcMaxConcurrentStreams = 1000000

var (
	grpcServerLog = core.Log.WithName("mads-server").WithName("grpc")
)

type grpcServer struct {
	server Server
	config mads_config.MonitoringAssignmentServerConfig
}

var (
	_ component.Component = &grpcServer{}
)

func (s *grpcServer) Start(stop <-chan struct{}) error {
	grpcOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	}
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	// register services
	observability_proto.RegisterMonitoringAssignmentDiscoveryServiceServer(grpcServer, s.server)
	grpc_prometheus.Register(grpcServer)

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
	grpcServerLog.Info("starting", "interface", "0.0.0.0", "port", s.config.GrpcPort)

	select {
	case <-stop:
		grpcServerLog.Info("stopping gracefully")
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *grpcServer) NeedLeaderElection() bool {
	return false
}
