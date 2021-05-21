package service

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/jsonpb"

	"github.com/emicklei/go-restful"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cache_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

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
	body, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		readErr := rest_error_types.Error{
			Title:   "Can not read request body",
			Details: err.Error(),
		}

		if err := res.WriteHeaderAndJson(http.StatusBadRequest, readErr, restful.MIME_JSON); err != nil {
			rest_errors.HandleError(res, err, "Could encode error")
			return
		}
	}

	discoveryReq := &v3.DiscoveryRequest{}
	err = jsonpb.UnmarshalString(string(body), discoveryReq)
	if err != nil {
		readErr := rest_error_types.Error{
			Title:   "Can not decode request body",
			Details: err.Error(),
		}

		if err := res.WriteHeaderAndJson(http.StatusBadRequest, readErr, restful.MIME_JSON); err != nil {
			rest_errors.HandleError(res, err, "Could encode error")
			return
		}
	}

	discoveryReq.TypeUrl = mads_v1.MonitoringAssignmentType

	ctx, cancel := context.WithTimeout(context.Background(), s.config.FetchTimeout)
	defer cancel()

	discoveryRes, err := s.server.FetchMonitoringAssignments(ctx, discoveryReq)
	if err != nil {
		if _, ok := err.(*cache_types.SkipFetchError); ok {
			// No update necessary, send 304 Not Modified
			s.log.V(1).Info("no update needed")
			res.WriteHeader(http.StatusNotModified)
		} else {
			rest_errors.HandleError(res, err, "Could not fetch MonitoringAssignments")
		}
		return
	}

	marshaller := &jsonpb.Marshaler{OrigName: true}
	resStr, err := marshaller.MarshalToString(discoveryRes)
	if err != nil {
		rest_errors.HandleError(res, err, "Could encode DiscoveryResponse")
		return
	}

	if _, err = res.Write([]byte(resStr)); err != nil {
		rest_errors.HandleError(res, err, "Could write DiscoveryResponse")
		return
	}
}
