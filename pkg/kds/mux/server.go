package mux

import (
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/multicluster"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

const grpcMaxConcurrentStreams = 1000000

var (
	muxServerLog = core.Log.WithName("mux-server")
)

type Callbacks interface {
	OnSessionStarted(session Session) error
}
type OnSessionStartedFunc func(session Session) error

func (f OnSessionStartedFunc) OnSessionStarted(session Session) error {
	return f(session)
}

type server struct {
	config    multicluster.KdsServerConfig
	callbacks Callbacks
}

var (
	_ component.Component = &server{}
)

func NewServer(callbacks Callbacks, config multicluster.KdsServerConfig) component.Component {
	return &server{callbacks: callbacks, config: config}
}

func (s *server) Start(stop <-chan struct{}) error {
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpcOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	}
	useTLS := s.config.TlsCertFile != ""
	if useTLS {
		creds, err := credentials.NewServerTLSFromFile(s.config.TlsCertFile, s.config.TlsKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to load TLS certificate")
		}
		grpcOptions = append(grpcOptions, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GrpcPort))
	if err != nil {
		return err
	}

	// register services
	mesh_proto.RegisterMultiplexServiceServer(grpcServer, s)
	grpc_prometheus.Register(grpcServer)

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
	log.Info("initializing KDS stream", "client-id", clientID)
	stop := make(chan struct{})
	session := NewSession(clientID, stream, stop)
	defer close(stop)
	if err := s.callbacks.OnSessionStarted(session); err != nil {
		return err
	}
	<-stream.Context().Done()
	log.Info("KDS stream is closed")
	return nil
}

func (s *server) NeedLeaderElection() bool {
	return false
}
