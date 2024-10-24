package generator

import (
	"context"
	"sort"

	envoy_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

const (
	IngressProxy = "ingress-proxy"

	// OriginIngress is a marker to indicate by which ProxyGenerator resources
	// were generated.
	OriginIngress = "ingress"
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
		serviceList := maps.Keys(mr.EndpointMap)
		sort.Strings(serviceList)

		meshResources := xds_context.Resources{MeshLocalResources: mr.Resources}
		// we only want to expose local mesh services
		localMs := localMeshServices(xdsCtx.ControlPlane.Zone, meshResources.MeshServices().Items)
		localMes := localMeshExternalServices(xdsCtx.ControlPlane.Zone, meshResources.MeshExternalServices().Items)
		dest := zoneproxy.BuildMeshDestinations(
			availableSvcsByMesh[meshName],
			meshResources,
			localMs,
			localMes,
			meshResources.MeshMultiZoneServices().Items,
			xdsCtx.ControlPlane.SystemNamespace,
			xdsCtx.Mesh.ResolveResourceIdentifier,
		)

		services := zoneproxy.AddFilterChains(availableSvcsByMesh[meshName], proxy.APIVersion, listenerBuilder, dest, mr.EndpointMap)

		cdsResources, err := zoneproxy.GenerateCDS(dest, services, proxy.APIVersion, meshName, OriginIngress)
		if err != nil {
			return nil, err
		}
		resources.Add(cdsResources...)

		edsResources, err := zoneproxy.GenerateEDS(services, mr.EndpointMap, proxy.APIVersion, meshName, OriginIngress)
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
		Origin:   OriginIngress,
		Resource: listener,
	})
	return resources, nil
}

func localMeshServices(zone string, meshServices []*meshservice_api.MeshServiceResource) []*meshservice_api.MeshServiceResource {
	var result []*meshservice_api.MeshServiceResource
	for _, ms := range meshServices {
		if labels := ms.GetMeta().GetLabels(); labels != nil && labels[mesh_proto.ZoneTag] != "" && labels[mesh_proto.ZoneTag] != zone {
			continue
		}
		result = append(result, ms)
	}
	return result
}

func localMeshExternalServices(zone string, meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource) []*meshexternalservice_api.MeshExternalServiceResource {
	var result []*meshexternalservice_api.MeshExternalServiceResource
	for _, ms := range meshExternalServices {
		if labels := ms.GetMeta().GetLabels(); labels != nil && labels[mesh_proto.ZoneTag] != "" && labels[mesh_proto.ZoneTag] != zone {
			continue
		}
		result = append(result, ms)
	}
	return result
}
