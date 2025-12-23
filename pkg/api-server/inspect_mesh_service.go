package api_server

import (
	"fmt"
	"maps"
	"sort"
	"strings"

	"github.com/emicklei/go-restful/v3"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/api/openapi/types"
	"github.com/kumahq/kuma/v2/pkg/core/resources/access"
	hostnamegenerator_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/hostnamegenerator/hostname"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	mes_hostname "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/hostname"
	meshmultizoneservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	mzms_hostname "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshmultizoneservice/hostname"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshservice_hostname "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/hostname"
	"github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/v2/pkg/core/rest/errors"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
	util_maps "github.com/kumahq/kuma/v2/pkg/util/maps"
)

func addInspectMeshServiceEndpoints(
	ws *restful.WebService,
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
	isGlobal bool,
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
			To(matchingHostnames(rm, isGlobal)).
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

func matchingHostnames(resManager manager.ResourceManager, isGlobal bool) restful.RouteFunction {
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

	generateAndRecord := func(svc model.Resource, svcType types.InspectHostnamesParamsServiceType, svcZone string, hg *hostnamegenerator_api.HostnameGeneratorResource, byHostname map[string]map[string]struct{}) error {
		host, err := generatorsForType[svcType].GenerateHostname(svcZone, hg, svc)
		if err != nil {
			return err
		}
		if host != "" {
			if byHostname[host] == nil {
				byHostname[host] = map[string]struct{}{}
			}
			byHostname[host][svcZone] = struct{}{}
		}
		return nil
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
			return
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

		byHostname := map[string]map[string]struct{}{}

		for _, hg := range hostnameGenerators.Items {
			svcZone := model.ZoneOfResource(svc)
			svcOrigin, _ := model.ResourceOrigin(svc.GetMeta())
			origins := []mesh_proto.ResourceOrigin{svcOrigin}
			// On Global control plane in multi-zone setup, hostname generators may use origin-based
			// selectors (e.g., kuma.io/origin label) to match services. Resources synced across zones
			// can appear with different origins, so we test both Global and Zone perspectives to capture
			// all possible hostname matches. Zone control planes only need to check the actual origin.
			if isGlobal {
				origins = []mesh_proto.ResourceOrigin{mesh_proto.ZoneResourceOrigin, mesh_proto.GlobalResourceOrigin}
			}

			for _, origin := range origins {
				overridden := ResourceMetaWithOverriddenOrigin{
					ResourceMeta: svc.GetMeta(),
					origin:       origin,
				}
				svc.SetMeta(overridden)

				if err := generateAndRecord(svc, svcType, svcZone, hg, byHostname); err != nil {
					rest_errors.HandleError(request.Request.Context(), response, err, "could not generate hostname")
					return
				}
			}
		}

		resp := types.InspectHostnames{}
		for _, host := range util_maps.SortedKeys(byHostname) {
			zoneSet := byHostname[host]
			zones := make([]types.InspectHostnameZone, 0, len(zoneSet))
			for zoneName := range zoneSet {
				zones = append(zones, types.InspectHostnameZone{
					Name: zoneName,
				})
			}
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
