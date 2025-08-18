package egress

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

func genInternalResources(
	proxy *core_xds.Proxy,
	resources *core_xds.MeshResources,
	localZone string,
) (*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error) {
	if !resources.Mesh.ZoneEgressEnabled() {
		return nil, nil, nil
	}

	rs := core_xds.NewResourceSet()

	meshName := resources.Mesh.GetMeta().GetName()

	availableServicesMap := map[string]*mesh_proto.ZoneIngress_AvailableService{}
	for _, zi := range proxy.ZoneEgressProxy.ZoneIngresses {
		for _, svc := range zi.Spec.GetAvailableServices() {
			serviceName := svc.Tags[mesh_proto.ServiceTag]
			endpoints := resources.EndpointMap[serviceName]

			switch {
			case svc.Mesh != meshName:
				continue
			case len(endpoints) == 0:
				continue
			case endpoints[0].IsExternalService() && endpoints[0].IsReachableFromZone(localZone):
				continue
			}

			svcTags := tags.Tags(svc.Tags).String()
			if _, ok := availableServicesMap[svcTags]; !ok {
				availableServicesMap[svcTags] = svc
			}
		}
	}

	availableServices := util_maps.AllValues(availableServicesMap)

	destinations := zoneproxy.BuildMeshDestinations(
		availableServices,
		"",
		xds_context.Resources{MeshLocalResources: resources.Resources},
	)

	services := zoneproxy.GetServices(proxy, destinations, resources.EndpointMap, availableServices)

	cds, err := zoneproxy.GenerateCDS(proxy, destinations, services, meshName, metadata.OriginEgress)
	if err != nil {
		return nil, nil, err
	}
	rs.AddSet(cds)

	eds, err := zoneproxy.GenerateEDS(proxy, resources.EndpointMap, services, meshName, metadata.OriginEgress)
	if err != nil {
		return nil, nil, err
	}
	rs.AddSet(eds)

	var filterChainBuilders []*envoy_listeners.FilterChainBuilder
	for _, cluster := range services.Clusters() {
		filterChainBuilders = append(filterChainBuilders, zoneproxy.CreateFilterChain(proxy, cluster))
	}

	return rs, filterChainBuilders, nil
}
