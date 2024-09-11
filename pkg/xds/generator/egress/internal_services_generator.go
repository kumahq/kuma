package egress

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

type InternalServicesGenerator struct{}

// Generate will generate envoy resources for one mesh (when mTLS enabled)
func (g *InternalServicesGenerator) Generate(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshResources *core_xds.MeshResources,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	meshName := meshResources.Mesh.GetMeta().GetName()

	servicesMap := g.buildServices(meshResources.EndpointMap, meshResources.Mesh.ZoneEgressEnabled(), xdsCtx.ControlPlane.Zone)

	availableServices := g.distinctAvailableServices(proxy.ZoneEgressProxy.ZoneIngresses, meshName, servicesMap)

	destinations := zoneproxy.BuildMeshDestinations(
		availableServices,
		xds_context.Resources{MeshLocalResources: meshResources.Resources},
		nil, // todo(jakubdyszkiewicz) add support for MeshService + egress
		nil, // todo(jakubdyszkiewicz) add support for MeshService + egress
		nil,
		"",
	)

	services := zoneproxy.AddFilterChains(availableServices, proxy.APIVersion, listenerBuilder, destinations, meshResources.EndpointMap)

	cds, err := zoneproxy.GenerateCDS(destinations, services, proxy.APIVersion, meshName, OriginEgress)
	if err != nil {
		return nil, err
	}
	resources.Add(cds...)

	eds, err := zoneproxy.GenerateEDS(services, meshResources.EndpointMap, proxy.APIVersion, meshName, OriginEgress)
	if err != nil {
		return nil, err
	}
	resources.Add(eds...)

	return resources, nil
}

func (*InternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
	zoneEgressEnabled bool,
	localZone string,
) map[string]bool {
	services := map[string]bool{}

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) == 0 {
			continue
		}
		internalService := !endpoints[0].IsExternalService()
		zoneExternalService := zoneEgressEnabled && endpoints[0].IsExternalService() && !endpoints[0].IsReachableFromZone(localZone)
		if internalService || zoneExternalService {
			services[serviceName] = true
		}
	}

	return services
}

func (*InternalServicesGenerator) distinctAvailableServices(
	zoneIngresses []*mesh.ZoneIngressResource,
	meshName string,
	services map[string]bool,
) []*mesh_proto.ZoneIngress_AvailableService {
	var result []*mesh_proto.ZoneIngress_AvailableService
	distinct := map[string]struct{}{}
	for _, zoneIngress := range zoneIngresses {
		for _, service := range zoneIngress.Spec.GetAvailableServices() {
			serviceName := service.Tags[mesh_proto.ServiceTag]
			if service.Mesh == meshName && services[serviceName] {
				tagsString := tags.Tags(service.Tags).String()
				if _, ok := distinct[tagsString]; !ok {
					distinct[tagsString] = struct{}{}
					result = append(result, service)
				}
			}
		}
	}
	return result
}
