package sync

import (
	"context"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_cache "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type IngressProxyBuilder struct {
	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	MetadataTracker    DataplaneMetadataTracker
	meshCache          *xds_cache.Cache

	apiVersion envoy.APIVersion
	zone       string
}

func (p *IngressProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey) (*xds.Proxy, error) {
	zoneIngress, err := p.getZoneIngress(ctx, key)
	if err != nil {
		return nil, err
	}

	zoneIngress, err = xds_topology.ResolveZoneIngressPublicAddress(p.LookupIP, zoneIngress)
	if err != nil {
		return nil, err
	}

	zoneEgressesList := &core_mesh.ZoneEgressResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, zoneEgressesList); err != nil {
		return nil, err
	}

	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return nil, err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	availableExternalServices, err := p.getIngressExternalServices(ctx)
	if err != nil {
		return nil, err
	}

	zoneIngressProxy, err := p.buildZoneIngressProxy(ctx)
	if err != nil {
		return nil, err
	}

	routing := p.resolveRouting(zoneIngress, zoneEgressesList, allMeshDataplanes, availableExternalServices, zoneIngressProxy.MeshGateways)

	proxy := &xds.Proxy{
		Id:               xds.FromResourceKey(key),
		APIVersion:       p.apiVersion,
		ZoneIngress:      zoneIngress,
		Metadata:         p.MetadataTracker.Metadata(key),
		Routing:          *routing,
		ZoneIngressProxy: zoneIngressProxy,
	}
	return proxy, nil
}

func (p *IngressProxyBuilder) buildZoneIngressProxy(ctx context.Context) (*xds.ZoneIngressProxy, error) {
	routes := &core_mesh.TrafficRouteResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, routes); err != nil {
		return nil, err
	}

	gatewayRoutes := &core_mesh.MeshGatewayRouteResourceList{}
	if _, err := registry.Global().DescriptorFor(core_mesh.MeshGatewayRouteType); err == nil { // GatewayRoute may not be registered
		if err := p.ReadOnlyResManager.List(ctx, gatewayRoutes); err != nil {
			return nil, err
		}
	}

	gateways := &core_mesh.MeshGatewayResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, gateways); err != nil {
		return nil, err
	}

	return &xds.ZoneIngressProxy{
		TrafficRouteList: routes,
		GatewayRoutes:    gatewayRoutes,
		MeshGateways:     gateways,
	}, nil
}

func (p *IngressProxyBuilder) getZoneIngress(ctx context.Context, key core_model.ResourceKey) (*core_mesh.ZoneIngressResource, error) {
	zoneIngress := core_mesh.NewZoneIngressResource()
	if err := p.ReadOnlyResManager.Get(ctx, zoneIngress, core_store.GetBy(key)); err != nil {
		return nil, err
	}
	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	if err := p.updateIngress(ctx, zoneIngress); err != nil {
		return nil, err
	}
	return zoneIngress, nil
}

func (p *IngressProxyBuilder) resolveRouting(
	zoneIngress *core_mesh.ZoneIngressResource,
	zoneEgresses *core_mesh.ZoneEgressResourceList,
	dataplanes *core_mesh.DataplaneResourceList,
	externalServices *core_mesh.ExternalServiceResourceList,
	meshGateways *core_mesh.MeshGatewayResourceList,
) *xds.Routing {
	destinations := ingress.BuildDestinationMap(zoneIngress)
	endpoints := ingress.BuildEndpointMap(
		destinations, dataplanes.Items, externalServices.Items, zoneEgresses.Items, meshGateways.Items,
	)

	routing := &xds.Routing{
		OutboundTargets: endpoints,
	}
	return routing
}

func (p *IngressProxyBuilder) updateIngress(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource) error {
	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return err
	}
	allMeshGateways := &core_mesh.MeshGatewayResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshGateways); err != nil {
		return err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	availableExternalServices, err := p.getIngressExternalServices(ctx)
	if err != nil {
		return err
	}

	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	return ingress.UpdateAvailableServices(ctx, p.ResManager, zoneIngress, allMeshDataplanes.Items, allMeshGateways.Items, availableExternalServices.Items)
}

func (p *IngressProxyBuilder) getIngressExternalServices(ctx context.Context) (*core_mesh.ExternalServiceResourceList, error) {
	var meshList core_mesh.MeshResourceList
	if err := p.ReadOnlyResManager.List(ctx, &meshList); err != nil {
		return nil, err
	}

	var meshes []*core_mesh.MeshResource

	for _, mesh := range meshList.Items {
		if mesh.ZoneEgressEnabled() {
			meshes = append(meshes, mesh)
		}
	}

	allMeshExternalServices := &core_mesh.ExternalServiceResourceList{}
	var externalServices []*core_mesh.ExternalServiceResource
	for _, mesh := range meshes {
		meshName := mesh.GetMeta().GetName()

		meshCtx, err := p.meshCache.GetMeshContext(ctx, syncLog, meshName)
		if err != nil {
			return nil, err
		}

		meshExternalServices := meshCtx.Resources.ExternalServices().Items

		// look for external services that are only available in my zone and expose them
		for _, es := range meshExternalServices {
			if es.Spec.Tags[mesh_proto.ZoneTag] == p.zone {
				externalServices = append(externalServices, es)
			}
		}
	}

	// It's done for achieving stable xds config
	sort.Slice(externalServices, func(a, b int) bool {
		return externalServices[a].GetMeta().GetName() < externalServices[b].GetMeta().GetName()
	})

	allMeshExternalServices.Items = externalServices
	return allMeshExternalServices, nil
}
