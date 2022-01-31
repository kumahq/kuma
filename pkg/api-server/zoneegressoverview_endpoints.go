package api_server

import (
	"context"

	"github.com/emicklei/go-restful"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
)

type zoneEgressOverviewEndpoints struct {
	resManager     manager.ResourceManager
	resourceAccess access.ResourceAccess
}

func (r *zoneEgressOverviewEndpoints) addFindEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/zoneegressoverviews/{name}").To(r.inspectZoneEgress).
		Doc("Inspect a zone egress").
		Param(ws.PathParameter("name", "Name of a zone egress").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *zoneEgressOverviewEndpoints) addListEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/zoneegressoverviews").To(r.inspectZoneEgresses).
		Doc("Inspect all zone egresses").
		Returns(200, "OK", nil))
}

func (r *zoneEgressOverviewEndpoints) inspectZoneEgress(
	request *restful.Request,
	response *restful.Response,
) {
	name := request.PathParameter("name")

	if err := r.resourceAccess.ValidateGet(
		model.ResourceKey{Name: name},
		mesh.NewZoneEgressOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	overview, err := r.fetchOverview(request.Request.Context(), name)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a zone egress overview")
		return
	}

	res := rest.From.Resource(overview)
	if err := response.WriteAsJson(res); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a zone egress overview")
	}
}

func (r *zoneEgressOverviewEndpoints) fetchOverview(
	ctx context.Context,
	name string,
) (*mesh.ZoneEgressOverviewResource, error) {
	zoneEgress := mesh.NewZoneEgressResource()
	if err := r.resManager.Get(ctx, zoneEgress, store.GetByKey(name, model.NoMesh)); err != nil {
		return nil, err
	}

	insight := mesh.NewZoneEgressInsightResource()
	err := r.resManager.Get(ctx, insight, store.GetByKey(name, model.NoMesh))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have zone egress without insight
		return nil, err
	}

	return &mesh.ZoneEgressOverviewResource{
		Meta: zoneEgress.Meta,
		Spec: &mesh_proto.ZoneEgressOverview{
			ZoneEgress:        zoneEgress.Spec,
			ZoneEgressInsight: insight.Spec,
		},
	}, nil
}

func (r *zoneEgressOverviewEndpoints) inspectZoneEgresses(
	request *restful.Request,
	response *restful.Response,
) {
	if err := r.resourceAccess.ValidateList(
		mesh.NewZoneEgressOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	page, err := pagination(request)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve zone egress overviews")
		return
	}

	overviews, err := r.fetchOverviews(request.Request.Context(), page)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve zone egress overviews")
		return
	}

	// pagination is not supported yet, so we need to override pagination total
	// items after retaining zone egresses
	overviews.GetPagination().SetTotal(uint32(len(overviews.Items)))
	restList := rest.From.ResourceList(&overviews)
	restList.Next = nextLink(request, overviews.GetPagination().NextOffset)
	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(response, err, "Could not list zone egress overviews")
	}
}

func (r *zoneEgressOverviewEndpoints) fetchOverviews(
	ctx context.Context,
	p page,
) (mesh.ZoneEgressOverviewResourceList, error) {
	zoneEgresses := mesh.ZoneEgressResourceList{}
	if err := r.resManager.List(ctx, &zoneEgresses, store.ListByPage(p.size, p.offset)); err != nil {
		return mesh.ZoneEgressOverviewResourceList{}, err
	}

	// we cannot paginate insights since there is no guarantee that the elements
	// will be the same as zone egresses
	insights := mesh.ZoneEgressInsightResourceList{}
	if err := r.resManager.List(ctx, &insights); err != nil {
		return mesh.ZoneEgressOverviewResourceList{}, err
	}

	return mesh.NewZoneEgressOverviews(zoneEgresses, insights), nil
}
