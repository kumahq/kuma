package mux

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multizone"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/service"
	"github.com/kumahq/kuma/pkg/kds/util"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

var (
	muxServerLog = core.Log.WithName("kds-mux-server")
)

type Filter interface {
	InterceptSession(session Session) error
}

type Callbacks interface {
	OnSessionStarted(session Session) error
}
type OnSessionStartedFunc func(session Session) error

func (f OnSessionStartedFunc) OnSessionStarted(session Session) error {
	return f(session)
}

type server struct {
	config        multizone.KdsServerConfig
	callbacks     Callbacks
	filters       []Filter
	metrics       core_metrics.Metrics
	serviceServer *service.GlobalKDSServiceServer
	mesh_proto.UnimplementedMultiplexServiceServer
}

var (
	_ component.Component = &server{}
)

func NewServer(
	callbacks Callbacks,
	filters []Filter,
	config multizone.KdsServerConfig,
	metrics core_metrics.Metrics,
	serviceServer *service.GlobalKDSServiceServer,
) component.Component {
	return &server{
		callbacks:     callbacks,
		filters:       filters,
		config:        config,
		metrics:       metrics,
		serviceServer: serviceServer,
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
	if s.config.TlsCertFile != "" && s.config.TlsDisable == false {
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
	grpcServer := grpc.NewServer(grpcOptions...)

	// register services
	mesh_proto.RegisterMultiplexServiceServer(grpcServer, s)
	mesh_proto.RegisterGlobalKDSServiceServer(grpcServer, s.serviceServer)
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
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *server) StreamMessage(stream mesh_proto.MultiplexService_StreamMessageServer) error {
	clientID, err := util.ClientIDFromIncomingCtx(stream.Context())
	if err != nil {
		return err
	}
	log := muxServerLog.WithValues("client-id", clientID)
	log.Info("initializing Kuma Discovery Service (KDS) stream for global-zone sync of resources")
	// The buffer size should be of a size of all inflight request, so we never write to a blocked buffer.
	// The buffer is separate for each direction (send/receive) on each multiplexed stream (global acting as server/global acting as client)
	// A CP never sends multiple DiscoveryRequests for one resource type.
	// A CP never sends multiple DiscoveryResponses for one resource type (it waits until peer answers with ACK/NACK)
	// Therefore the maximum number of inflight requests are number of synced resources.
	// For the simplicity we just take all resources available in Kuma (.
	bufferSize := len(registry.Global().ObjectTypes())
	session := NewSession(clientID, stream, uint32(bufferSize), s.config.MsgSendTimeout.Duration)
	for _, filter := range s.filters {
		if err := filter.InterceptSession(session); err != nil {
			log.Error(err, "closing KDS stream following a callback error")
			return err
		}
	}
	if err := s.callbacks.OnSessionStarted(session); err != nil {
		log.Error(err, "closing KDS stream following a callback error")
		return err
	}
	err = <-session.Error()
	log.Info("KDS stream is closed", "reason", err.Error())
	return nil
}

func (s *server) NeedLeaderElection() bool {
	return false
}
