package api_server

import (
	"regexp"

	"github.com/emicklei/go-restful/v3"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

var namePathParams = []string{"name", "mesh", "dataplane"}

var namePathPattern = regexp.MustCompile(core_model.MeshNamePattern)

const invalidNameCharsMessage = "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_', '.' symbols."

func addNameCharsetViolation(verr *validators.ValidationError, param, value string) {
	if value != "" && !namePathPattern.MatchString(value) {
		verr.AddViolation(param, invalidNameCharsMessage)
	}
}

func rejectInvalidNamePathParams(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	var verr validators.ValidationError
	for _, param := range namePathParams {
		addNameCharsetViolation(&verr, param, request.PathParameter(param))
	}
	if verr.HasViolations() {
		rest_errors.HandleError(request.Request.Context(), response, &verr, "Bad Request")
		return
	}
	chain.ProcessFilter(request, response)
}
