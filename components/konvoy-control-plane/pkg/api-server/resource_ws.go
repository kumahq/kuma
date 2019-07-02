package api_server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"log"
)

const namespace = "default"

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
	SpecFactory         func() model.ResourceSpec
	SampleSpec			interface{}
}

type resourceWs struct {
	resourceStore store.ResourceStore
	readOnly      bool
	ResourceWsDefinition
}

type ResourceResponse struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Mesh string `json:"mesh"`
	Spec model.ResourceSpec `json:"-"`
}

// normally Spec would be embedded under Spec key. embedding model.ResourceSpec does not work because the marshaller
// cannot expand embedded struct over the interface. We have to implement our marshaller to expand the Spec
func (f ResourceResponse) MarshalJSON() ([]byte, error) {
	type tmp ResourceResponse // otherwise we've got stackoverflow
	g := tmp(f)
	first, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	second, err := json.Marshal(f.Spec)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	if err := json.Unmarshal(first, &data); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(second, &data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (r *resourceWs) NewWs() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path(fmt.Sprintf("/meshes/{mesh}/%s", r.Path)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("mesh-name", "Name of the Mesh").DataType("string"))

	ws.Route(ws.GET("/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		//Writes(r.SpecFactory()).
		Returns(200, "OK", nil). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceResponse
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("").To(r.listResources).
		Doc(fmt.Sprintf("List of %s", r.Name)).
		//Writes(r.SampleListSpec).
		Returns(200, "OK", nil)) // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceResponse

	if !r.readOnly {
		ws.Route(ws.PUT("/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.Name)).DataType("string")).
			Reads(r.SampleSpec).
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))

		ws.Route(ws.DELETE("/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}

	return ws
}

type resourceSpecList struct {
	Items []ResourceResponse `json:"items"`
}

func (r *resourceWs) listResources(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")

	list := r.ResourceListFactory()
	err := r.resourceStore.List(context.Background(), list, store.ListByNamespace(namespace))
	if err != nil {
		log.Printf("Could not retrieve resources %v", err)
		writeError(response, 500, "Could not list a resource")
	} else {
		var items []ResourceResponse
		for _, item := range list.GetItems() {
			items = append(items, ResourceResponse{
				Type: string(item.GetType()),
				Name: item.GetMeta().GetName(),
				Mesh: meshName,
				Spec: item.GetSpec(),
			})
		}
		specList := resourceSpecList{Items: items}
		if err := response.WriteAsJson(specList); err != nil {
			log.Printf("Could not write as JSON %v", err)
			writeError(response, 500, "Could not list a resource")
		}
	}
}

func (r *resourceWs) findResource(request *restful.Request, response *restful.Response) {
	resource := r.ResourceFactory()
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	// todo(jakubdyszkiewicz) find by mesh?
	err := r.resourceStore.Get(context.Background(), resource, store.GetByName(namespace, name));
	if err != nil {
		if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
			writeError(response, 404, "")
		} else {
			log.Printf("Could not retrieve a resource %v", err)
			writeError(response, 500, "Could not retrieve a resource")
		}
	} else {
		res := ResourceResponse{
			Type: string(resource.GetType()),
			Name: name,
			Mesh: meshName,
			Spec: resource.GetSpec(),
		}
		err = response.WriteAsJson(res);
		if err != nil {
			log.Printf("Could not write the response: %v", err)
		}
	}
}

func (r *resourceWs) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	spec := r.SpecFactory()
	err := request.ReadEntity(spec);
	if err != nil {
		log.Printf("Could not read an entity %+v", err)
		writeError(response, 400, "Could not process the resource")
	}

	resource := r.ResourceFactory()
	// todo(jakubdyszkiewicz) find by mesh?
	err = r.resourceStore.Get(context.Background(), resource, store.GetByName(namespace, name))
	if err != nil {
		if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
			r.createResource(name, spec, response)
		} else {
			log.Printf("Could not get a resource from the store %v", err)
			writeError(response, 500, "Could not create a resource")
		}
	} else {
		r.updateResource(resource, spec, response)
	}
}

func (r *resourceWs) createResource(name string, spec model.ResourceSpec, response *restful.Response) {
	res := r.ResourceFactory()
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Create(context.Background(), res, store.CreateByName(namespace, name)); err != nil {
		log.Printf("Could not create a resource %v", err)
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceWs) updateResource(res model.Resource, spec model.ResourceSpec, response *restful.Response) {
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Update(context.Background(), res); err != nil {
		log.Printf("Could not create a resource %v", err)
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceWs) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	resource := r.ResourceFactory()
	// todo(jakubdyszkiewicz) delete by mesh?
	err := r.resourceStore.Delete(context.Background(), resource, store.DeleteByName(namespace, name))
	if err != nil {
		writeError(response, 500, "Could not delete a resource")
		log.Printf("Could not delete a resource %+v", err)
	}
}

func writeError(response *restful.Response, httpStatus int, msg string) {
	if err := response.WriteErrorString(httpStatus, msg); err != nil {
		log.Printf("Could not write the response: %v", err)
	}
}
