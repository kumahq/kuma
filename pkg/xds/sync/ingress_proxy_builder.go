package sync

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_cache "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type IngressProxyBuilder struct {
	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	meshCache          *xds_cache.Cache

	apiVersion core_xds.APIVersion
	zone       string
}

func (p *IngressProxyBuilder) Build(ctx context.Context, key core_model.ResourceKey) (*core_xds.Proxy, error) {
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

	zoneIngressProxy, err := p.buildZoneIngressProxy(ctx, zoneIngress, zoneEgressesList)
	if err != nil {
		return nil, err
	}

	proxy := &core_xds.Proxy{
		Id:               core_xds.FromResourceKey(key),
		APIVersion:       p.apiVersion,
		ZoneIngressProxy: zoneIngressProxy,
	}
	return proxy, nil
}

func (p *IngressProxyBuilder) buildZoneIngressProxy(
	ctx context.Context,
	zoneIngress *core_mesh.ZoneIngressResource,
	zoneEgressesList *core_mesh.ZoneEgressResourceList,
) (*core_xds.ZoneIngressProxy, error) {
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

	meshHTTPRoutes := &meshhttproute_api.MeshHTTPRouteResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, meshHTTPRoutes); err != nil {
		return nil, err
	}

	var meshList core_mesh.MeshResourceList
	if err := p.ReadOnlyResManager.List(ctx, &meshList); err != nil {
		return nil, err
	}

	var meshResourceList []*core_xds.MeshIngressResources

	for _, mesh := range meshList.Items {
		meshName := mesh.GetMeta().GetName()

		meshCtx, err := p.meshCache.GetMeshContext(ctx, meshName)
		if err != nil {
			return nil, err
		}

		destinations := ingress.BuildDestinationMap(meshName, zoneIngress)

		meshResources := &core_xds.MeshIngressResources{
			Mesh: mesh,
			EndpointMap: ingress.BuildEndpointMap(
				destinations,
				meshCtx.Resources.Dataplanes().Items,
				meshCtx.Resources.ExternalServices().Items,
				zoneEgressesList.Items,
				gateways.Items,
			),
		}

		meshResourceList = append(meshResourceList, meshResources)
	}

	return &core_xds.ZoneIngressProxy{
		ZoneIngressResource: zoneIngress,
		GatewayRoutes:       gatewayRoutes,
		MeshGateways:        gateways,
		PolicyResources: map[core_model.ResourceType]core_model.ResourceList{
			meshhttproute_api.MeshHTTPRouteType: meshHTTPRoutes,
			core_mesh.TrafficRouteType:          routes,
		},
		MeshResourceList: meshResourceList,
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
	meshList := &core_mesh.MeshResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, meshList, store.ListOrdered()); err != nil {
		return nil, err
	}

	allMeshExternalServices := &core_mesh.ExternalServiceResourceList{}
	var externalServices []*core_mesh.ExternalServiceResource
	for _, mesh := range meshList.Items {
		if !mesh.ZoneEgressEnabled() {
			continue
		}
		meshName := mesh.GetMeta().GetName()

		meshCtx, err := p.meshCache.GetMeshContext(ctx, meshName)
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

	allMeshExternalServices.Items = externalServices
	return allMeshExternalServices, nil
}
