package sds

import (
	"fmt"
	"net"

	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"google.golang.org/grpc"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	sds_server "github.com/Kong/kuma/pkg/sds/server"
)

const grpcMaxConcurrentStreams = 1000000

var (
	grpcServerLog = core.Log.WithName("sds-server").WithName("grpc")
)

type grpcServer struct {
	server  sds_server.Server
	address string
}

var (
	_ component.Component = &grpcServer{}
)

func (s *grpcServer) Start(stop <-chan struct{}) error {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("unix", s.address)
	if err != nil {
		return err
	}

	// register services
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
	grpcServerLog.Info("starting SDS server", "address", fmt.Sprintf("unix://%s", s.address))

	select {
	case <-stop:
		grpcServerLog.Info("stopping gracefully")
		grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}
