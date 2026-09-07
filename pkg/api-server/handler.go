package api_server

import (
	"errors"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	rest_errors "github.com/kumahq/kuma/v3/pkg/core/rest/errors"
)

// handlerFunc is an HTTP handler that returns a response body (written as JSON
// with 200 OK) or an error (rendered by rest_errors.HandleError).
type handlerFunc func(request *restful.Request) (any, error)

// handle adapts a handlerFunc to restful.RouteFunction, centralizing error
// rendering and response writing so handlers don't repeat it per call site.
func handle(fn handlerFunc) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		body, err := fn(request)
		if err != nil {
			title := ""
			if terr, ok := errors.AsType[*titledError](err); ok {
				title = terr.title
				err = terr.err
			}
			rest_errors.HandleError(request.Request.Context(), response, err, title)
			return
		}
		status := http.StatusOK
		if sr, ok := body.(statusResponse); ok {
			status = sr.status
			body = sr.body
		}
		if raw, ok := body.(rawResponse); ok {
			response.AddHeader(restful.HEADER_ContentType, raw.contentType)
			if _, err := response.Write(raw.body); err != nil {
				log.Error(err, "Could not write the response")
			}
			return
		}
		if err := response.WriteHeaderAndJson(status, body, "application/json"); err != nil {
			log.Error(err, "Could not write the response")
		}
	}
}

// titledError carries the title used in the error response body, preserving
// the per-call-site titles handlers previously passed to HandleError.
type titledError struct {
	err   error
	title string
}

func (t *titledError) Error() string { return t.err.Error() }
func (t *titledError) Unwrap() error { return t.err }

// withTitle annotates err with the title used in the error response body.
func withTitle(err error, title string) error {
	if err == nil {
		return nil
	}
	return &titledError{err: err, title: title}
}

// statusResponse wraps a body with a non-200 status code.
type statusResponse struct {
	status int
	body   any
}

// created responds with 201 Created.
func created(body any) any {
	return statusResponse{status: http.StatusCreated, body: body}
}

// rawResponse is a response body written as-is with the given Content-Type,
// for endpoints that don't serve JSON (e.g. Envoy admin output).
type rawResponse struct {
	contentType string
	body        []byte
}
