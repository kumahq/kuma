package bootstrap

import (
	"encoding/json"
	"io/ioutil"
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

	resp.WriteHeader(http.StatusOK)
	resp.Header().Set("content-type", "text/x-yaml")
	_, err = resp.Write(bytes)
	if err != nil {
		logger.Error(err, "Error while writing the response")
		return
	}
}

func handleError(resp http.ResponseWriter, err error, logger logr.Logger) {
	if err == DpTokenRequired {
		resp.WriteHeader(http.StatusUnprocessableEntity)
		_, err = resp.Write([]byte(err.Error()))
		if err != nil {
			logger.Error(err, "Error while writing the response")
		}
		return
	}
	if store.IsResourceNotFound(err) {
		resp.WriteHeader(http.StatusNotFound)
		return
	}
	if store.IsResourcePreconditionFailed(err) {
		resp.WriteHeader(http.StatusUnprocessableEntity)
		_, err = resp.Write([]byte(err.Error()))
		if err != nil {
			logger.Error(err, "Error while writing the response")
		}
		return
	}
	if validators.IsValidationError(err) {
		resp.WriteHeader(http.StatusUnprocessableEntity)
		_, err = resp.Write([]byte(err.Error()))
		if err != nil {
			logger.Error(err, "Error while writing the response")
		}
		return
	}
	logger.Error(err, "Could not generate a bootstrap configuration")
	resp.WriteHeader(http.StatusInternalServerError)
}
