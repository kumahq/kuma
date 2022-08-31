package bootstrap

import (
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/bootstrap/types"
)

var log = core.Log.WithName("bootstrap")

type BootstrapHandler struct {
	Generator BootstrapGenerator
}

func (b *BootstrapHandler) Handle(resp http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
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

	host := req.Host
	if host == "" {
		host = req.URL.Host
	}

	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		// The host doesn't have a port so we just use it directly
		hostname = host
	}

	reqParams.Host = hostname
	logger := log.WithValues("params", reqParams)

	config, kumaDpBootstrap, err := b.Generator.Generate(req.Context(), reqParams)
	if err != nil {
		handleError(resp, err, logger)
		return
	}

	bootstrapBytes, err := proto.ToYAML(config)
	if err != nil {
		logger.Error(err, "Could not convert to json")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	var responseBytes []byte
	if req.Header.Get("accept") == "application/json" {
		resp.Header().Set("content-type", "application/json")
		response := createBootstrapResponse(bootstrapBytes, &kumaDpBootstrap)
		responseBytes, err = json.Marshal(response)
		if err != nil {
			logger.Error(err, "Could not convert to json")
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// backwards compatibility
		resp.Header().Set("content-type", "text/x-yaml")
		responseBytes = bootstrapBytes
	}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write(responseBytes)
	if err != nil {
		logger.Error(err, "Error while writing the response")
		return
	}
}

func handleError(resp http.ResponseWriter, err error, logger logr.Logger) {
	if err == DpTokenRequired || store.IsResourcePreconditionFailed(err) || validators.IsValidationError(err) {
		resp.WriteHeader(http.StatusUnprocessableEntity)
		_, err = resp.Write([]byte(err.Error()))
		if err != nil {
			logger.Error(err, "Error while writing the response")
		}
		return
	}
	if ISSANMismatchErr(err) || err == NotCA {
		resp.WriteHeader(http.StatusBadRequest)
		if _, err := resp.Write([]byte(err.Error())); err != nil {
			logger.Error(err, "Error while writing the response")
		}
		return
	}
	if store.IsResourceNotFound(err) {
		resp.WriteHeader(http.StatusNotFound)
		return
	}
	logger.Error(err, "Could not generate a bootstrap configuration")
	resp.WriteHeader(http.StatusInternalServerError)
}

func createBootstrapResponse(bootstrap []byte, config *KumaDpBootstrap) *types.BootstrapResponse {
	bootstrapConfig := types.BootstrapResponse{
		Bootstrap: bootstrap,
	}
	aggregate := []types.Aggregate{}
	for _, value := range config.AggregateMetricsConfig {
		aggregate = append(aggregate, types.Aggregate{
			Address: value.Address,
			Name:    value.Name,
			Port:    value.Port,
			Path:    value.Path,
		})
	}
	bootstrapConfig.KumaSidecarConfiguration = types.KumaSidecarConfiguration{
		Metrics: types.MetricsConfiguration{
			Aggregate: aggregate,
		},
		Networking: types.NetworkingConfiguration{
			IsUsingTransparentProxy: config.NetworkingConfig.IsUsingTransparentProxy,
		},
	}
	return &bootstrapConfig
}
