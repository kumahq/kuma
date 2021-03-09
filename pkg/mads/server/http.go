package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful"
	mads_config "github.com/kumahq/kuma/pkg/config/mads"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_prometheus "github.com/kumahq/kuma/pkg/util/prometheus"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
)


var (
	httpServerLog = core.Log.WithName("mads-server").WithName("http")
)

type httpServer struct {
	services []HttpService
	config  *mads_config.MonitoringAssignmentServerConfig
	metrics core_metrics.Metrics
}

type HttpService interface {
	RegisterRoutes(ws *restful.WebService)
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

	for _, service := range s.services {
		service.RegisterRoutes(ws)
	}

	container.Add(ws)

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: []string{}, // TODO serverConfig.CorsAllowedDomains, allow to be specified in component config
		Container:      container,
	}
	container.Filter(cors.Filter)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "0.0.0.0", s.config.HttpPort),
		Handler: container.ServeMux,
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		err := server.ListenAndServe()
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
		return server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (s *httpServer) NeedLeaderElection() bool {
	return false
}
