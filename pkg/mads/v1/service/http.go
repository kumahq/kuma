package service

import (
	"context"
	"net/http"
	"time"

	"github.com/emicklei/go-restful/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cache_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/golang/protobuf/jsonpb" // nolint: depguard

	"github.com/kumahq/kuma/pkg/core"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	rest_error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
)

const FetchMonitoringAssignmentsPath = "/v3/discovery:monitoringassignments"

func (s *service) RegisterRoutes(ws *restful.WebService) {
	ws.Route(ws.POST(FetchMonitoringAssignmentsPath).
		Doc("Exposes the observability/v1 API").
		Returns(http.StatusOK, "OK", v3.DiscoveryResponse{}).
		Returns(http.StatusNotModified, "Not Modified", nil).
		Returns(http.StatusBadRequest, "Invalid request", rest_error_types.Error{}).
		Returns(http.StatusInternalServerError, "Server error", rest_error_types.Error{}).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(s.handleDiscovery))
}

func (s *service) handleDiscovery(req *restful.Request, res *restful.Response) {
	discoveryReq := &v3.DiscoveryRequest{}
	if err := jsonpb.Unmarshal(req.Request.Body, discoveryReq); err != nil {
		writeBadRequestError(res, rest_error_types.Error{
			Title:   "Could not parse request body",
			Detail:  err.Error(),
			Details: err.Error(),
		})
		return
	}

	discoveryReq.TypeUrl = mads_v1.MonitoringAssignmentType

	timeout, err := s.parseFetchTimeout(req.QueryParameter("fetch-timeout"))
	if err != nil {
		writeBadRequestError(res, rest_error_types.Error{
			Title:   "Could not parse fetch-timeout",
			Detail:  err.Error(),
			Details: err.Error(),
		})
		return
	}

	// If the timeout is 0 or less, this indicates a synchronous fetch
	// of the latest snapshot, without polling
	var ctx context.Context
	if timeout <= 0 {
		ctx = req.Request.Context()
	} else {
		var cancelFunc context.CancelFunc
		ctx, cancelFunc = context.WithTimeout(req.Request.Context(), timeout)
		defer cancelFunc()
	}

	discoveryRes, err := s.server.FetchMonitoringAssignments(ctx, discoveryReq)
	if err != nil {
		if _, ok := err.(*cache_types.SkipFetchError); ok {
			// No update necessary, send 304 Not Modified
			s.log.V(1).Info("no update needed")
			res.WriteHeader(http.StatusNotModified)
		} else {
			rest_errors.HandleError(req.Request.Context(), res, err, "Could not fetch MonitoringAssignments")
		}
		return
	}

	marshaller := &jsonpb.Marshaler{OrigName: true}
	if err := marshaller.Marshal(res, discoveryRes); err != nil {
		rest_errors.HandleError(req.Request.Context(), res, err, "Could write DiscoveryResponse")
		return
	}
}

// writeBadRequestError writes the given error as a a 400 Bad Request, encoded as JSON.
// Any errors during that process are handled by errors.HandleError
func writeBadRequestError(res *restful.Response, err rest_error_types.Error) {
	if writeErr := res.WriteHeaderAndJson(http.StatusBadRequest, err, restful.MIME_JSON); writeErr != nil {
		core.Log.Error(writeErr, "Could not write the error response")
		return
	}
}

// parseFetchTimeout either parses the provided fetch-timeout string according to time.ParseDuration
// or, if no timeout value was supplied, returns the configured default
func (s *service) parseFetchTimeout(timeoutStr string) (time.Duration, error) {
	if timeoutStr == "" {
		return s.config.DefaultFetchTimeout.Duration, nil
	}

	return time.ParseDuration(timeoutStr)
}
