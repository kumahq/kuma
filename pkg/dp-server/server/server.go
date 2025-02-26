package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"
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
	GrpcKeepAliveTime        = 15 * time.Second
)

type Filter func(writer http.ResponseWriter, request *http.Request) bool

type DpServer struct {
	config         dp_server.DpServerConfig
	httpMux        *http.ServeMux
	grpcServer     *grpc.Server
	filter         Filter
	promMiddleware middleware.Middleware
	ready          atomic.Bool
}

var _ component.Component = &DpServer{}

func NewDpServer(config dp_server.DpServerConfig, metrics metrics.Metrics, filter Filter) *DpServer {
	grpcOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    GrpcKeepAliveTime,
			Timeout: GrpcKeepAliveTime,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             GrpcKeepAliveTime,
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
		filter:         filter,
		promMiddleware: promMiddleware,
	}
}

func (d *DpServer) Ready() bool {
	return d.ready.Load()
}

func (d *DpServer) Start(stop <-chan struct{}) error {
	var err error
	cert, err := tls.LoadX509KeyPair(d.config.TlsCertFile, d.config.TlsKeyFile)
	if err != nil {
		return errors.Wrap(err, "failed to load TLS certificate")
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12} // To make gosec happy
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
		ReadHeaderTimeout: d.config.ReadHeaderTimeout.Duration,
		Addr:              fmt.Sprintf(":%d", d.config.Port),
		Handler:           http.HandlerFunc(d.handle),
		TLSConfig:         tlsConfig,
		ErrorLog:          adapter.ToStd(log),
	}
	l, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		d.ready.Store(true)
		if err := server.ServeTLS(l, d.config.TlsCertFile, d.config.TlsKeyFile); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("terminated normally")
			} else {
				log.Error(err, "terminated with an error")
				errChan <- err
			}
		}
	}()
	log.Info("starting", "interface", "0.0.0.0", "port", d.config.Port, "tls", true)

	select {
	case <-stop:
		log.Info("stopping")
		d.ready.Store(false)
		return server.Shutdown(context.Background())
	case err := <-errChan:
		d.ready.Store(false)
		return err
	}
}

func (d *DpServer) NeedLeaderElection() bool {
	return false
}

func (d *DpServer) handle(writer http.ResponseWriter, request *http.Request) {
	if !d.filter(writer, request) {
		return
	}
	// add filter function that will be in runtime, and we will implement it in kong-mesh
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

func (d *DpServer) SetFilter(filter Filter) {
	d.filter = filter
}
