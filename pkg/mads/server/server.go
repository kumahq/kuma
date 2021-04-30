package server

import (
<<<<<<< HEAD
=======
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	mads_config "github.com/kumahq/kuma/pkg/config/mads"
>>>>>>> 1a546686c... fix(kumactl) mads-server graceful shutdown (#1912)
	"github.com/kumahq/kuma/pkg/core"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	madsServerLog = core.Log.WithName("mads-server")
)

<<<<<<< HEAD
=======
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

>>>>>>> 1a546686c... fix(kumactl) mads-server graceful shutdown (#1912)
func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := NewXdsContext(madsServerLog)
	generator := NewSnapshotGenerator(rt)
	versioner := NewVersioner()
	reconciler := NewReconciler(hasher, cache, generator, versioner)
	syncTracker := NewSyncTracker(reconciler, rt.Config().MonitoringAssignmentServer.AssignmentRefreshInterval)
	callbacks := util_xds_v2.CallbacksChain{
		util_xds_v2.AdaptCallbacks(util_xds.LoggingCallbacks{Log: madsServerLog}),
		syncTracker,
	}
	srv := NewServer(cache, callbacks)
	return rt.Add(
		&grpcServer{
			server:  srv,
			config:  *rt.Config().MonitoringAssignmentServer,
			metrics: rt.Metrics(),
		},
	)
}
