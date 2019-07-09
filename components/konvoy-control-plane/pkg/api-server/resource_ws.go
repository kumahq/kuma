package api_server

import (
	"context"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"
)

const namespace = "default"

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
}

type resourceWs struct {
	resourceStore store.ResourceStore
	readOnly      bool
	ResourceWsDefinition
}

func (r *resourceWs) NewWs() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path(fmt.Sprintf("/meshes/{mesh}/%s", r.Path)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("mesh", "Name of the Mesh").DataType("string"))

	ws.Route(ws.GET("/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		//Writes(r.SpecFactory()).
		Returns(200, "OK", nil). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("").To(r.listResources).
		Doc(fmt.Sprintf("List of %s", r.Name)).
		//Writes(r.SampleListSpec).
		Returns(200, "OK", nil)) // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp

	if !r.readOnly {
		ws.Route(ws.PUT("/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.Name)).DataType("string")).
			//Reads(r.SampleSpec). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))

		ws.Route(ws.DELETE("/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}

	return ws
}

func (r *resourceWs) findResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	// todo(jakubdyszkiewicz) find by mesh?
	resource := r.ResourceFactory()
	err := r.resourceStore.Get(request.Request.Context(), resource, store.GetByName(namespace, name))
	if err != nil {
		if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
			writeError(response, 404, "")
		} else {
			core.Log.Error(err, "Could not retrieve a resource", "name", name)
			writeError(response, 500, "Could not retrieve a resource")
		}
	} else {
		res := &rest.Resource{
			Meta: rest.ResourceMeta{
				Mesh: meshName,
				Type: string(resource.GetType()),
				Name: name,
			},
			Spec: resource.GetSpec(),
		}
		err = response.WriteAsJson(res)
		if err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}

type resourceSpecList struct {
	Items []*rest.Resource `json:"items"`
}

func (r *resourceWs) listResources(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")

	list := r.ResourceListFactory()
	// todo(jakubdyszkiewicz) find by mesh?
	if err := r.resourceStore.List(request.Request.Context(), list, store.ListByNamespace(namespace)); err != nil {
		core.Log.Error(err, "Could not retrieve resources")
		writeError(response, 500, "Could not list a resource")
	} else {
		var items []*rest.Resource
		for _, item := range list.GetItems() {
			items = append(items, &rest.Resource{
				Meta: rest.ResourceMeta{
					Mesh: meshName,
					Type: string(item.GetType()),
					Name: item.GetMeta().GetName(),
				},
				Spec: item.GetSpec(),
			})
		}
		specList := resourceSpecList{Items: items}
		if err := response.WriteAsJson(specList); err != nil {
			core.Log.Error(err, "Could not write as JSON", "type", string(list.GetItemType()))
			writeError(response, 500, "Could not list a resource")
		}
	}
}

func (r *resourceWs) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	resourceRes := rest.Resource{
		Spec: r.ResourceFactory().GetSpec(),
	}

	err := request.ReadEntity(&resourceRes)
	if err != nil {
		core.Log.Error(err, "Could not read an entity")
		writeError(response, 400, "Could not process the resource")
	}

	if err := r.validateResourceRequest(request, &resourceRes); err != nil {
		writeError(response, 400, err.Error())
	} else {
		resource := r.ResourceFactory()
		// todo(jakubdyszkiewicz) find by mesh?
		if err := r.resourceStore.Get(request.Request.Context(), resource, store.GetByName(namespace, name)); err != nil {
			if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
				r.createResource(request.Request.Context(), name, resourceRes.Spec, response)
			} else {
				core.Log.Error(err, "Could get a resource from the store", "namespace", namespace, "name", name, "type", string(resource.GetType()))
				writeError(response, 500, "Could not create a resource")
			}
		} else {
			r.updateResource(request.Request.Context(), resource, resourceRes.Spec, response)
		}
	}
}

func (r *resourceWs) validateResourceRequest(request *restful.Request, resource *rest.Resource) error {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")
	if name != resource.Meta.Name {
		return errors.New("Name from the URL has to be the same as in body")
	}
	if string(r.ResourceFactory().GetType()) != resource.Meta.Type {
		return errors.New("Type from the URL has to be the same as in body")
	}
	if meshName != resource.Meta.Mesh {
		return errors.New("Mesh from the URL has to be the same as in body")
	}
	return nil
}

func (r *resourceWs) createResource(ctx context.Context, name string, spec model.ResourceSpec, response *restful.Response) {
	res := r.ResourceFactory()
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Create(ctx, res, store.CreateByName(namespace, name)); err != nil {
		core.Log.Error(err, "Could not create a resource")
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceWs) updateResource(ctx context.Context, res model.Resource, spec model.ResourceSpec, response *restful.Response) {
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Update(ctx, res); err != nil {
		core.Log.Error(err, "Could not update a resource")
		writeError(response, 500, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceWs) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	resource := r.ResourceFactory()
	// todo(jakubdyszkiewicz) delete by mesh?
	err := r.resourceStore.Delete(request.Request.Context(), resource, store.DeleteByName(namespace, name))
	if err != nil {
		writeError(response, 500, "Could not delete a resource")
		core.Log.Error(err, "Could not delete a resource", "namespace", namespace, "name", name, "type", string(resource.GetType()))
	}
}

func writeError(response *restful.Response, httpStatus int, msg string) {
	if err := response.WriteErrorString(httpStatus, msg); err != nil {
		core.Log.Error(err, "Cloud not write the response")
	}
}
