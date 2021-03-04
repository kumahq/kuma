package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_prometheus "github.com/kumahq/kuma/pkg/util/prometheus"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"io"
	"net/http"
)


var (
	httpServerLog = core.Log.WithName("mads-server").WithName("http")
)

// TODO: should this be merged with the grpc server ?
type httpServer struct {
	server  Server
	config  mads_config.MonitoringAssignmentServerConfig
	metrics core_metrics.Metrics
}

var (
	_ component.Component = &httpServer{}
)

func init() {
	// TODO: is this needed?
	// turn off escape & character so the link in "next" fields for resources is user friendly
	restful.NewEncoder = func(w io.Writer) *json.Encoder {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder
	}
	restful.MarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		var buf bytes.Buffer
		encoder := restful.NewEncoder(&buf)
		encoder.SetIndent(prefix, indent)
		if err := encoder.Encode(v); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}


func (s *httpServer) handleDiscovery(req *restful.Request, res *restful.Response) {
	discoveryReq := &v3.DiscoveryRequest{}
	if err := req.ReadEntity(discoveryReq); err != nil {
		rest_errors.HandleError(res, err, "Could not decode DiscoveryRequest from body")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.HttpTimeout)
	defer cancel()

	discoveryRes, err := s.server.FetchMonitoringAssignments(ctx, discoveryReq)
	if err != nil {
		rest_errors.HandleError(res, err, "Could not fetch MonitoringAssignments")
		return
	}

	if discoveryRes.VersionInfo == discoveryReq.VersionInfo {
		// No update necessary, send 304
		res.WriteHeader(304)
		return
	}

	if err = res.WriteEntity(discoveryRes); err != nil {
		rest_errors.HandleError(res, err, "Could encode DiscoveryResponse")
		return
	}
}

func (s *httpServer) Start(stop <-chan struct{}) error {
	container := restful.NewContainer()

	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: s.metrics,
			Prefix:   "mads_server",
		}),
	})
	container.Filter(util_prometheus.MetricsHandler("", promMiddleware))

	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/v3/discovery:monitoringassignment").
		Doc("Exposes the observability/v1 API").
		Returns(200, "OK", v3.DiscoveryResponse{}).
		Returns(304, "Not Modified", nil).
		To(s.handleDiscovery))

	container.Add(ws)

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: []string{}, // TODO serverConfig.CorsAllowedDomains, allow to be specified in component config
		Container:      container,
	}
	container.Filter(cors.Filter)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "0.0.0.0", s.config.HttpPort),
		Handler: container.ServeMux,
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		err := httpServer.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				httpServerLog.Info("shutting down server")
			default:
				httpServerLog.Error(err, "could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	httpServerLog.Info("starting", "interface", "0.0.0.0", "port", s.config.HttpPort)

	select {
	case <-stop:
		httpServerLog.Info("stopping gracefully")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (s *httpServer) NeedLeaderElection() bool {
	return false
}
