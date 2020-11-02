package api_server

import (
	"context"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
)

type Zone struct {
	Name    string `json:"name"`
	Address string `json:"url"`
	Active  bool   `json:"active"`
}

type Zones []Zone

func zonesWs(resManager manager.ResourceManager) *restful.WebService {
	ws := new(restful.WebService).Path("/status/zones")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		zoneOverviews, err := fetchOverviews(resManager, request.Request.Context())
		if err != nil {
			rest_errors.HandleError(response, err, "Could not retrieve a zone overview")
			return
		}

		if err := response.WriteAsJson(toZones(zoneOverviews)); err != nil {
			log.Error(err, "failed marshaling response")
		}
	}))
}

func fetchOverviews(resManager manager.ResourceManager, ctx context.Context) (system.ZoneOverviewResourceList, error) {
	zones := system.ZoneResourceList{}
	if err := resManager.List(ctx, &zones, store.ListByMesh(core_model.DefaultMesh)); err != nil {
		return system.ZoneOverviewResourceList{}, err
	}

	// we cannot paginate insights since there is no guarantee that the elements will be the same as dataplanes
	insights := system.ZoneInsightResourceList{}
	if err := resManager.List(ctx, &insights, store.ListByMesh(core_model.DefaultMesh)); err != nil {
		return system.ZoneOverviewResourceList{}, err
	}

	return system.NewZoneOverviews(zones, insights), nil
}

func toZones(rlist system.ZoneOverviewResourceList) Zones {
	var zones Zones
	for _, overview := range rlist.Items {
		zones = append(zones, Zone{
			Name:    overview.GetMeta().GetName(),
			Address: "", // todo remove?
			Active:  overview.Spec.ZoneInsight.IsOnline(),
		})
	}
	return zones
}
