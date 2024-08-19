package mux

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multizone"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/service"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

var muxServerLog = core.Log.WithName("kds-mux-server")

type OnGlobalToZoneSyncStartedFunc func(session mesh_proto.KDSSyncService_GlobalToZoneSyncClient, errorCh chan error)

func (f OnGlobalToZoneSyncStartedFunc) OnGlobalToZoneSyncStarted(session mesh_proto.KDSSyncService_GlobalToZoneSyncClient, errorCh chan error) {
	f(session, errorCh)
}

type OnZoneToGlobalSyncStartedFunc func(session mesh_proto.KDSSyncService_ZoneToGlobalSyncClient, errorCh chan error)

func (f OnZoneToGlobalSyncStartedFunc) OnZoneToGlobalSyncStarted(session mesh_proto.KDSSyncService_ZoneToGlobalSyncClient, errorCh chan error) {
	f(session, errorCh)
}

type server struct {
	config               multizone.KdsServerConfig
	CallbacksGlobal      OnGlobalToZoneSyncConnectFunc
	CallbacksZone        OnZoneToGlobalSyncConnectFunc
	metrics              core_metrics.Metrics
	serviceServer        *service.GlobalKDSServiceServer
	kdsSyncServiceServer *KDSSyncServiceServer
	streamInterceptors   []grpc.StreamServerInterceptor
	unaryInterceptors    []grpc.UnaryServerInterceptor
	mesh_proto.UnimplementedMultiplexServiceServer
}

var _ component.Component = &server{}

func NewServer(
	streamInterceptors []grpc.StreamServerInterceptor,
	unaryInterceptors []grpc.UnaryServerInterceptor,
	config multizone.KdsServerConfig,
	metrics core_metrics.Metrics,
	serviceServer *service.GlobalKDSServiceServer,
	kdsSyncServiceServer *KDSSyncServiceServer,
) component.Component {
	return &server{
		config:               config,
		metrics:              metrics,
		serviceServer:        serviceServer,
		kdsSyncServiceServer: kdsSyncServiceServer,
		streamInterceptors:   streamInterceptors,
		unaryInterceptors:    unaryInterceptors,
	}
}

func (s *server) Start(stop <-chan struct{}) error {
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
		grpc.MaxRecvMsgSize(int(s.config.MaxMsgSize)),
		grpc.MaxSendMsgSize(int(s.config.MaxMsgSize)),
	}
	grpcOptions = append(grpcOptions, s.metrics.GRPCServerInterceptors()...)
	if s.config.TlsCertFile != "" && s.config.TlsEnabled {
		cert, err := tls.LoadX509KeyPair(s.config.TlsCertFile, s.config.TlsKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to load TLS certificate")
		}
		tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
		if tlsCfg.MinVersion, err = config_types.TLSVersion(s.config.TlsMinVersion); err != nil {
			return err
		}
		if tlsCfg.MaxVersion, err = config_types.TLSVersion(s.config.TlsMaxVersion); err != nil {
			return err
		}
		if tlsCfg.CipherSuites, err = config_types.TLSCiphers(s.config.TlsCipherSuites); err != nil {
			return err
		}
		grpcOptions = append(grpcOptions, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}
	for _, interceptor := range s.streamInterceptors {
		grpcOptions = append(grpcOptions, grpc.ChainStreamInterceptor(interceptor))
	}
	grpcOptions = append(
		grpcOptions,
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	mesh_proto.RegisterGlobalKDSServiceServer(grpcServer, s.serviceServer)
	mesh_proto.RegisterKDSSyncServiceServer(grpcServer, s.kdsSyncServiceServer)
	s.metrics.RegisterGRPC(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err = grpcServer.Serve(lis); err != nil {
			muxServerLog.Error(err, "terminated with an error")
			errChan <- err
		} else {
			muxServerLog.Info("terminated normally")
		}
	}()
	muxServerLog.Info("starting", "interface", "0.0.0.0", "port", s.config.GrpcPort)

	select {
	case <-stop:
		muxServerLog.Info("stopping gracefully")
		grpcServer.GracefulStop()
		muxServerLog.Info("stopped")
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *server) NeedLeaderElection() bool {
	return false
}
