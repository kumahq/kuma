package egress

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

type InternalServicesGenerator struct{}

// relevantTags is a set of tags for which it actually makes sense to do LB split on.
// If the endpoint list is the same with or without the tag, we should just not do the split.
// However, we should preserve full SNI, because the client expects Zone Proxy to support it.
// This solves the problem that Envoy deduplicate endpoints of the same address and different metadata.
// example 1:
// Ingress1 (10.0.0.1) supports service:a,version:1 and service:a,version:2
// Ingress2 (10.0.0.2) supports service:a,version:1 and service:a,version:2
// If we want to split by version, we don't need to do LB subset on version.
//
// example 2:
// Ingress1 (10.0.0.1) supports service:a,version:1
// Ingress2 (10.0.0.2) supports service:a,version:2
// If we want to split by version, we need LB subset.
func relevantTags(
	destination tags.Tags, serviceEndpoints []core_xds.Endpoint,
) tags.Tags {
	relevantTags := tags.Tags{}
	for key, value := range destination {
		matchedTargets := map[string]struct{}{}
		allTargets := map[string]struct{}{}
		for _, endpoint := range serviceEndpoints {
			address := endpoint.Address()
			if endpoint.Tags[key] == value || value == mesh_proto.MatchAllTag {
				matchedTargets[address] = struct{}{}
			}
			allTargets[address] = struct{}{}
		}
		if len(matchedTargets) < len(allTargets) {
			relevantTags[key] = value
		}
	}

	return relevantTags
}

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
	services := util_maps.SortedKeys(servicesMap)

	availableServices := g.distinctAvailableServices(proxy.ZoneEgressProxy.ZoneIngresses, meshName, servicesMap)

	destinations := zoneproxy.BuildMeshDestinations(
		availableServices,
		xds_context.Resources{MeshLocalResources: meshResources.Resources},
	)

	zoneproxy.AddFilterChains(availableServices, proxy.APIVersion, listenerBuilder, destinations, meshResources.EndpointMap, relevantTags)

	cds, err := zoneproxy.GenerateCDS(services, destinations, proxy.APIVersion, meshName, OriginEgress)
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
