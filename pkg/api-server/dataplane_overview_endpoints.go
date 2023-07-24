package api_server

import (
	"context"

	"github.com/emicklei/go-restful/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
)

type dataplaneOverviewEndpoints struct {
	resManager     manager.ResourceManager
	resourceAccess access.ResourceAccess
}

func (r *dataplaneOverviewEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/dataplanes+insights/{name}").To(r.inspectDataplane).
		Doc("Inspect a dataplane").
		Param(ws.PathParameter("name", "Name of a dataplane").DataType("string")).
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *dataplaneOverviewEndpoints) addListEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/dataplanes+insights").To(r.inspectDataplanes).
		Doc("Inspect all dataplanes").
		Param(ws.PathParameter("mesh", "Name of a mesh").DataType("string")).
		Param(ws.QueryParameter("tag", "Tag to filter in key:value format").DataType("string")).
		Param(ws.QueryParameter("gateway", "Param to filter gateway dataplanes").DataType("boolean")).
		Param(ws.QueryParameter("ingress", "Param to filter ingress dataplanes").DataType("boolean")).
		Returns(200, "OK", nil))
}

func (r *dataplaneOverviewEndpoints) inspectDataplane(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	if err := r.resourceAccess.ValidateGet(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: name},
		mesh.NewDataplaneOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	overview, err := r.fetchOverview(request.Request.Context(), name, meshName)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a dataplane overview")
		return
	}

	res := rest.From.Resource(overview)
	if err := response.WriteAsJson(res); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a dataplane overview")
	}
}

func (r *dataplaneOverviewEndpoints) fetchOverview(ctx context.Context, name string, meshName string) (*mesh.DataplaneOverviewResource, error) {
	dataplane := mesh.NewDataplaneResource()
	if err := r.resManager.Get(ctx, dataplane, store.GetByKey(name, meshName)); err != nil {
		return nil, err
	}

	insight := mesh.NewDataplaneInsightResource()
	err := r.resManager.Get(ctx, insight, store.GetByKey(name, meshName))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have dataplane without insight
		return nil, err
	}

	return &mesh.DataplaneOverviewResource{
		Meta: dataplane.Meta,
		Spec: &mesh_proto.DataplaneOverview{
			Dataplane:        dataplane.Spec,
			DataplaneInsight: insight.Spec,
		},
	}, nil
}

func (r *dataplaneOverviewEndpoints) inspectDataplanes(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")

	if err := r.resourceAccess.ValidateList(
		request.Request.Context(),
		meshName,
		mesh.NewDataplaneOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	page, err := pagination(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve dataplane overviews")
		return
	}

	filter, err := genFilter(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve dataplane overviews")
		return
	}
	nameContains := request.QueryParameter("name")

	overviews, err := r.fetchOverviews(request.Request.Context(), page, meshName, nameContains, filter)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve dataplane overviews")
		return
	}

	// pagination is not supported yet so we need to override pagination total items after retaining dataplanes
	overviews.GetPagination().SetTotal(uint32(len(overviews.Items)))
	restList := rest.From.ResourceList(&overviews)
	restList.Next = nextLink(request, overviews.GetPagination().NextOffset)
	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not list dataplane overviews")
	}
}

func (r *dataplaneOverviewEndpoints) fetchOverviews(ctx context.Context, p page, meshName string, nameContains string, filter store.ListFilterFunc) (mesh.DataplaneOverviewResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := r.resManager.List(ctx, &dataplanes, store.ListByMesh(meshName), store.ListByPage(p.size, p.offset), store.ListByFilterFunc(filter), store.ListByNameContains(nameContains)); err != nil {
		return mesh.DataplaneOverviewResourceList{}, err
	}

	// we cannot paginate insights since there is no guarantee that the elements will be the same as dataplanes
	insights := mesh.DataplaneInsightResourceList{}
	if err := r.resManager.List(ctx, &insights, store.ListByMesh(meshName)); err != nil {
		return mesh.DataplaneOverviewResourceList{}, err
	}

	return mesh.NewDataplaneOverviews(dataplanes, insights), nil
}
