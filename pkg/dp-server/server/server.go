package server

import (
	"context"
	"crypto/tls"
	std_errors "errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	dp_server "github.com/kumahq/kuma/v2/pkg/config/dp-server"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	"github.com/kumahq/kuma/v2/pkg/metrics"
)

var log = core.Log.WithName("dp-server")

const (
	grpcMaxConcurrentStreams = 1000000
	GrpcKeepAliveTime        = 15 * time.Second
)

const (
	forceStopReasonTimeout = "shutdown_timeout"
	forceStopReasonError   = "http_error"
)

type Filter func(writer http.ResponseWriter, request *http.Request) bool

type DpServer struct {
	config           dp_server.DpServerConfig
	httpMux          *http.ServeMux
	grpcServer       *grpc.Server
	filter           Filter
	promMiddleware   middleware.Middleware
	ready            atomic.Bool
	started          atomic.Bool
	forceStopCount   *prometheus.CounterVec
	shutdownDuration prometheus.Histogram
	done             chan struct{}
}

var _ component.GracefulComponent = &DpServer{}

type httpShutdowner interface {
	Shutdown(context.Context) error
}

type grpcStopper interface {
	Stop()
}

func NewDpServer(config dp_server.DpServerConfig, metrics metrics.Metrics, filter Filter) (*DpServer, error) {
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
		// Recover from panics in handlers so a single failing call is aborted
		// instead of bringing down the whole server.
		grpc.ChainUnaryInterceptor(recoveryUnaryInterceptor),
		grpc.ChainStreamInterceptor(recoveryStreamInterceptor),
	}
	grpcOptions = append(grpcOptions, metrics.GRPCServerInterceptors()...)
	grpcServer := grpc.NewServer(grpcOptions...)

	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: metrics,
			Prefix:   "dp_server",
		}),
	})

	forceStopCount := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dp_server_force_stop_total",
		Help: "Number of times dp-server force-stopped the gRPC server during shutdown, labeled by reason.",
	}, []string{"reason"})
	shutdownDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "dp_server_shutdown_duration_seconds",
		Help: "Wall-clock time spent in dp-server graceful shutdown.",
		// Buckets tuned around the 10s default deadline: sub-second
		// resolution for the common clean-drain case, dense between
		// 5s and 10s to separate slow drains from the force-stop,
		// and a long tail up to 60s for operators who bump
		// terminationGracePeriodSeconds.
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 7.5, 9, 10, 15, 30, 60},
	})
	if err := metrics.BulkRegister(forceStopCount, shutdownDuration); err != nil {
		return nil, err
	}

	return &DpServer{
		config:           config,
		httpMux:          http.NewServeMux(),
		grpcServer:       grpcServer,
		filter:           filter,
		promMiddleware:   promMiddleware,
		forceStopCount:   forceStopCount,
		shutdownDuration: shutdownDuration,
		done:             make(chan struct{}),
	}, nil
}

// shutdownDpServer drains the HTTP server within ctx's deadline, then
// force-stops the gRPC server. xDS handlers are long-lived and only
// return when the application context is cancelled, so the HTTP
// shutdown unblocks once that propagation happens. The trailing
// force-stop is a backstop that aborts any stream whose handler
// ignored the cancellation. gRPC's graceful-stop path is unsafe when
// the server is mounted as an HTTP/2 handler: its draining call on
// that transport panics. Returns the error from the HTTP shutdown
// (nil on clean drain, ctx.Err() on grace expiry).
func shutdownDpServer(ctx context.Context, httpSrv httpShutdowner, grpcSrv grpcStopper) error {
	err := httpSrv.Shutdown(ctx)
	grpcSrv.Stop()
	return err
}

func (d *DpServer) Ready() bool {
	return d.ready.Load()
}

func (d *DpServer) Start(stop <-chan struct{}) error {
	if !d.started.CompareAndSwap(false, true) {
		return errors.New("dp-server already started")
	}
	defer close(d.done)

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

	// Buffered so a late ServeTLS error (e.g., after force-stop) does
	// not leak the goroutine when the outer select has already picked
	// <-stop.
	errChan := make(chan error, 1)

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
		timeout := d.config.GracefulShutdownTimeout.Duration
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		start := time.Now()
		err := shutdownDpServer(shutdownCtx, server, d.grpcServer)
		d.shutdownDuration.Observe(time.Since(start).Seconds())
		switch {
		case err == nil:
			log.Info("terminated normally")
			return nil
		case std_errors.Is(err, context.DeadlineExceeded):
			// Expected outcome when xDS streams can't drain in time;
			// the bounded deadline exists precisely so the pod can
			// exit cleanly. Keep this at INFO to avoid false-positive
			// alerts on ERROR-level kuma-cp lines.
			d.forceStopCount.WithLabelValues(forceStopReasonTimeout).Inc()
			log.Info("dp-server force-stopped", "reason", forceStopReasonTimeout, "timeout", timeout)
			return nil
		default:
			d.forceStopCount.WithLabelValues(forceStopReasonError).Inc()
			log.Error(err, "dp-server force-stopped", "reason", forceStopReasonError)
			return err
		}
	case err := <-errChan:
		d.ready.Store(false)
		return err
	}
}

func (d *DpServer) NeedLeaderElection() bool {
	return false
}

func (d *DpServer) WaitForDone() {
	if !d.started.Load() {
		return
	}
	<-d.done
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
