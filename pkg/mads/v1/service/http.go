package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/emicklei/go-restful"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cache_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"

	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	rest_error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
)

func (s *service) RegisterRoutes(ws *restful.WebService) {
	ws.Route(ws.POST("/v3/discovery:monitoringassignment").
		Doc("Exposes the observability/v1 API").
		Returns(200, "OK", v3.DiscoveryResponse{}).
		Returns(304, "Not Modified", nil).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		To(s.handleDiscovery))
}

func (s *service) handleDiscovery(req *restful.Request, res *restful.Response) {
	var discoveryReq v3.DiscoveryRequest
	if err := req.ReadEntity(&discoveryReq); err != nil {
		rest_errors.HandleError(res, err, "Could not decode DiscoveryRequest from body")
		return
	}

	if discoveryReq.TypeUrl != mads_v1.MonitoringAssignmentType {
		discoveryErr := rest_error_types.Error{
			Title:   "Can not handle MADS DiscoveryRequest",
			Details: fmt.Sprintf("Invalid MADS type: %s", discoveryReq.TypeUrl),
		}

		if err := res.WriteHeaderAndJson(400, discoveryErr, restful.MIME_JSON); err != nil {
			rest_errors.HandleError(res, err, "Could encode error")
			return
		}

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.FetchTimeout)
	defer cancel()

	discoveryRes, err := s.server.FetchMonitoringAssignments(ctx, &discoveryReq)
	if err != nil {
		if errors.Is(err, &cache_types.SkipFetchError{}) {
			// No update necessary, send 304
			s.log.V(1).Info("no update needed")
			res.WriteHeader(304)
		} else {
			rest_errors.HandleError(res, err, "Could not fetch MonitoringAssignments")
		}
		return
	}

	if err = res.WriteEntity(discoveryRes); err != nil {
		rest_errors.HandleError(res, err, "Could encode DiscoveryResponse")
		return
	}
}
