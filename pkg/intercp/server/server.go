package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc/filters"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/config/intercp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("intercp-server")

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

type InterCpServer struct {
	config     intercp.InterCpServerConfig
	grpcServer *grpc.Server
	instanceId string
	ready      atomic.Bool
}

var _ component.Component = &InterCpServer{}

func New(
	config intercp.InterCpServerConfig,
	metrics metrics.Metrics,
	certificate tls.Certificate,
	caCert x509.Certificate,
	instanceId string,
) (*InterCpServer, error) {
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

	caPool := x509.NewCertPool()
	caPool.AddCert(&caCert)
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS12, // there to make gosec happy this is always set after.
	}
	var err error
	if tlsCfg.MinVersion, err = config_types.TLSVersion(config.TlsMinVersion); err != nil {
		return nil, err
	}
	if tlsCfg.MaxVersion, err = config_types.TLSVersion(config.TlsMaxVersion); err != nil {
		return nil, err
	}
	if tlsCfg.CipherSuites, err = config_types.TLSCiphers(config.TlsCipherSuites); err != nil {
		return nil, err
	}

	grpcOptions = append(grpcOptions, grpc.Creds(credentials.NewTLS(tlsCfg)))
	grpcOptions = append(grpcOptions, metrics.GRPCServerInterceptors()...)
	grpcOptions = append(grpcOptions, grpc.StatsHandler(otelgrpc.NewServerHandler(
		otelgrpc.WithFilter(filters.Not(filters.ServiceName(system_proto.InterCpPingService_ServiceDesc.ServiceName))),
	)))

	grpcServer := grpc.NewServer(grpcOptions...)

	reflection.Register(grpcServer)

	return &InterCpServer{
		config:     config,
		grpcServer: grpcServer,
		instanceId: instanceId,
	}, nil
}

func (d *InterCpServer) Ready() bool {
	return d.ready.Load()
}

func (d *InterCpServer) Start(stop <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Port))
	if err != nil {
		return err
	}
	log := log.WithValues(
		"instanceId",
		d.instanceId,
	)

	log.Info("starting", "interface", "0.0.0.0", "port", d.config.Port, "tls", true)
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		d.ready.Store(true)
		if err := d.grpcServer.Serve(lis); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error(err, "terminated with an error")
				errChan <- err
			}
		}
		log.Info("shutting down server")
	}()

	select {
	case <-stop:
		log.Info("stopping gracefully")
		d.ready.Store(false)
		d.grpcServer.GracefulStop()
		log.Info("stopped")
		return nil
	case err := <-errChan:
		d.ready.Store(false)
		return err
	}
}

func (d *InterCpServer) NeedLeaderElection() bool {
	return false
}

func (d *InterCpServer) GrpcServer() *grpc.Server {
	return d.grpcServer
}
