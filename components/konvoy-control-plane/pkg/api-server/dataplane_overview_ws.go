package api_server

import (
	"context"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"strings"
)

type overviewWs struct {
	resources core_resources.Resources
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
		if store.IsResourceNotFound(err) {
			writeError(response, 404, "")
		} else {
			core.Log.Error(err, "Could not retrieve a dataplane overview", "name", name)
			writeError(response, 500, "Could not retrieve a dataplane overview")
		}
	}

	res := rest.From.Resource(overview)
	if err := response.WriteAsJson(res); err != nil {
		core.Log.Error(err, "Could not write the response")
		writeError(response, 500, "Could not write the response")
	}
}

func (r *overviewWs) fetchOverview(ctx context.Context, name string, meshName string) (*mesh.DataplaneOverviewResource, error) {
	dataplane := mesh.DataplaneResource{}
	if err := r.resources.Get(ctx, &dataplane, store.GetByKey(namespace, name, meshName)); err != nil {
		return nil, err
	}

	insight := mesh.DataplaneInsightResource{}
	err := r.resources.Get(ctx, &insight, store.GetByKey(namespace, name, meshName))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have dataplane without insight
		return nil, err
	}

	return &mesh.DataplaneOverviewResource{
		Meta: dataplane.Meta,
		Spec: mesh_proto.DataplaneOverview{
			Dataplane:        dataplane.Spec,
			DataplaneInsight: insight.Spec,
		},
	}, nil
}

func (r *overviewWs) inspectDataplanes(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")
	overviews, err := r.fetchOverviews(request.Request.Context(), meshName)
	if err != nil {
		core.Log.Error(err, "Could not retrieve dataplane overviews")
		writeError(response, 500, "Could not list dataplane overviews")
		return
	}

	tags := parseTags(request.QueryParameters("tag"))
	overviews.RetainMatchingTags(tags)

	restList := rest.From.ResourceList(&overviews)
	if err := response.WriteAsJson(restList); err != nil {
		core.Log.Error(err, "Could not write DataplaneOverview as JSON")
		writeError(response, 500, "Could not list dataplane overviews")
	}
}

func (r *overviewWs) fetchOverviews(ctx context.Context, meshName string) (mesh.DataplaneOverviewResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := r.resources.List(ctx, &dataplanes, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneOverviewResourceList{}, err
	}

	insights := mesh.DataplaneInsightResourceList{}
	if err := r.resources.List(ctx, &insights, store.ListByMesh(meshName)); err != nil {
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
