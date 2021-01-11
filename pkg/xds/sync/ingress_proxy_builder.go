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

func (p *IngressProxyBuilder) build(key core_model.ResourceKey, streamId int64) (*xds.Proxy, error) {
	ctx := context.Background()

	dp, err := p.resolveDataplane(ctx, key)
	if err != nil {
		return nil, err
	}

	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return nil, err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	if err := ingress.UpdateAvailableServices(ctx, p.ResManager, dp, allMeshDataplanes.Items); err != nil {
		return nil, err
	}

	routing, err := p.resolveRouting(ctx, dp, allMeshDataplanes)
	if err != nil {
		return nil, err
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		Dataplane:  dp,
		Metadata:   p.MetadataTracker.Metadata(streamId),
		Routing:    *routing,
	}
	return proxy, nil
}

func (p *IngressProxyBuilder) resolveDataplane(ctx context.Context, key core_model.ResourceKey) (*core_mesh.DataplaneResource, error) {
	dataplane := core_mesh.NewDataplaneResource()

	if err := p.ReadOnlyResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return nil, err
	}

	// Envoy requires IPs instead of Hostname therefore we need to resolve an address. Consider moving this outside of this component.
	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return nil, err
	}
	return resolvedDp, nil
}

func (p *IngressProxyBuilder) resolveRouting(ctx context.Context, dataplane *core_mesh.DataplaneResource, dataplanes *core_mesh.DataplaneResourceList) (*xds.Routing, error) {
	destinations := ingress.BuildDestinationMap(dataplane)
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
