package api_server

import (
	"fmt"
	"sort"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	mes_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/hostname"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	mzms_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/hostname"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/hostname"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/validators"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

func addInspectMeshServiceEndpoints(
	ws *restful.WebService,
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/meshservices/{name}/_resources/dataplanes").
			To(matchingDataplanesForMeshServices(rm, resourceAccess)).
			Doc("inspect dataplane configuration and stats").
			Param(ws.PathParameter("name", "mesh service name").DataType("string")).
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")),
	)
	ws.Route(
		ws.GET("/meshes/{mesh}/{serviceType}/{name}/_hostnames").
			To(matchingHostnames(rm)).
			Doc("inspect service hostnames").
			Param(ws.PathParameter("name", "mesh service name").DataType("string")).
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")),
	)
}

func matchingDataplanesForMeshServices(resManager manager.ResourceManager, resourceAccess access.ResourceAccess) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		matchingDataplanesForFilter(
			request,
			response,
			meshservice_api.MeshServiceResourceTypeDescriptor,
			resManager,
			resourceAccess,
			func(resource model.Resource) store.ListFilterFunc {
				meshService := resource.(*meshservice_api.MeshServiceResource)
				return func(rs model.Resource) bool {
					return meshservice.MatchesDataplane(meshService.Spec, rs.(*core_mesh.DataplaneResource))
				}
			},
		)
	}
}

var availableServiceTypes = []string{
	string(types.Meshservices),
	string(types.Meshexternalservices),
	string(types.Meshmultizoneservices),
}

func matchingHostnames(resManager manager.ResourceManager) restful.RouteFunction {
	generatorsForType := map[types.InspectHostnamesParamsServiceType]hostname.HostnameGenerator{
		types.Meshservices:          meshservice_hostname.NewMeshServiceHostnameGenerator(resManager),
		types.Meshexternalservices:  mes_hostname.NewMeshExternalServiceHostnameGenerator(resManager),
		types.Meshmultizoneservices: mzms_hostname.NewMeshMultiZoneServiceHostnameGenerator(resManager),
	}
	typeDescForType := map[types.InspectHostnamesParamsServiceType]model.ResourceTypeDescriptor{
		types.Meshservices:          meshservice_api.MeshServiceResourceTypeDescriptor,
		types.Meshexternalservices:  meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
		types.Meshmultizoneservices: meshmultizoneservice_api.MeshMultiZoneServiceResourceTypeDescriptor,
	}

	return func(request *restful.Request, response *restful.Response) {
		svcName := request.PathParameter("name")
		svcMesh := request.PathParameter("mesh")
		svcType := types.InspectHostnamesParamsServiceType(request.PathParameter("serviceType"))

		desc, ok := typeDescForType[svcType]
		if !ok {
			rest_errors.HandleError(
				request.Request.Context(),
				response,
				&validators.ValidationError{},
				fmt.Sprintf("only %q are available for inspection", strings.Join(availableServiceTypes, ",")),
			)
		}

		svc := desc.NewObject()
		if err := resManager.Get(request.Request.Context(), svc, store.GetByKey(svcName, svcMesh)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "could not retrieve service")
			return
		}

		hostnameGenerators := hostnamegenerator_api.HostnameGeneratorResourceList{}
		if err := resManager.List(request.Request.Context(), &hostnameGenerators); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "could not retrieve hostname generators")
			return
		}

		byHostname := map[string][]types.InspectHostnameZone{}

		for _, hg := range hostnameGenerators.Items {
			svcZone := model.ZoneOfResource(svc)
			hgZone := model.ZoneOfResource(hg)
			hgOrigin, _ := model.ResourceOrigin(hg.GetMeta())

			// rewrite origin to simulate matching from a perspective of a single zone
			var origin mesh_proto.ResourceOrigin
			if hgOrigin == mesh_proto.ZoneResourceOrigin && hgZone == svcZone {
				origin = mesh_proto.ZoneResourceOrigin
			} else {
				origin = mesh_proto.GlobalResourceOrigin
			}
			overridden := ResourceMetaWithOverriddenOrigin{
				ResourceMeta: svc.GetMeta(),
				origin:       origin,
			}
			svc.SetMeta(overridden)

			host, err := generatorsForType[svcType].GenerateHostname(svcZone, hg, svc)
			if err != nil {
				rest_errors.HandleError(request.Request.Context(), response, err, "could not generate hostname")
				return
			}
			if host == "" {
				// hostname generator did not match the service
				continue
			}
			byHostname[host] = append(byHostname[host], types.InspectHostnameZone{
				Name: svcZone,
			})
		}

		resp := types.InspectHostnames{}
		for _, host := range util_maps.SortedKeys(byHostname) {
			zones := byHostname[host]
			sort.Slice(zones, func(i, j int) bool {
				return zones[i].Name < zones[j].Name
			})
			resp.Items = append(resp.Items, types.InspectHostname{
				Hostname: host,
				Zones:    zones,
			})
		}
		resp.Total = len(resp.Items)
		if err := response.WriteAsJson(resp); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Failed writing response")
		}
	}
}

type ResourceMetaWithOverriddenOrigin struct {
	model.ResourceMeta
	origin mesh_proto.ResourceOrigin
}

func (r ResourceMetaWithOverriddenOrigin) GetLabels() map[string]string {
	cloned := maps.Clone(r.ResourceMeta.GetLabels())
	if cloned == nil {
		cloned = map[string]string{}
	}
	cloned[mesh_proto.ResourceOriginLabel] = string(r.origin)
	return cloned
}
