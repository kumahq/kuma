package api_server

import (
	"context"
	"fmt"
	"github.com/Kong/kuma/pkg/api-server/definitions"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/emicklei/go-restful"
	"regexp"
)

const namespace = "default"

type resourceWs struct {
	resManager      manager.ResourceManager
	readOnly        bool
	nameFromRequest func(*restful.Request) string
	meshFromRequest func(*restful.Request) string
	definitions.ResourceWsDefinition
}

func (r *resourceWs) AddToWs(ws *restful.WebService) {
	pathPrefix := ""
	if r.ResourceFactory().GetType() == mesh.MeshType {
		r.nameFromRequest = func(request *restful.Request) string {
			return request.PathParameter("name")
		}
		r.meshFromRequest = func(request *restful.Request) string {
			return request.PathParameter("name")
		}
	} else {
		pathPrefix += "/{mesh}/" + r.Path
		r.nameFromRequest = func(request *restful.Request) string {
			return request.PathParameter("name")
		}
		r.meshFromRequest = func(request *restful.Request) string {
			return request.PathParameter("mesh")
		}
	}

	ws.Route(ws.GET(pathPrefix+"/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		//Writes(r.SpecFactory()).
		Returns(200, "OK", nil). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
		Returns(404, "Not found", nil))

	ws.Route(ws.GET(pathPrefix).To(r.listResources).
		Doc(fmt.Sprintf("List of %s", r.Name)).
		//Writes(r.SampleListSpec).
		Returns(200, "OK", nil)) // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp

	if !r.readOnly {
		ws.Route(ws.PUT(pathPrefix+"/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.Name)).DataType("string")).
			//Reads(r.SampleSpec). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))

		ws.Route(ws.DELETE(pathPrefix+"/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}
}

func (r *resourceWs) findResource(request *restful.Request, response *restful.Response) {
	name := r.nameFromRequest(request)
	meshName := r.meshFromRequest(request)

	resource := r.ResourceFactory()
	err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(namespace, name, meshName))
	if err != nil {
		handleError(response, err, "Could not retrieve a resource")
	} else {
		res := rest.From.Resource(resource)
		if err := response.WriteAsJson(res); err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}

func (r *resourceWs) listResources(request *restful.Request, response *restful.Response) {
	meshName := r.meshFromRequest(request)

	list := r.ResourceListFactory()
	if err := r.resManager.List(request.Request.Context(), list, store.ListByMesh(meshName)); err != nil {
		handleError(response, err, "Could not retrieve resources")
	} else {
		restList := rest.From.ResourceList(list)
		if err := response.WriteAsJson(restList); err != nil {
			handleError(response, err, "Could not list resources")
		}
	}
}

func (r *resourceWs) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := r.nameFromRequest(request)
	meshName := r.meshFromRequest(request)

	resourceRes := rest.Resource{
		Spec: r.ResourceFactory().GetSpec(),
	}

	if err := request.ReadEntity(&resourceRes); err != nil {
		handleError(response, err, "Could not process a resource")
		return
	}

	if err := r.validateResourceRequest(request, &resourceRes); err != nil {
		handleError(response, err, "Could not process a resource")
		return
	}

	resource := r.ResourceFactory()
	if err := r.resManager.Get(request.Request.Context(), resource, store.GetByKey(namespace, name, meshName)); err != nil {
		if store.IsResourceNotFound(err) {
			r.createResource(request.Request.Context(), name, meshName, resourceRes.Spec, response)
		} else {
			handleError(response, err, "Could not find a resource")
		}
	} else {
		r.updateResource(request.Request.Context(), resource, resourceRes, response)
	}
}

var nameMeshRegexp = regexp.MustCompile("^[0-9a-z-_]*$")

func (r *resourceWs) validateResourceRequest(request *restful.Request, resource *rest.Resource) error {
	var err validators.ValidationError
	name := r.nameFromRequest(request)
	meshName := r.meshFromRequest(request)
	if name != resource.Meta.Name {
		err.AddViolation("name", "name from the URL has to be the same as in body")
	}
	if !nameMeshRegexp.MatchString(name) {
		err.AddViolation("name", "invalid characters. Valid characters are numbers, letters and '-', '_' symbols.")
	}
	if string(r.ResourceFactory().GetType()) != resource.Meta.Type {
		err.AddViolation("type", "type from the URL has to be the same as in body")
	}
	if meshName != resource.Meta.Mesh && r.ResourceFactory().GetType() != mesh.MeshType {
		err.AddViolation("mesh", "mesh from the URL has to be the same as in body")
	}
	if !nameMeshRegexp.MatchString(meshName) {
		err.AddViolation("mesh", "invalid characters. Valid characters are numbers, letters and '-', '_' symbols.")
	}
	return err.OrNil()
}

func (r *resourceWs) createResource(ctx context.Context, name string, meshName string, spec model.ResourceSpec, response *restful.Response) {
	res := r.ResourceFactory()
	_ = res.SetSpec(spec)
	if err := r.resManager.Create(ctx, res, store.CreateByKey(namespace, name, meshName)); err != nil {
		handleError(response, err, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceWs) updateResource(ctx context.Context, res model.Resource, restRes rest.Resource, response *restful.Response) {
	_ = res.SetSpec(restRes.Spec)
	if err := r.resManager.Update(ctx, res); err != nil {
		handleError(response, err, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceWs) deleteResource(request *restful.Request, response *restful.Response) {
	name := r.nameFromRequest(request)
	meshName := r.meshFromRequest(request)

	resource := r.ResourceFactory()
	if err := r.resManager.Delete(request.Request.Context(), resource, store.DeleteByKey(namespace, name, meshName)); err != nil {
		handleError(response, err, "Could not delete a resource")
	}
}
