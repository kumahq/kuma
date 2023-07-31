package api_server

import (
	"context"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
)

type Zone struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type Zones []Zone

func addZoneEndpoints(ws *restful.WebService, resManager manager.ResourceManager) {
	ws.Route(ws.GET("/status/zones").To(func(request *restful.Request, response *restful.Response) {
		zoneOverviews, err := fetchOverviews(resManager, request.Request.Context())
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a zone overview")
			return
		}

		if err := response.WriteAsJson(toZones(zoneOverviews)); err != nil {
			log.Error(err, "failed marshaling response")
		}
	}))
}

func fetchOverviews(resManager manager.ResourceManager, ctx context.Context) (system.ZoneOverviewResourceList, error) {
	zones := system.ZoneResourceList{}
	if err := resManager.List(ctx, &zones); err != nil {
		return system.ZoneOverviewResourceList{}, err
	}

	// we cannot paginate insights since there is no guarantee that the elements will be the same as dataplanes
	insights := system.ZoneInsightResourceList{}
	if err := resManager.List(ctx, &insights); err != nil {
		return system.ZoneOverviewResourceList{}, err
	}

	return system.NewZoneOverviews(zones, insights), nil
}

func toZones(rlist system.ZoneOverviewResourceList) Zones {
	var zones Zones
	for _, overview := range rlist.Items {
		zones = append(zones, Zone{
			Name:   overview.GetMeta().GetName(),
			Active: overview.Spec.GetZoneInsight().IsOnline() && overview.Spec.GetZone().IsEnabled(),
		})
	}
	return zones
}
