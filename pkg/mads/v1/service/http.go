package service

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful"
	v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	rest_error_types "github.com/kumahq/kuma/pkg/core/rest/errors/types"
	mads_v1 "github.com/kumahq/kuma/pkg/mads/v1"
)

func (s *service) RegisterRoutes(ws *restful.WebService) {
	ws.Route(ws.POST("/v3/discovery:monitoringassignment").
		Doc("Exposes the observability/v1 API").
		Returns(200, "OK", v3.DiscoveryResponse{}).
		Returns(304, "Not Modified", nil).
		To(s.handleDiscovery))
}

func (s *service) handleDiscovery(req *restful.Request, res *restful.Response) {
	var discoveryReq v3.DiscoveryRequest
	if err := req.ReadEntity(&discoveryReq); err != nil {
		rest_errors.HandleError(res, err, "Could not decode DiscoveryRequest from body")
		return
	}
	s.log.Info("handling request", "req", &discoveryReq)

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

	ctx, cancel := context.WithTimeout(context.Background(), s.config.HttpTimeout)
	defer cancel()

	// TODO: need to figure out how to bootstrap without the Fetch request thinking it's already in sync
	//if !s.snapshotCache.HasSnapshot(s.hasher.ID(discoveryReq.Node)) {
	//	if err := s.reconciler.Reconcile(ctx, discoveryReq.Node); err != nil {
	//		rest_errors.HandleError(res, err, "Could generate snapshot")
	//		return
	//	}
	//}

	discoveryRes, err := s.server.FetchMonitoringAssignments(ctx, &discoveryReq)
	if err != nil {
		switch err.Error() {
		case "skip fetch: version up to date": // hack hack hack
			// No update necessary, send 304
			res.WriteHeader(304)
		default:
			rest_errors.HandleError(res, err, "Could not fetch MonitoringAssignments")
		}
		return
	}

	if discoveryRes.VersionInfo == discoveryReq.VersionInfo {
		res.WriteHeader(304)
		return
	}

	if err = res.WriteEntity(discoveryRes); err != nil {
		rest_errors.HandleError(res, err, "Could encode DiscoveryResponse")
		return
	}
}
