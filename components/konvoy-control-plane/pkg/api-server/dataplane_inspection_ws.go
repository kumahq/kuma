package api_server

import (
	"context"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"strings"
)

type inspectionWs struct {
	resourceStore store.ResourceStore
}

func (r *inspectionWs) AddToWs(ws *restful.WebService) {
	ws.Route(ws.GET("/{mesh}/dataplane-inspections/{name}").To(r.inspectDataplane).
		Doc("Inspect a dataplane").
		Param(ws.PathParameter("name", "Name of a dataplane").DataType("string")).
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("/{mesh}/dataplane-inspections").To(r.inspectDataplanes).
		Doc("Inspect all dataplanes").
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Param(ws.QueryParameter("tag", "Tag to filter in key:value format").DataType("string")).
		Returns(200, "OK", nil))
}

func (r *inspectionWs) inspectDataplane(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	inspection, err := r.fetchInspection(request.Request.Context(), name, meshName)
	if err != nil {
		if store.IsResourceNotFound(err) {
			writeError(response, 404, "")
		} else {
			core.Log.Error(err, "Could not retrieve a dataplane inspection", "name", name)
			writeError(response, 500, "Could not retrieve a dataplane inspection")
		}
	}

	res := rest.From.Resource(inspection)
	if err := response.WriteAsJson(res); err != nil {
		core.Log.Error(err, "Could not write the response")
		writeError(response, 500, "Could not write the response")
	}
}

func (r *inspectionWs) fetchInspection(ctx context.Context, name string, meshName string) (*mesh.DataplaneInspectionResource, error) {
	dataplane := mesh.DataplaneResource{}
	if err := r.resourceStore.Get(ctx, &dataplane, store.GetByKey(namespace, name, meshName)); err != nil {
		return nil, err
	}

	insight := mesh.DataplaneInsightResource{}
	err := r.resourceStore.Get(ctx, &insight, store.GetByKey(namespace, name, meshName))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have dataplane without insight
		return nil, err
	}

	return &mesh.DataplaneInspectionResource{
		Meta: dataplane.Meta,
		Spec: mesh_proto.DataplaneInspection{
			Dataplane:        dataplane.Spec,
			DataplaneInsight: insight.Spec,
		},
	}, nil
}

func (r *inspectionWs) inspectDataplanes(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")
	inspections, err := r.fetchInspections(request.Request.Context(), meshName)
	if err != nil {
		core.Log.Error(err, "Could not retrieve dataplane inspections")
		writeError(response, 500, "Could not list dataplane inspections")
		return
	}

	tags := parseTags(request.QueryParameters("tag"))
	inspections.RetainMatchingTags(tags)

	restList := rest.From.ResourceList(&inspections)
	if err := response.WriteAsJson(restList); err != nil {
		core.Log.Error(err, "Could not write DataplaneInspection as JSON")
		writeError(response, 500, "Could not list dataplane inspections")
	}
}

func (r *inspectionWs) fetchInspections(ctx context.Context, meshName string) (mesh.DataplaneInspectionResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := r.resourceStore.List(ctx, &dataplanes, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneInspectionResourceList{}, err
	}

	insights := mesh.DataplaneInsightResourceList{}
	if err := r.resourceStore.List(ctx, &insights, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneInspectionResourceList{}, err
	}

	return mesh.NewDataplaneInspections(dataplanes, insights), nil
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
