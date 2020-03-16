package api_server

import (
	"context"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	rest_errors "github.com/Kong/kuma/pkg/core/rest/errors"
)

type overviewWs struct {
	resManager manager.ResourceManager
}

func (r *overviewWs) AddToWs(ws *restful.WebService) {
	ws.Route(ws.GET("/{mesh}/dataplanes+insights/{name}").To(r.inspectDataplane).
		Doc("Inspect a dataplane").
		Param(ws.PathParameter("name", "Name of a dataplane").DataType("string")).
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("/{mesh}/dataplanes+insights").To(r.inspectDataplanes).
		Doc("Inspect all dataplanes").
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Param(ws.QueryParameter("tag", "Tag to filter in key:value format").DataType("string")).
		Returns(200, "OK", nil))
}

func (r *overviewWs) inspectDataplane(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	overview, err := r.fetchOverview(request.Request.Context(), name, meshName)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a dataplane overview")
		return
	}

	res := rest.From.Resource(overview)
	if err := response.WriteAsJson(res); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a dataplane overview")
	}
}

func (r *overviewWs) fetchOverview(ctx context.Context, name string, meshName string) (*mesh.DataplaneOverviewResource, error) {
	dataplane := mesh.DataplaneResource{}
	if err := r.resManager.Get(ctx, &dataplane, store.GetByKey(name, meshName)); err != nil {
		return nil, err
	}

	insight := mesh.DataplaneInsightResource{}
	err := r.resManager.Get(ctx, &insight, store.GetByKey(name, meshName))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have dataplane without insight
		return nil, err
	}

	return &mesh.DataplaneOverviewResource{
		Meta: dataplane.Meta,
		Spec: mesh_proto.DataplaneOverview{
			Dataplane:        proto.Clone(&dataplane.Spec).(*mesh_proto.Dataplane),
			DataplaneInsight: proto.Clone(&insight.Spec).(*mesh_proto.DataplaneInsight),
		},
	}, nil
}

func (r *overviewWs) inspectDataplanes(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")
	overviews, err := r.fetchOverviews(request.Request.Context(), meshName)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve dataplane overviews")
		return
	}

	tags := parseTags(request.QueryParameters("tag"))
	overviews.RetainMatchingTags(tags)

	restList := rest.From.ResourceList(&overviews)
	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(response, err, "Could not list dataplane overviews")
	}
}

func (r *overviewWs) fetchOverviews(ctx context.Context, meshName string) (mesh.DataplaneOverviewResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := r.resManager.List(ctx, &dataplanes, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneOverviewResourceList{}, err
	}

	insights := mesh.DataplaneInsightResourceList{}
	if err := r.resManager.List(ctx, &insights, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneOverviewResourceList{}, err
	}

	return mesh.NewDataplaneOverviews(dataplanes, insights), nil
}

// Tags should be passed in form of ?tag=service:mobile&tag=version:v1
func parseTags(queryParamValues []string) map[string]string {
	tags := make(map[string]string)
	for _, value := range queryParamValues {
		tagKv := strings.Split(value, ":")
		if len(tagKv) != 2 {
			// ignore invalid formatted tags
			continue
		}
		tags[tagKv[0]] = tagKv[1]
	}
	return tags
}
