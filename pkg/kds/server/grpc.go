package server

import (
	"fmt"
	"net"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kds_config "github.com/Kong/kuma/pkg/config/kds"

	"google.golang.org/grpc"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime/component"
)

const grpcMaxConcurrentStreams = 1000000

var (
	grpcServerLog = core.Log.WithName("kds-server").WithName("grpc")
)

type grpcServer struct {
	server Server
	config kds_config.KumaDiscoveryServerConfig
}

var (
	_ component.Component = &grpcServer{}
)

func NewKDSServer(srv Server, config kds_config.KumaDiscoveryServerConfig) component.Component {
	return &grpcServer{server: srv, config: config}
}

func (s *grpcServer) Start(stop <-chan struct{}) error {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	// register services
	mesh_proto.RegisterKumaDiscoveryServiceServer(grpcServer, s.server)

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
