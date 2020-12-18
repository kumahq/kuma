package sync

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type ingressProxyBuilder struct {
	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	MetadataTracker    DataplaneMetadataTracker
}

func (p *ingressProxyBuilder) build(key core_model.ResourceKey, streamId int64) (*xds.Proxy, error) {
	ctx := context.Background()
	proxy := &xds.Proxy{
		Id: xds.FromResourceKey(key),
		Metadata: p.MetadataTracker.Metadata(streamId),
	}

	if err := p.resolveDataplane(ctx, key, proxy); err != nil {
		return nil, err
	}

	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return nil, err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)

	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience. Consider moving to the outside component (follow the pattern of updating VIP outbounds)
	if err := ingress.UpdateAvailableServices(ctx, p.ResManager, proxy.Dataplane, allMeshDataplanes.Items); err != nil {
		return nil, err
	}

	if err := p.resolveRouting(ctx, proxy, allMeshDataplanes); err != nil {
		return nil, err
	}
	return proxy, nil
}

func (p *ingressProxyBuilder) resolveDataplane(ctx context.Context, key core_model.ResourceKey, proxy *xds.Proxy) error {
	dataplane := core_mesh.NewDataplaneResource()

	if err := p.ReadOnlyResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return err
	}

	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return err
	}
	dataplane = resolvedDp

	proxy.Dataplane = resolvedDp
	return nil
}

func (p *ingressProxyBuilder) resolveRouting(ctx context.Context, proxy *xds.Proxy, dataplanes *core_mesh.DataplaneResourceList) error {
	destinations := ingress.BuildDestinationMap(proxy.Dataplane)

	endpoints := ingress.BuildEndpointMap(destinations, dataplanes.Items)
	proxy.OutboundTargets = endpoints

	routes := &core_mesh.TrafficRouteResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, routes); err != nil {
		return err
	}
	proxy.TrafficRouteList = routes

	return nil
}
