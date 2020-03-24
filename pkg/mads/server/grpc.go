package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"

	mads_config "github.com/Kong/kuma/pkg/config/mads"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime/component"
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
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	// register services
	observability_proto.RegisterMonitoringAssignmentDiscoveryServiceServer(grpcServer, s.server)

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
