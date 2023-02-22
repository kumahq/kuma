package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("dp-server")

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

type DpServer struct {
	config         dp_server.DpServerConfig
	httpMux        *http.ServeMux
	grpcServer     *grpc.Server
	promMiddleware middleware.Middleware
}

var _ component.Component = &DpServer{}

func NewDpServer(config dp_server.DpServerConfig, metrics metrics.Metrics) *DpServer {
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
	grpcOptions = append(grpcOptions, metrics.GRPCServerInterceptors()...)
	grpcServer := grpc.NewServer(grpcOptions...)

	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: metrics,
			Prefix:   "dp_server",
		}),
	})

	return &DpServer{
		config:         config,
		httpMux:        http.NewServeMux(),
		grpcServer:     grpcServer,
		promMiddleware: promMiddleware,
	}
}

func (d *DpServer) Start(stop <-chan struct{}) error {
	var err error
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12} // To make gosec pass this is always set after
	if tlsConfig.MinVersion, err = config_types.TLSVersion(d.config.TlsMinVersion); err != nil {
		return err
	}
	if tlsConfig.MaxVersion, err = config_types.TLSVersion(d.config.TlsMaxVersion); err != nil {
		return err
	}
	if tlsConfig.CipherSuites, err = config_types.TLSCiphers(d.config.TlsCipherSuites); err != nil {
		return err
	}
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Addr:              fmt.Sprintf(":%d", d.config.Port),
		Handler:           http.HandlerFunc(d.handle),
		TLSConfig:         tlsConfig,
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := server.ListenAndServeTLS(d.config.TlsCertFile, d.config.TlsKeyFile); err != nil {
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
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (d *DpServer) NeedLeaderElection() bool {
	return false
}

func (d *DpServer) handle(writer http.ResponseWriter, request *http.Request) {
	if request.ProtoMajor == 2 && strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
		d.grpcServer.ServeHTTP(writer, request)
	} else {
		// we only want to measure HTTP not GRPC requests because they can mess up metrics
		// for example ADS bi-directional stream counts as one really long request
		std.Handler("", d.promMiddleware, d.httpMux).ServeHTTP(writer, request)
	}
}

func (d *DpServer) HTTPMux() *http.ServeMux {
	return d.httpMux
}

func (d *DpServer) GrpcServer() *grpc.Server {
	return d.grpcServer
}
