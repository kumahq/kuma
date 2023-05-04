package api_server

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	rest_unversioned "github.com/kumahq/kuma/pkg/core/resources/model/rest/unversioned"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/insights"
)

type serviceInsightEndpoints struct {
	resourceEndpoints
}

func (s *serviceInsightEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{service}").To(s.findResource).
		Doc(fmt.Sprintf("Get a %s", s.descriptor.WsPath)).
		Param(ws.PathParameter("service", fmt.Sprintf("Name of a %s", s.descriptor.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (s *serviceInsightEndpoints) findResource(request *restful.Request, response *restful.Response) {
	service := request.PathParameter("service")
	meshName := s.meshFromRequest(request)

	serviceInsight := mesh.NewServiceInsightResource()
	err := s.resManager.Get(request.Request.Context(), serviceInsight, store.GetBy(insights.ServiceInsightKey(meshName)))
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a resource")
	} else {
		stat := serviceInsight.Spec.Services[service]
		if stat == nil {
			stat = &v1alpha1.ServiceInsight_Service{
				Dataplanes: &v1alpha1.ServiceInsight_Service_DataplaneStat{},
			}
		}
		res := rest_unversioned.From.Resource(serviceInsight)
		res.Meta.Name = service
		res.Spec = stat
		if err := response.WriteAsJson(res); err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}

func (s *serviceInsightEndpoints) addListEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix).To(s.listResources).
		Doc(fmt.Sprintf("List of %s", s.descriptor.Name)).
		Param(ws.PathParameter("size", "size of page").DataType("int")).
		Param(ws.PathParameter("offset", "offset of page to list").DataType("string")).
		Returns(200, "OK", nil))
}

func (s *serviceInsightEndpoints) listResources(request *restful.Request, response *restful.Response) {
	meshName := s.meshFromRequest(request)

	serviceInsightList := &mesh.ServiceInsightResourceList{}
	err := s.resManager.List(request.Request.Context(), serviceInsightList, store.ListByMesh(meshName))
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve resources")
		return
	}

	restList := s.expandInsights(serviceInsightList)
	restList.Total = uint32(len(restList.Items))

	if err := s.paginateResources(request, &restList); err != nil {
		rest_errors.HandleError(response, err, "Could not paginate resources")
		return
	}

	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(response, err, "Could not list resources")
	}
}

// ServiceInsight is a resource that tracks insights about a Services (kuma.io/service tag in Dataplane)
// All of those statistics are put into a single ServiceInsight for a mesh for two reasons
// 1) It simpler and more efficient to manage 1 object because 1 Service != 1 Dataplane
// 2) Mesh+Name is a key on Universal, but not on Kubernetes, so if there are two services of the same name in different Meshes we would have problems with naming.
// From the API perspective it's better to provide ServiceInsight per Service, not per Mesh.
// For this reason, this method expand the one ServiceInsight resource for the mesh to resource per service
func (s *serviceInsightEndpoints) expandInsights(serviceInsightList *mesh.ServiceInsightResourceList) rest.ResourceList {
	restItems := []*rest_unversioned.Resource{}
	for _, insight := range serviceInsightList.Items {
		for serviceName, stat := range insight.Spec.Services {
			res := rest_unversioned.From.Resource(insight)
			res.Meta.Name = serviceName
			res.Spec = stat
			restItems = append(restItems, res)
		}
	}

	sort.Sort(rest_unversioned.ByMeta(restItems))

	restList := rest.ResourceList{}
	for _, item := range restItems {
		restList.Items = append(restList.Items, item)
	}
	return restList
}

// paginateResources paginates resources manually, because we are expanding resources.
func (s *serviceInsightEndpoints) paginateResources(request *restful.Request, restList *rest.ResourceList) error {
	page, err := pagination(request)
	if err != nil {
		return err
	}

	offset := 0
	if page.offset != "" {
		o, err := strconv.Atoi(page.offset)
		if err != nil {
			return store.ErrorInvalidOffset
		}
		offset = o
	}

	total := int(restList.Total)
	start := offset
	if offset >= total {
		start = total
	}
	end := start + page.size
	if end >= total {
		end = total
	}
	restList.Items = restList.Items[start:end]

	nextOffset := ""
	if offset+page.size < total {
		nextOffset = strconv.Itoa(offset + page.size)
	}

	restList.Next = nextLink(request, nextOffset)
	return nil
}
