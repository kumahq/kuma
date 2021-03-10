package server

import (
	"fmt"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)


const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

var (
	grpcServerLog = core.Log.WithName("mads-server").WithName("grpc")
)

// grpcServer is a runtime component.Component that
// serves all MADs resources over gRPC
type grpcServer struct {
	services []GrpcService
	config   *mads_config.MonitoringAssignmentServerConfig
	metrics  core_metrics.Metrics
}

type GrpcService interface {
	RegisterWithGrpcServer(server *grpc.Server)
}

var (
	_ component.Component = &grpcServer{}
)

func (s *grpcServer) Start(stop <-chan struct{}) error {
	grpcOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepAliveTime,
			Timeout: grpcKeepAliveTime,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepAliveTime,
			PermitWithoutStream: true,
		}),
	}
	grpcOptions = append(grpcOptions, s.metrics.GRPCServerInterceptors()...)
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	server := grpc.NewServer(grpcOptions...)

	// register services
	for _, service := range s.services {
		service.RegisterWithGrpcServer(server)
	}

	s.metrics.RegisterGRPC(server)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err = server.Serve(lis); err != nil {
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
		server.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *grpcServer) NeedLeaderElection() bool {
	return false
}
