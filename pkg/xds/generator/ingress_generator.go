package generator

import (
	"context"

	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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
	resources := core_xds.NewResourceSet()

	networking := proxy.ZoneIngressProxy.ZoneIngressResource.Spec.GetNetworking()
	address, port := networking.GetAddress(), networking.GetPort()
	listenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, address, port, core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TLSInspector())

	availableSvcsByMesh := map[string][]*mesh_proto.ZoneIngress_AvailableService{}
	for _, service := range proxy.ZoneIngressProxy.ZoneIngressResource.Spec.AvailableServices {
		availableSvcsByMesh[service.Mesh] = append(availableSvcsByMesh[service.Mesh], service)
	}

	for _, mr := range proxy.ZoneIngressProxy.MeshResourceList {
		meshName := mr.Mesh.GetMeta().GetName()

		meshResources := xds_context.Resources{MeshLocalResources: mr.Resources}
		// we only want to expose local mesh services
		localMS := &meshservice_api.MeshServiceResourceList{}
		for _, ms := range meshResources.MeshServices().GetItems() {
			if labels := ms.GetMeta().GetLabels(); labels == nil || labels[mesh_proto.ZoneTag] == "" || labels[mesh_proto.ZoneTag] == xdsCtx.ControlPlane.Zone {
				_ = localMS.AddItem(ms)
			}
		}

		dest := zoneproxy.BuildMeshDestinations(
			availableSvcsByMesh[meshName],
			xdsCtx.ControlPlane.SystemNamespace,
			meshResources,
			localMS,
			meshResources.MeshMultiZoneServices(),
		)

		services := zoneproxy.AddFilterChains(availableSvcsByMesh[meshName], proxy.APIVersion, listenerBuilder, dest, mr.EndpointMap)

		cdsResources, err := zoneproxy.GenerateCDS(dest, services, proxy.APIVersion, meshName, metadata.OriginIngress)
		if err != nil {
			return nil, err
		}
		resources.Add(cdsResources...)

		edsResources, err := zoneproxy.GenerateEDS(services, mr.EndpointMap, proxy.APIVersion, meshName, metadata.OriginIngress)
		if err != nil {
			return nil, err
		}
		resources.Add(edsResources...)
	}

	listener, err := listenerBuilder.Build()
	if err != nil {
		return nil, err
	}

	hasFilterChain := len(listener.(*envoy_listener_v3.Listener).FilterChains) > 0
	if !hasFilterChain {
		listener, err = zoneproxy.GenerateEmptyDirectResponseListener(proxy, address, port)
		if err != nil {
			return nil, err
		}
	}

	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   metadata.OriginIngress,
		Resource: listener,
	})
	return resources, nil
}
