package api_server

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/api-server/definitions"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	rest_errors "github.com/Kong/kuma/pkg/core/rest/errors"
	"github.com/Kong/kuma/pkg/core/validators"
)

type meshFromRequestFn = func(*restful.Request) string

func meshFromPathParam(param string) meshFromRequestFn {
	return func(request *restful.Request) string {
		return request.PathParameter(param)
	}
}

type resourceEndpoints struct {
	resManager      manager.ResourceManager
	meshFromRequest meshFromRequestFn
	definitions.ResourceWsDefinition
}

func (r *resourceEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (r *resourceEndpoints) findResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	resource := r.ResourceFactory()
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
		Doc(fmt.Sprintf("List of %s", r.Name)).
		Returns(200, "OK", nil))
}

func (r *resourceEndpoints) listResources(request *restful.Request, response *restful.Response) {
	meshName := r.meshFromRequest(request)

	list := r.ResourceListFactory()
	if err := r.resManager.List(request.Request.Context(), list, store.ListByMesh(meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve resources")
	} else {
		restList := rest.From.ResourceList(list)
		if err := response.WriteAsJson(restList); err != nil {
			rest_errors.HandleError(response, err, "Could not list resources")
		}
	}
}

func (r *resourceEndpoints) addCreateOrUpdateEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.createOrUpdateResource).
		Doc(fmt.Sprintf("Updates a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(201, "Created", nil))
}

func (r *resourceEndpoints) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	resourceRes := rest.Resource{
		Spec: r.ResourceFactory().GetSpec(),
	}

	if err := request.ReadEntity(&resourceRes); err != nil {
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	if err := r.validateResourceRequest(request, &resourceRes); err != nil {
		rest_errors.HandleError(response, err, "Could not process a resource")
		return
	}

	resource := r.ResourceFactory()
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
	res := r.ResourceFactory()
	_ = res.SetSpec(spec)
	if err := r.resManager.Create(ctx, res, store.CreateByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceEndpoints) updateResource(ctx context.Context, res model.Resource, restRes rest.Resource, response *restful.Response) {
	_ = res.SetSpec(restRes.Spec)
	if err := r.resManager.Update(ctx, res); err != nil {
		rest_errors.HandleError(response, err, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceEndpoints) addDeleteEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.deleteResource).
		Doc(fmt.Sprintf("Deletes a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		Returns(200, "OK", nil))
}

func (r *resourceEndpoints) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)

	resource := r.ResourceFactory()
	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(name, meshName)); err != nil {
		rest_errors.HandleError(response, err, "Could not delete a resource")
	}
}

func (r *resourceEndpoints) validateResourceRequest(request *restful.Request, resource *rest.Resource) error {
	var err validators.ValidationError
	name := request.PathParameter("name")
	meshName := r.meshFromRequest(request)
	if name != resource.Meta.Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if string(r.ResourceFactory().GetType()) != resource.Meta.Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if meshName != resource.Meta.Mesh && r.ResourceFactory().GetType() != mesh.MeshType {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}
	err.AddError("", mesh.ValidateMeta(name, meshName))
	return err.OrNil()
}
