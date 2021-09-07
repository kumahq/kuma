package mux

import (
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multizone"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

var (
	muxServerLog = core.Log.WithName("kds-mux-server")
)

type Callbacks interface {
	OnSessionStarted(session Session) error
}
type OnSessionStartedFunc func(session Session) error

func (f OnSessionStartedFunc) OnSessionStarted(session Session) error {
	return f(session)
}

type server struct {
	config    multizone.KdsServerConfig
	callbacks Callbacks
	metrics   core_metrics.Metrics
}

var (
	_ component.Component = &server{}
)

func NewServer(callbacks Callbacks, config multizone.KdsServerConfig, metrics core_metrics.Metrics) component.Component {
	return &server{
		callbacks: callbacks,
		config:    config,
		metrics:   metrics,
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
	useTLS := s.config.TlsCertFile != ""
	if useTLS {
		creds, err := credentials.NewServerTLSFromFile(s.config.TlsCertFile, s.config.TlsKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to load TLS certificate")
		}
		grpcOptions = append(grpcOptions, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(grpcOptions...)

	// register services
	mesh_proto.RegisterMultiplexServiceServer(grpcServer, s)
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
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("metadata is not provided")
	}
	if len(md["client-id"]) == 0 {
		return errors.New("'client-id' is not present in metadata")
	}
	clientID := md["client-id"][0]
	log := muxServerLog.WithValues("client-id", clientID)
	log.Info("initializing Kuma Discovery Service (KDS) stream for global-zone sync of resources")
	session := NewSession(clientID, stream)
	if err := s.callbacks.OnSessionStarted(session); err != nil {
		log.Error(err, "closing KDS stream following a callback error")
		return err
	}
	err := <-session.Error()
	log.Info("KDS stream is closed", "reason", err.Error())
	return nil
}

func (s *server) NeedLeaderElection() bool {
	return false
}
