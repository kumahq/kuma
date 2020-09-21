package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"

	"github.com/kumahq/kuma/pkg/config/xds/bootstrap"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var log = core.Log.WithName("bootstrap-server")

type BootstrapServer struct {
	Config    *bootstrap.BootstrapServerConfig
	Generator BootstrapGenerator
	Metrics   prometheus.Registerer
}

func (b *BootstrapServer) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &BootstrapServer{}

func (b *BootstrapServer) Start(stop <-chan struct{}) error {
	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: b.Metrics,
			Prefix:   "bootstrap_server",
		}),
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/bootstrap", b.handleBootstrapRequest)

	bootstrapServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", b.Config.Port),
		Handler: std.Handler("", promMiddleware, mux),
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := bootstrapServer.ListenAndServeTLS(b.Config.TlsCertFile, b.Config.TlsKeyFile); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("terminated normally")
	}()
	log.Info("starting", "interface", "0.0.0.0", "port", b.Config.Port, "tls", true)

	select {
	case <-stop:
		log.Info("stopping")
		return bootstrapServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (b *BootstrapServer) handleBootstrapRequest(resp http.ResponseWriter, req *http.Request) {
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error(err, "Could not read a request")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	reqParams := types.BootstrapRequest{}
	if err := json.Unmarshal(bytes, &reqParams); err != nil {
		log.Error(err, "Could not parse a request")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	config, err := b.Generator.Generate(req.Context(), reqParams)
	if err != nil {
		if store.IsResourceNotFound(err) {
			resp.WriteHeader(http.StatusNotFound)
			return
		}
		if store.IsResourcePreconditionFailed(err) {
			resp.WriteHeader(http.StatusUnprocessableEntity)
			_, err = resp.Write([]byte(err.Error()))
			if err != nil {
				log.WithValues("params", reqParams).Error(err, "Error while writing the response")
				return
			}
			return
		}
		if validators.IsValidationError(err) {
			resp.WriteHeader(http.StatusUnprocessableEntity)
			_, err = resp.Write([]byte(err.Error()))
			if err != nil {
				log.WithValues("params", reqParams).Error(err, "Error while writing the response")
				return
			}
			return
		}
		log.WithValues("params", reqParams).Error(err, "Could not generate a bootstrap configuration")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	bytes, err = proto.ToYAML(config)
	if err != nil {
		log.WithValues("params", reqParams).Error(err, "Could not convert to json")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("content-type", "text/x-yaml")
	_, err = resp.Write(bytes)
	if err != nil {
		log.WithValues("params", reqParams).Error(err, "Error while writing the response")
		return
	}
}
