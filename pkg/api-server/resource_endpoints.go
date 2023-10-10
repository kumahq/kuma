package api_server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/emicklei/go-restful/v3"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	rest_v1alpha1 "github.com/kumahq/kuma/pkg/core/resources/model/rest/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

const (
	k8sReadOnlyMessage = "On Kubernetes you cannot change the state of Kuma resources with 'kumactl apply' or via the HTTP API." +
		" As a best practice, you should always be using 'kubectl apply' instead." +
		" You can still use 'kumactl' or the HTTP API to make read-only operations. On Universal this limitation does not apply.\n"
	globalReadOnlyMessage = "On global control plane you can not modify dataplane resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify them on the zone control plane.\n"
	zoneReadOnlyMessage = "On zone control plane you can only modify zone resources with 'kumactl apply' or via the HTTP API." +
		" You can still use 'kumactl' or the HTTP API to modify the rest of the resource on the global control plane.\n"
)

type resourceEndpoints struct {
	mode           config_core.CpMode
	zoneName       string
	resManager     manager.ResourceManager
	descriptor     model.ResourceTypeDescriptor
	resourceAccess access.ResourceAccess
	k8sMapper      k8s.ResourceMapperFunc
	filter         func(request *restful.Request) (store.ListFilterFunc, error)
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
		request.Request.Context(),
		model.ResourceKey{Mesh: meshName, Name: name},
		r.descriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	resource := r.descriptor.NewObject()
	err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName))
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
	} else {
		var res interface{}
		switch request.QueryParameter("format") {
		case "k8s", "kubernetes":
			res, err = r.k8sMapper(resource, request.QueryParameter("namespace"))
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "k8s mapping failed")
				return
			}
		case "universal", "":
			res = rest.From.Resource(resource)
		default:
			err := validators.MakeFieldMustBeOneOfErr("format", "k8s", "kubernetes", "universal")
			rest_errors.HandleError(request.Request.Context(), response, err.OrNil(), "invalid format")
		}
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
		request.Request.Context(),
		meshName,
		r.descriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	page, err := pagination(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
		return
	}
	filter, err := r.filter(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
		return
	}

	list := r.descriptor.NewList()
	if err := r.resManager.List(request.Request.Context(), list, store.ListByMesh(meshName), store.ListByFilterFunc(filter), store.ListByPage(page.size, page.offset)); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
	} else {
		restList := rest.From.ResourceList(list)
		restList.Next = nextLink(request, list.GetPagination().NextOffset)
		if err := response.WriteAsJson(restList); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not list resources")
		}
	}
}

func (r *resourceEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(func(request *restful.Request, response *restful.Response) {
			rest_errors.HandleError(request.Request.Context(), response, &rest_errors.MethodNotAllowed{}, r.readOnlyMessage())
		}).
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

	bodyBytes, err := io.ReadAll(request.Request.Body)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	resourceRest, err := rest.JSON.Unmarshal(bodyBytes, r.descriptor)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	create := false
	resource := r.descriptor.NewObject()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(name, meshName)); err != nil && store.IsResourceNotFound(err) {
		create = true
	} else if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not find a resource")
	}

	if err := r.validateResourceRequest(request, resourceRest.GetMeta(), create); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not process a resource")
		return
	}

	if create {
		r.createResource(request.Request.Context(), name, meshName, resourceRest.GetSpec(), response)
	} else {
		r.updateResource(request.Request.Context(), resource, resourceRest.GetSpec(), response)
	}
}

func (r *resourceEndpoints) createResource(ctx context.Context, name string, meshName string, spec model.ResourceSpec, response *restful.Response) {
	if err := r.resourceAccess.ValidateCreate(
		ctx,
		model.ResourceKey{Mesh: meshName, Name: name},
		spec,
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	res := r.descriptor.NewObject()
	_ = res.SetSpec(spec)
	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName)); err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceEndpoints) updateResource(
	ctx context.Context,
	currentRes model.Resource,
	newSpec model.ResourceSpec,
	response *restful.Response,
) {
	if err := r.resourceAccess.ValidateUpdate(
		ctx,
		model.ResourceKey{Mesh: currentRes.GetMeta().GetMesh(), Name: currentRes.GetMeta().GetName()},
		currentRes.GetSpec(),
		newSpec,
		r.descriptor,
		user.FromCtx(ctx),
	); err != nil {
		rest_errors.HandleError(ctx, response, err, "Access Denied")
		return
	}

	_ = currentRes.SetSpec(newSpec)

	if err := r.resManager.Update(ctx, currentRes); err != nil {
		rest_errors.HandleError(ctx, response, err, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceEndpoints) addDeleteEndpoint(ws *restful.WebService, pathPrefix string) {
	if r.descriptor.ReadOnly {
		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(func(request *restful.Request, response *restful.Response) {
			rest_errors.HandleError(request.Request.Context(), response, &rest_errors.MethodNotAllowed{}, r.readOnlyMessage())
		}).
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
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not delete a resource")
		return
	}

	if err := r.resourceAccess.ValidateDelete(
		request.Request.Context(),
		model.ResourceKey{Mesh: meshName, Name: name},
		resource.GetSpec(),
		resource.Descriptor(),
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not delete a resource")
	}
}

func (r *resourceEndpoints) validateResourceRequest(request *restful.Request, resourceMeta rest_v1alpha1.ResourceMeta, create bool) error {
	var err validators.ValidationError
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)
	if name != resourceMeta.Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if r.mode == config_core.Zone && !r.doesNameLengthFitsGlobal(name) {
		err.AddViolation("name", "the length of the name must be shorter")
	}
	if string(r.descriptor.Name) != resourceMeta.Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if r.descriptor.Scope == model.ScopeMesh && meshName != resourceMeta.Mesh {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}
	if create {
		err.AddError("", mesh.ValidateMeta(resourceMeta, r.descriptor.Scope))
	} else {
		if verr, msg := mesh.ValidateMetaBackwardsCompatible(resourceMeta, r.descriptor.Scope); verr.HasViolations() {
			err.AddError("", verr)
		} else if msg != "" {
			core.Log.Info(msg)
		}
	}
	return err.OrNil()
}

// The resource is prefixed with the zone name when it is synchronized
// to global control-plane. It is important to notice that the zone is unaware
// of the type of the store used by the global control-plane, so we must prepare
// for the worst-case scenario. We don't have to check other plugabble policies
// because zone doesn't allow to create policies on the zone.
func (r *resourceEndpoints) doesNameLengthFitsGlobal(name string) bool {
	return len(fmt.Sprintf("%s.%s", r.zoneName, name)) < 253
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
