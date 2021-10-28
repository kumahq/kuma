package api_server

import (
	"time"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/rbac"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
)

type globalInsightsEndpoints struct {
	resManager     manager.ResourceManager
	resourceAccess rbac.ResourceAccess
}

type globalInsightsStat struct {
	Total uint32 `json:"total"`
}

type globalInsightsResponse struct {
	Type          string             `json:"type"`
	CreationTime  time.Time          `json:"creationTime"`
	Meshes        globalInsightsStat `json:"meshes"`
	Zones         globalInsightsStat `json:"zones"`
	ZoneIngresses globalInsightsStat `json:"zoneIngresses"`
}

func newGlobalInsightsResponse(meshes, zones, zoneIngresses globalInsightsStat) *globalInsightsResponse {
	return &globalInsightsResponse{
		Type:          "GlobalInsights",
		CreationTime:  core.Now(),
		Meshes:        meshes,
		Zones:         zones,
		ZoneIngresses: zoneIngresses,
	}
}

func (r *globalInsightsEndpoints) addEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/global-insights").To(r.inspectGlobalResources).
		Doc("Inspect all global resources").
		Returns(200, "OK", nil))
}

func (r *globalInsightsEndpoints) inspectGlobalResources(request *restful.Request, response *restful.Response) {
	meshes := &mesh.MeshResourceList{}
	if err := r.resManager.List(request.Request.Context(), meshes); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve global insights")
		return
	}

	zones := &system.ZoneResourceList{}
	if err := r.resManager.List(request.Request.Context(), zones); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve global insights")
		return
	}

	zoneIngresses := &mesh.ZoneIngressResourceList{}
	if err := r.resManager.List(request.Request.Context(), zoneIngresses); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve global insights")
		return
	}

	insights := newGlobalInsightsResponse(
		globalInsightsStat{Total: uint32(len(meshes.Items))},
		globalInsightsStat{Total: uint32(len(zones.Items))},
		globalInsightsStat{Total: uint32(len(zoneIngresses.Items))},
	)

	if err := response.WriteAsJson(insights); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve global insights")
	}
}
