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

func (p *ingressProxyBuilder) Build(key core_model.ResourceKey, streamId int64) (*xds.Proxy, error) {
	ctx := context.Background()
	dataplane := core_mesh.NewDataplaneResource()
	proxyID := xds.FromResourceKey(key)

	if err := p.ReadOnlyResManager.Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
		return nil, err
	}

	resolvedDp, err := xds_topology.ResolveAddress(p.LookupIP, dataplane)
	if err != nil {
		return nil, err
	}
	dataplane = resolvedDp

	// update Ingress
	allMeshDataplanes := &core_mesh.DataplaneResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, allMeshDataplanes); err != nil {
		return nil, err
	}
	allMeshDataplanes.Items = xds_topology.ResolveAddresses(syncLog, p.LookupIP, allMeshDataplanes.Items)
	if err := ingress.UpdateAvailableServices(ctx, p.ResManager, dataplane, allMeshDataplanes.Items); err != nil {
		return nil, err
	}
	destinations := ingress.BuildDestinationMap(dataplane)
	endpoints := ingress.BuildEndpointMap(destinations, allMeshDataplanes.Items)

	routes := &core_mesh.TrafficRouteResourceList{}
	if err := p.ReadOnlyResManager.List(ctx, routes); err != nil {
		return nil, err
	}

	proxy := xds.Proxy{
		Id:               proxyID,
		Dataplane:        dataplane,
		OutboundTargets:  endpoints,
		TrafficRouteList: routes,
		Metadata:         p.MetadataTracker.Metadata(streamId),
	}
	return &proxy, nil
}
