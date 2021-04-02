package server

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/mads"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1/service"
	mads_v1alpha1 "github.com/kumahq/kuma/pkg/mads/v1alpha1/service"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_prometheus "github.com/kumahq/kuma/pkg/util/prometheus"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net/http"
	"strings"
	"time"
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

// muxHandler implements http.Handler for gRPC by intercepting
// all gRPC traffic and passing through all other traffic
type muxHandler struct {
	grpcServer *grpc.Server
	// passthrough handles all non-gRPC traffic
	passthrough http.Handler
}

func (m *muxHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.ProtoMajor == 2 &&
		strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		m.grpcServer.ServeHTTP(writer, req)
		return
	}
	m.passthrough.ServeHTTP(writer, req)
}

//func (m *muxHandler)

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

	var webSvcs []*restful.WebService
	for _, service := range s.httpServices {
		ws := new(restful.WebService)
		ws.Filter(promFilterFunc)
		service.RegisterRoutes(ws)
		webSvcs = append(webSvcs, ws)
		container.Add(ws)
	}

	return container
}

func (s *muxServer) Start(stop <-chan struct{}) error {
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.config.Port),
		Handler: &muxHandler{
			grpcServer:  s.createGRPCServer(),
			passthrough: s.createHttpServicesHandler(),
		},
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		err := server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("shutting down server")
			default:
				log.Error(err, "could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	log.Info("starting", "interface", "0.0.0.0", "port", s.config.Port)

	select {
	case <-stop:
		log.Info("stopping gracefully")
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (s *muxServer) NeedLeaderElection() bool {
	return false
}

func SetupServer(rt core_runtime.Runtime) error {
	config := rt.Config().MonitoringAssignmentServer

	if !config.GrpcEnabled && !config.HttpEnabled {
		return nil
	}

	rm := rt.ReadOnlyResourceManager()

	var grpcServices []GrpcService
	var httpServices []HttpService

	addGRPCService := func(svc GrpcService) {
		if config.GrpcEnabled {
			grpcServices = append(grpcServices, svc)
		}
	}

	addHTTPService := func(svc HttpService) {
		if config.HttpEnabled {
			httpServices = append(httpServices, svc)
		}
	}

	if config.VersionIsEnabled(mads.API_V1_ALPHA1) {
		log.Info("MADS v1alpha1 is enabled")
		svc := mads_v1alpha1.NewService(config, rm, log.WithValues("apiVersion", mads.API_V1_ALPHA1))
		addGRPCService(svc)
	}

	if config.VersionIsEnabled(mads.API_V1) {
		log.Info("MADS v1 is enabled")
		svc := mads_v1.NewService(config, rm, log.WithValues("apiVersion", mads.API_V1))
		addGRPCService(svc)
		addHTTPService(svc)
	}

	return rt.Add(&muxServer{
		httpServices: httpServices,
		grpcServices: grpcServices,
		config:       config,
		metrics:      rt.Metrics(),
	})
}
