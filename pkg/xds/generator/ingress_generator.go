package generator

import (
	"context"
	"fmt"
	"slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/naming"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

type IngressGenerator struct{}

func (i IngressGenerator) Generate(
	_ context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()
	cp := xdsCtx.ControlPlane
	unifiedNaming := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)

	zoneIngress := proxy.ZoneIngressProxy.ZoneIngressResource
	address := zoneIngress.Spec.GetNetworking().GetAddress()
	port := zoneIngress.Spec.GetNetworking().GetPort()

	listenerName := getName(kri.From(zoneIngress).String(), envoy_names.GetInboundListenerName(address, port))
	statPrefix := getName(naming.MustContextualInboundName(zoneIngress, port), "")

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.TLSInspector())

	availableServices := map[string][]*mesh_proto.ZoneIngress_AvailableService{}
	for _, service := range zoneIngress.Spec.AvailableServices {
		availableServices[service.Mesh] = append(availableServices[service.Mesh], service)
	}

	var clusters []envoy_common.Cluster

	for _, mr := range proxy.ZoneIngressProxy.MeshResourceList {
		meshName := mr.Mesh.GetMeta().GetName()

		meshResources := xds_context.Resources{MeshLocalResources: mr.Resources}

		// we only want to expose local mesh services
		localMS := &meshservice_api.MeshServiceResourceList{}
		for _, ms := range meshResources.MeshServices().GetItems() {
			if labels := ms.GetMeta().GetLabels(); labels == nil || labels[mesh_proto.ZoneTag] == "" || labels[mesh_proto.ZoneTag] == cp.Zone {
				_ = localMS.AddItem(ms)
			}
		}

		dest := zoneproxy.BuildMeshDestinations(
			availableServices[meshName],
			cp.SystemNamespace,
			meshResources,
			localMS,
			meshResources.MeshMultiZoneServices(),
		)

		services := zoneproxy.GetServices(proxy, dest, mr.EndpointMap, availableServices[meshName])

		clusters = slices.Concat(clusters, services.Clusters())

		cds, err := zoneproxy.GenerateCDS(proxy, dest, services, meshName, metadata.OriginIngress)
		if err != nil {
			return nil, err
		}
		rs.AddSet(cds)

		eds, err := zoneproxy.GenerateEDS(proxy, mr.EndpointMap, services, meshName, metadata.OriginIngress)
		if err != nil {
			return nil, err
		}
		rs.AddSet(eds)
	}

	for _, cluster := range clusters {
		listener.Configure(envoy_listeners.FilterChain(zoneproxy.CreateFilterChain(proxy, cluster)))
	}

	if len(clusters) == 0 {
		response := fmt.Sprintf(`{"proxy":%q,"zone":%q}`, proxy.Id.String(), proxy.Zone)

		filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.NetworkDirectResponse(response))

		listener.Configure(envoy_listeners.FilterChain(filterChain))
	}

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}

	rs.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   metadata.OriginIngress,
		Resource: resource,
	})

	return rs, nil
}
