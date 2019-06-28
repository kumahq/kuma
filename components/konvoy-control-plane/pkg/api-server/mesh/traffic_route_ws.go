package mesh

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
	sample_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
	"github.com/emicklei/go-restful"
	"log"
)

const namespace = "default"

type TrafficRouteWs struct {
	ResourceStore store.ResourceStore
}

func NewProxyTemplateWs(resourceStore store.ResourceStore) *restful.WebService {
	p := &TrafficRouteWs{ResourceStore: resourceStore}
	ws := new(restful.WebService)

	ws.
		Path("/meshes/{mesh-mesh}/traffic-routes").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("mesh-name", "Name of the Mesh").DataType("string"))

	ws.Route(ws.GET("/{name}").To(p.findResource).
		Doc("Get a Traffic Route").
		Param(ws.PathParameter("name", "Name of the Traffic Route").DataType("string")).
		Writes(sample_proto.TrafficRoute{}).
		Returns(200, "OK", sample_proto.TrafficRoute{}).
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("").To(p.listResources).
		Doc("List of Traffic Route").
		Writes(sample_proto.TrafficRoute{}).
		Returns(200, "OK", sample_proto.TrafficRoute{}))

	ws.Route(ws.PUT("/{name}").To(p.createOrUpdateResource).
		Doc("Updates a Traffic Route").
		Param(ws.PathParameter("name", "Name of the Traffic Route").DataType("string")).
		Reads(sample_proto.TrafficRoute{}).
		Returns(200, "OK", nil).
		Returns(201, "Created", nil))

	ws.Route(ws.DELETE("/{name}").To(p.deleteResource).
		Doc("Deletes a Traffic Route").
		Param(ws.PathParameter("name", "Name of the Traffic Route").DataType("string")).
		Returns(200, "OK", nil))

	return ws
}

type resourceSpecList struct {
	Items []model.ResourceSpec `json:"items"`
}

func (ws *TrafficRouteWs) listResources(request *restful.Request, response *restful.Response) {
	list := sample_model.TrafficRouteResourceList{}
	err := ws.ResourceStore.List(context.Background(), &list, store.ListByNamespace(namespace))
	if err != nil {
		// log
		writeError(response, 500, "Could not list a resource")
	} else {
		var items []model.ResourceSpec
		for _, item := range list.Items {
			items = append(items, &item.Spec)
		}
		specList := resourceSpecList{Items:items}
		if err := response.WriteAsJson(specList); err != nil {
			log.Printf("Could not write as JSON %+v", err)
			writeError(response, 500, "Could not list a resource")
		}
	}
}

func (ws *TrafficRouteWs) findResource(request *restful.Request, response *restful.Response) {
	r := sample_model.TrafficRouteResource{}
	// todo mesh is not ignored
	name := request.PathParameter("name")
	// validate if not empty?
	err := ws.ResourceStore.Get(context.Background(), &r, store.GetByName(namespace, name)); if err != nil {
		if err.Error() == store.ErrorResourceNotFound(r.GetType(), namespace, name).Error() {
			writeError(response, 404, "")
		} else {
			// log
			writeError(response, 500, "Could not retrieve a resource")
		}
	} else {
		err = response.WriteAsJson(r.Spec); if err != nil {
			log.Printf("Could not write the response: %v", err)
		}
	}
}

func (ws *TrafficRouteWs) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	r := sample_proto.TrafficRoute{}
	// todo mesh is not ignored
	name := request.PathParameter("name")
	err := request.ReadEntity(&r); if err != nil {
		log.Printf("Could not read an entity %+v", err)
		writeError(response, 400, "Could not process the resource")
	}

	res := sample_model.TrafficRouteResource{}
	err = ws.ResourceStore.Get(context.Background(), &res, store.GetByName(namespace, name))
	if err != nil {
		if err.Error() == store.ErrorResourceNotFound(res.GetType(), namespace, name).Error() {
			ws.createResource(name, r, response)
		} else {
			log.Printf("Could not get a resource from the store %v", err)
			writeError(response, 500, "Could not create a resource")
		}
	} else {
		ws.updateResource(res, r, response)
	}
}

func (ws *TrafficRouteWs) createResource(name string, r sample_proto.TrafficRoute, response *restful.Response) {
	res := sample_model.TrafficRouteResource{}
	res.Spec = r
	if err := ws.ResourceStore.Create(context.Background(), &res, store.CreateByName(namespace, name)); err != nil {
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (ws *TrafficRouteWs) updateResource(res sample_model.TrafficRouteResource, r sample_proto.TrafficRoute, response *restful.Response) {
	res.Spec = r
	if err := ws.ResourceStore.Update(context.Background(), &res); err != nil {
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (ws *TrafficRouteWs) deleteResource(request *restful.Request, response *restful.Response) {
	resource := sample_model.TrafficRouteResource{}
	// todo mesh is not ignored
	name := request.PathParameter("name")

	err := ws.ResourceStore.Delete(context.Background(), &resource, store.DeleteByName(namespace, name))
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