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

	hostname, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		log.Error(err, "Could not parse a request")
		resp.WriteHeader(http.StatusBadRequest)
		_, err := resp.Write([]byte("Invalid Host header"))
		if err != nil {
			log.Error(err, "Error while writing the response")
			return
		}
		return
	}
	reqParams.Host = hostname
	logger := log.WithValues("params", reqParams)

	config, err := b.Generator.Generate(req.Context(), reqParams)
	if err != nil {
		handleError(resp, err, logger)
		return
	}

	bytes, err = proto.ToYAML(config)
	if err != nil {
		logger.Error(err, "Could not convert to json")
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Header().Set("content-type", "text/x-yaml")
	// backwards compatibility
	resp.Header().Set(types.BootstrapVersionHeader, string(types.BootstrapV3))
	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write(bytes)
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
