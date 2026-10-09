package api_server

import (
	"regexp"

	"github.com/emicklei/go-restful/v3"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

// namePathParams are path parameters that carry a resource or mesh name.
var namePathParams = []string{"name", "mesh", "dataplane"}

// namePathPattern matches the characters any resource or mesh name can contain.
// It is the charset shared by core_model.NamePattern and
// core_model.MeshNamePattern, applied route-agnostically because the same
// WebService serves Mesh routes whose legacy names may contain '_'. The
// stricter per-resource rules keep being enforced where bodies are validated.
var namePathPattern = regexp.MustCompile(core_model.MeshNamePattern)

const invalidNameCharsMessage = "invalid characters. Valid characters are numbers, lowercase latin letters and '-', '_', '.' symbols."

// addNameCharsetViolation appends a violation named param to verr when value
// holds a character no Kuma name can contain.
func addNameCharsetViolation(verr *validators.ValidationError, param, value string) {
	if value != "" && !namePathPattern.MatchString(value) {
		verr.AddViolation(param, invalidNameCharsMessage)
	}
}

// rejectInvalidNamePathParams answers 400 before the request reaches a handler
// when a path parameter carrying a name holds characters no name can have.
// Without it a NUL byte in a URL segment (e.g. .../ma-1%00) reaches the
// Postgres store, which fails the whole request with SQLSTATE 22021 and a 500
// plus an ERROR log, where the Kubernetes store answers 400.
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
