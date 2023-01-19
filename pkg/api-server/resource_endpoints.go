package api_server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
)

const (
	k8sReadOnlyMessage = "On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
		" As a best practice, you should always be using 'kubectl apply' instead." +
		" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"
	globalReadOnlyMessage = "On global control plane you can not modify dataplane resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify them on the zone control plane.\n"
	zoneReadOnlyMessage = "On zone control plane you can only modify dataplane resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify the rest of the resource on the global control plane.\n"
)

type resourceEndpoints struct {
	mode           config_core.CpMode
	resManager     manager.ResourceManager
	descriptor     model.ResourceTypeDescriptor
	resourceAccess access.ResourceAccess
}

func (r *resourceEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.descriptor.WsPath)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *resourceEndpoints) findResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	if err := r.resourceAccess.ValidateGet(
		model.ResourceKey{Mesh: meshName, Name: name},
		r.descriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	resource := r.descriptor.NewObject()
	err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName))
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a resource")
	} else {
		res := rest.From.Resource(resource)
		if err := response.WriteAsJson(res); err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}

func (r *resourceEndpoints) addListEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix).To(r.listResources).
		Doc(fmt.Sprintf("List of %s", r.descriptor.Name)).
		Param(ws.PathParameter("size", "size of page").DataType("int")).
		Param(ws.PathParameter("offset", "offset of page to list").DataType("string")).
		Returns(200, "OK", nil))
}

func (r *resourceEndpoints) listResources(request *restful.Request, response *restful.Response) {
	meshName := r.meshFromRequest(request)

	if err := r.resourceAccess.ValidateList(
		r.descriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	page, err := pagination(request)
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve resources")
		return
	}

	list := r.descriptor.NewList()
	if err := r.resManager.List(request.Request.Context(), list, store.ListByMesh(meshName), store.ListByPage(page.size, page.offset)); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve resources")
	} else {
		restList := rest.From.ResourceList(list)
		restList.Next = nextLink(request, list.GetPagination().NextOffset)
		if err := response.WriteAsJson(restList); err != nil {
			rest_errors.HandleError(response, err, "Could not list resources")
		}
	}
}

func (r *resourceEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.createOrUpdateResourceReadOnly).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.descriptor.WsPath)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.descriptor.WsPath)).DataType("string")).
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))
	}
}

func (r *resourceEndpoints) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	resourceRes := rest.Resource{
		Spec: r.descriptor.NewObject().GetSpec(),
	}

	if err := request.ReadEntity(&resourceRes); err != nil {
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	if err := r.validateResourceRequest(request, &resourceRes); err != nil {
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	resource := r.descriptor.NewObject()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
		if store.IsResourceNotFound(err) {
			r.createResource(request.Request.Context(), name, meshName, resourceRes.Spec, response)
		} else {
			rest_errors.HandleError(response, err, "Could not find a resource")
		}
	} else {
		r.updateResource(request.Request.Context(), resource, resourceRes, response)
	}
}

func (r *resourceEndpoints) createResource(ctx context.Context, name string, meshName string, spec model.ResourceSpec, response *restful.Response) {
	if err := r.resourceAccess.ValidateCreate(
		model.ResourceKey{Mesh: meshName, Name: name},
		spec,
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	res := r.descriptor.NewObject()
	_ = res.SetSpec(spec)
	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceEndpoints) updateResource(ctx context.Context, res model.Resource, restRes rest.Resource, response *restful.Response) {
	_ = res.SetSpec(restRes.Spec)

	if err := r.resourceAccess.ValidateUpdate(
		model.ResourceKey{Mesh: res.GetMeta().GetMesh(), Name: res.GetMeta().GetName()},
		res.GetSpec(),
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	if err := r.resManager.Update(ctx, res); err != nil {
		rest_errors.HandleError(response, err, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceEndpoints) createOrUpdateResourceReadOnly(request *restful.Request, response *restful.Response) {
	err := response.WriteErrorString(http.StatusMethodNotAllowed, r.readOnlyMessage())
	if err != nil {
		core.Log.Error(err, "Could not write the response")
	}
}

func (r *resourceEndpoints) addDeleteEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.deleteResourceReadOnly).
			Doc("Not allowed in read-only mode.").
			Returns(http.StatusMethodNotAllowed, "Not allowed in read-only mode.", restful.ServiceError{}))
	} else {
		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.descriptor.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.descriptor.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}
}

func (r *resourceEndpoints) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)
	resource := r.descriptor.NewObject()

	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not delete a resource")
		return
	}

	if err := r.resourceAccess.ValidateDelete(
		model.ResourceKey{Mesh: meshName, Name: name},
		resource.GetSpec(),
		resource.Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(response, err, "Access Denied")
		return
	}

	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not delete a resource")
	}
}

func (r *resourceEndpoints) deleteResourceReadOnly(request *restful.Request, response *restful.Response) {
	err := response.WriteErrorString(http.StatusMethodNotAllowed, r.readOnlyMessage())
	if err != nil {
		core.Log.Error(err, "Could not write the response")
	}
}

func (r *resourceEndpoints) validateResourceRequest(request *restful.Request, resource *rest.Resource) error {
	var err validators.ValidationError
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	if name != resource.Meta.Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if string(r.descriptor.Name) != resource.Meta.Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if r.descriptor.Scope == model.ScopeMesh && meshName != resource.Meta.Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}
	err.AddError("", mesh.ValidateMeta(name, meshName, r.descriptor.Scope))
	return err.OrNil()
}

func (r *resourceEndpoints) meshFromRequest(request *restful.Request) string {
	if r.descriptor.Scope == model.ScopeMesh {
		return request.PathParameter("mesh")
	}
	return ""
}

func (r *resourceEndpoints) readOnlyMessage() string {
	switch r.mode {
	case config_core.Global:
		return globalReadOnlyMessage
	case config_core.Zone:
		return zoneReadOnlyMessage
	default:
		return k8sReadOnlyMessage
	}
}
