package sync

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type IngressProxyBuilder struct {
	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	MetadataTracker    DataplaneMetadataTracker

	apiVersion envoy.APIVersion
}

func (p *IngressProxyBuilder) build(key core_model.ResourceKey) (*xds.Proxy, error) {
	ctx := context.Background()

	zoneIngress, err := p.getZoneIngress(key)
	if err != nil {
		return nil, err
	}
	zoneIngress, err = xds_topology.ResolveZoneIngressPublicAddress(p.LookupIP, zoneIngress)
	if err != nil {
		return nil, err
	}

	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return nil, err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	routing, err := p.resolveRouting(ctx, zoneIngress, allMeshDataplanes)
	if err != nil {
		return nil, err
	}

	proxy := &xds.Proxy{
		Id:          xds.FromResourceKey(key),
		APIVersion:  p.apiVersion,
		ZoneIngress: zoneIngress,
		Metadata:    p.MetadataTracker.Metadata(key),
		Routing:     *routing,
	}
	return proxy, nil
}

func (p *IngressProxyBuilder) getZoneIngress(key core_model.ResourceKey) (*core_mesh.ZoneIngressResource, error) {
	ctx := context.Background()

	zoneIngress := core_mesh.NewZoneIngressResource()
	if err := p.ReadOnlyResManager.Get(ctx, zoneIngress, core_store.GetBy(key)); err != nil {
		return nil, err
	}
	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	if err := p.updateIngress(zoneIngress); err != nil {
		return nil, err
	}
	return zoneIngress, nil
}

func (p *IngressProxyBuilder) resolveRouting(ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource, dataplanes *core_mesh.DataplaneResourceList) (*xds.Routing, error) {
	destinations := ingress.BuildDestinationMap(zoneIngress)
	endpoints := ingress.BuildEndpointMap(destinations, dataplanes.Items)
	routes := &core_mesh.TrafficRouteResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, routes); err != nil {
		return nil, err
	}

	routing := &xds.Routing{
		OutboundTargets:  endpoints,
		TrafficRouteList: routes,
	}
	return routing, nil
}

func (p *IngressProxyBuilder) updateIngress(zoneIngress *core_mesh.ZoneIngressResource) error {
	ctx := context.Background()

	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	return ingress.UpdateAvailableServices(ctx, p.ResManager, zoneIngress, allMeshDataplanes.Items)
}
