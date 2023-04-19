package api_server

import (
	"context"

	"github.com/emicklei/go-restful/v3"

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

type zoneIngressOverviewEndpoints struct {
	resManager     manager.ResourceManager
	resourceAccess access.ResourceAccess
}

func (r *zoneIngressOverviewEndpoints) addFindEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/zoneingresses+insights/{name}").To(r.inspectZoneIngress).
		Doc("Inspect a zone ingress").
		Param(ws.PathParameter("name", "Name of a zone ingress").DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *zoneIngressOverviewEndpoints) addListEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/zoneingresses+insights").To(r.inspectZoneIngresses).
		Doc("Inspect all zone ingresses").
		Returns(200, "OK", nil))
}

func (r *zoneIngressOverviewEndpoints) inspectZoneIngress(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	if err := r.resourceAccess.ValidateGet(
		request.Request.Context(),
		model.ResourceKey{Name: name},
		mesh.NewZoneIngressOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	overview, err := r.fetchOverview(request.Request.Context(), name)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a zone ingress overview")
		return
	}

	res := rest.From.Resource(overview)
	if err := response.WriteAsJson(res); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a zone ingress overview")
	}
}

func (r *zoneIngressOverviewEndpoints) fetchOverview(ctx context.Context, name string) (*mesh.ZoneIngressOverviewResource, error) {
	zoneIngress := mesh.NewZoneIngressResource()
	if err := r.resManager.Get(ctx, zoneIngress, store.GetByKey(name, model.NoMesh)); err != nil {
		return nil, err
	}

	insight := mesh.NewZoneIngressInsightResource()
	err := r.resManager.Get(ctx, insight, store.GetByKey(name, model.NoMesh))
	if err != nil && !store.IsResourceNotFound(err) { // It's fine to have zone ingress without insight
		return nil, err
	}

	return &mesh.ZoneIngressOverviewResource{
		Meta: zoneIngress.Meta,
		Spec: &mesh_proto.ZoneIngressOverview{
			ZoneIngress:        zoneIngress.Spec,
			ZoneIngressInsight: insight.Spec,
		},
	}, nil
}

func (r *zoneIngressOverviewEndpoints) inspectZoneIngresses(request *restful.Request, response *restful.Response) {
	if err := r.resourceAccess.ValidateList(
		request.Request.Context(),
		"",
		mesh.NewZoneIngressOverviewResource().Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	page, err := pagination(request)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve dataplane overviews")
		return
	}

	overviews, err := r.fetchOverviews(request.Request.Context(), page)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve dataplane overviews")
		return
	}

	// pagination is not supported yet so we need to override pagination total items after retaining dataplanes
	overviews.GetPagination().SetTotal(uint32(len(overviews.Items)))
	restList := rest.From.ResourceList(&overviews)
	restList.Next = nextLink(request, overviews.GetPagination().NextOffset)
	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(response, err, "Could not list dataplane overviews")
	}
}

func (r *zoneIngressOverviewEndpoints) fetchOverviews(ctx context.Context, p page) (mesh.ZoneIngressOverviewResourceList, error) {
	zoneIngresses := mesh.ZoneIngressResourceList{}
	if err := r.resManager.List(ctx, &zoneIngresses, store.ListByPage(p.size, p.offset)); err != nil {
		return mesh.ZoneIngressOverviewResourceList{}, err
	}

	// we cannot paginate insights since there is no guarantee that the elements will be the same as zone ingresses
	insights := mesh.ZoneIngressInsightResourceList{}
	if err := r.resManager.List(ctx, &insights); err != nil {
		return mesh.ZoneIngressOverviewResourceList{}, err
	}

	return mesh.NewZoneIngressOverviews(zoneIngresses, insights), nil
}
