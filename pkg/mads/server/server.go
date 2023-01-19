package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/mads"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1/service"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_prometheus "github.com/kumahq/kuma/pkg/util/prometheus"
)

const (
	grpcMaxConcurrentStreams = 1000000
	grpcKeepAliveTime        = 15 * time.Second
)

var (
	log = core.Log.WithName("mads-server")
)

// muxServer is a runtime component.Component that
// multiplexes all MADs resources over HTTP and gRPC
type muxServer struct {
	httpServices []HttpService
	grpcServices []GrpcService
	config       *mads_config.MonitoringAssignmentServerConfig
	metrics      core_metrics.Metrics
}

type HttpService interface {
	RegisterRoutes(ws *restful.WebService)
}

type GrpcService interface {
	RegisterWithGrpcServer(server *grpc.Server)
}

var (
	_ component.Component = &muxServer{}
)

func (s *muxServer) createGRPCServer() *grpc.Server {
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

	for _, service := range s.grpcServices {
		service.RegisterWithGrpcServer(server)
	}

	s.metrics.RegisterGRPC(server)

	return server
}

func (s *muxServer) createHttpServicesHandler() http.Handler {
	container := restful.NewContainer()
	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: s.metrics,
			Prefix:   "mads_server",
		}),
	})
	promFilterFunc := util_prometheus.MetricsHandler("", promMiddleware)

	for _, service := range s.httpServices {
		ws := new(restful.WebService)
		ws.Filter(promFilterFunc)
		service.RegisterRoutes(ws)
		container.Add(ws)
	}

	return container
}

func (s *muxServer) Start(stop <-chan struct{}) error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Error(err, "unable to close the listener")
		}
	}()

	m := cmux.New(l)

	grpcL := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
	)
	httpL := m.Match(cmux.HTTP1Fast())

	grpcS := s.createGRPCServer()
	errChanGrpc := make(chan error)
	go func() {
		defer close(errChanGrpc)
		if err := grpcS.Serve(grpcL); err != nil {
			switch err {
			case cmux.ErrServerClosed:
				log.Info("shutting down a GRPC Server")
			default:
				log.Error(err, "could not start an GRPC Server")
				errChanGrpc <- err
			}
		}
	}()

	httpS := &http.Server{
		Handler: s.createHttpServicesHandler(),
	}
	errChanHttp := make(chan error)
	go func() {
		defer close(errChanHttp)
		if err := httpS.Serve(httpL); err != nil {
			switch err {
			case cmux.ErrServerClosed:
				log.Info("shutting down an HTTP Server")
			default:
				log.Error(err, "could not start an HTTP Server")
				errChanHttp <- err
			}
		}
	}()

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		err := m.Serve()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Info("shutting down a mux server")
			} else {
				log.Error(err, "could not start a mux Server")
				errChan <- err
			}
		}
	}()

	log.Info("starting", "interface", "0.0.0.0", "port", s.config.Port)

	defer m.Close()

	select {
	case <-stop:
		log.Info("stopping gracefully")
		return nil
	case err := <-errChanGrpc:
		return err
	case err := <-errChanHttp:
		return err
	case err := <-errChan:
		return err
	}
}

func (s *muxServer) NeedLeaderElection() bool {
	return false
}

func SetupServer(rt core_runtime.Runtime) error {
	config := rt.Config().MonitoringAssignmentServer

	rm := rt.ReadOnlyResourceManager()

	var grpcServices []GrpcService
	var httpServices []HttpService

	if config.VersionIsEnabled(mads.API_V1) {
		log.Info("MADS v1 is enabled")
		svc := mads_v1.NewService(config, rm, log.WithValues("apiVersion", mads.API_V1))
		grpcServices = append(grpcServices, svc)
		httpServices = append(httpServices, svc)
	}

	return rt.Add(&muxServer{
		httpServices: httpServices,
		grpcServices: grpcServices,
		config:       config,
		metrics:      rt.Metrics(),
	})
}
