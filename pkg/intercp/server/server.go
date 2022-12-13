package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

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
}

var _ component.Component = &InterCpServer{}

func New(
	config intercp.InterCpServerConfig,
	metrics metrics.Metrics,
	certificate tls.Certificate,
	caCert x509.Certificate,
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
	grpcServer := grpc.NewServer(grpcOptions...)

	return &InterCpServer{
		config:     config,
		grpcServer: grpcServer,
	}, nil
}

func (d *InterCpServer) Start(stop <-chan struct{}) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", d.config.Port))
	if err != nil {
		return err
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := d.grpcServer.Serve(lis); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("terminated normally")
	}()
	log.Info("starting", "interface", "0.0.0.0", "port", d.config.Port, "tls", true)

	select {
	case <-stop:
		log.Info("stopping")
		d.grpcServer.GracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

func (d *InterCpServer) NeedLeaderElection() bool {
	return false
}

func (d *InterCpServer) GrpcServer() *grpc.Server {
	return d.grpcServer
}
