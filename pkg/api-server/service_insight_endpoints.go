package api_server

import (
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"

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
	addressPortGenerator func(string) string
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
	meshName, err := s.meshFromRequest(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
		return
	}

	serviceInsight := mesh.NewServiceInsightResource()
	err = s.resManager.Get(request.Request.Context(), serviceInsight, store.GetBy(insights.ServiceInsightKey(meshName)))
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
	} else {
		stat := serviceInsight.Spec.Services[service]
		if stat == nil {
			stat = &v1alpha1.ServiceInsight_Service{
				Dataplanes: &v1alpha1.ServiceInsight_Service_DataplaneStat{},
			}
		}
		s.fillStaticInfo(service, stat)
		out := rest.From.Resource(serviceInsight)
		res := out.(*rest_unversioned.Resource)
		res.Meta.Name = service
		res.Spec = stat

		mapperFn := removeDisplayNameLabel()
		mapperFn(res)

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
		Param(ws.QueryParameter("name", "a pattern to select only services that contain these characters").DataType("string")).
		Returns(200, "OK", nil))
}

func (s *serviceInsightEndpoints) listResources(request *restful.Request, response *restful.Response) {
	meshName, err := s.meshFromRequest(request)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Mesh")
		return
	}

	serviceInsightList := &mesh.ServiceInsightResourceList{}
	err = s.resManager.List(request.Request.Context(), serviceInsightList, store.ListByMesh(meshName))
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve resources")
		return
	}

	nameContains := request.QueryParameter("name")
	filters := request.QueryParameter("type")
	filterMap := map[v1alpha1.ServiceInsight_Service_Type]struct{}{}
	if filters != "" {
		for _, f := range strings.Split(filters, ",") {
			f = strings.ToLower(strings.TrimSpace(f))
			i, exists := v1alpha1.ServiceInsight_Service_Type_value[f]
			if !exists {
				rest_errors.HandleError(request.Request.Context(), response, rest_errors.NewBadRequestError("unsupported service type"), "Invalid response type")
				return
			}
			filterMap[v1alpha1.ServiceInsight_Service_Type(i)] = struct{}{}
		}
	}

	items := s.expandInsights(serviceInsightList, nameContains,
		func(service *v1alpha1.ServiceInsight_Service) bool {
			if len(filterMap) == 0 {
				return true
			}
			_, exists := filterMap[service.ServiceType]
			return exists
		},
		removeDisplayNameLabel(),
	)

	restList := rest.ResourceList{
		Total: uint32(len(items)),
		Items: items,
	}

	if err := s.paginateResources(request, &restList); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not paginate resources")
		return
	}

	if err := response.WriteAsJson(restList); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not list resources")
	}
}

// fillStaticInfo fills static information, so we won't have to store this in the DB
func (s *serviceInsightEndpoints) fillStaticInfo(name string, stat *v1alpha1.ServiceInsight_Service) {
	stat.Dataplanes.Total = stat.Dataplanes.Online + stat.Dataplanes.Offline
	if stat.ServiceType == v1alpha1.ServiceInsight_Service_internal {
		stat.AddressPort = s.addressPortGenerator(name)
	}
}

// ServiceInsight is a resource that tracks insights about a Services (kuma.io/service tag in Dataplane)
// All of those statistics are put into a single ServiceInsight for a mesh for two reasons
// 1) It simpler and more efficient to manage 1 object because 1 Service != 1 Dataplane
// 2) Mesh+Name is a key on Universal, but not on Kubernetes, so if there are two services of the same name in different Meshes we would have problems with naming.
// From the API perspective it's better to provide ServiceInsight per Service, not per Mesh.
// For this reason, this method expand the one ServiceInsight resource for the mesh to resource per service
func (s *serviceInsightEndpoints) expandInsights(
	serviceInsightList *mesh.ServiceInsightResourceList,
	nameContains string,
	filterFn func(service *v1alpha1.ServiceInsight_Service) bool,
	mapperFn unversionedResourceMapper,
) []rest.Resource {
	restItems := []rest.Resource{} // Needs to be set to avoid returning nil and have the api return []
	for _, insight := range serviceInsightList.Items {
		for serviceName, service := range insight.Spec.Services {
			if strings.Contains(serviceName, nameContains) && filterFn(service) {
				s.fillStaticInfo(serviceName, service)
				out := rest.From.Resource(insight)
				res := out.(*rest_unversioned.Resource)
				res.Meta.Name = serviceName
				res.Spec = service
				mapperFn(res)
				restItems = append(restItems, out)
			}
		}
	}

	sort.Slice(restItems, func(i, j int) bool {
		metai := restItems[i].GetMeta()
		metaj := restItems[j].GetMeta()
		if metai.Mesh == metaj.Mesh {
			return metai.Name < metaj.Name
		}
		return metai.Mesh < metaj.Mesh
	})
	return restItems
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

type unversionedResourceMapper func(resource *rest_unversioned.Resource)

// Since the value of label "kuma.io/display-name" is same with the ServiceInsight resource name,
// in which it looks weird for the API to each service. Ref: https://github.com/kumahq/kuma/issues/9729
func removeDisplayNameLabel() unversionedResourceMapper {
	return func(resource *rest_unversioned.Resource) {
		tmpMeta := resource.GetMeta()
		maps.DeleteFunc(tmpMeta.Labels, func(key string, val string) bool {
			return key == v1alpha1.DisplayName
		})
		resource.Meta = tmpMeta
	}
}
