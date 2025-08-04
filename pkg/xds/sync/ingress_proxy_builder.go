package sync

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type IngressProxyBuilder struct {
	ResManager manager.ResourceManager
	LookupIP   lookup.LookupIPFunc

	apiVersion        core_xds.APIVersion
	InternalAddresses []core_xds.InternalAddress
	zone              string
	ingressTagFilters []string
}

func (p *IngressProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) (*core_xds.Proxy, error) {
	zoneIngress, err := p.getZoneIngress(ctx, key)
	if err != nil {
		return nil, err
	}

	zoneIngress, err = xds_topology.ResolveZoneIngressPublicAddress(p.LookupIP, zoneIngress)
	if err != nil {
		return nil, err
	}

	proxy := &core_xds.Proxy{
		Id:                core_xds.FromResourceKey(key),
		APIVersion:        p.apiVersion,
		InternalAddresses: p.InternalAddresses,
		Zone:              p.zone,
		ZoneIngressProxy:  p.buildZoneIngressProxy(zoneIngress, aggregatedMeshCtxs),
	}
	for k, pl := range core_plugins.Plugins().ProxyPlugins() {
		err := pl.Apply(ctx, xds_context.MeshContext{}, proxy) // No mesh context for zone proxies
		if err != nil {
			return nil, errors.Wrapf(err, "Failed applying proxy plugin: %s", k)
		}
	}
	return proxy, nil
}

func (p *IngressProxyBuilder) buildZoneIngressProxy(
	zoneIngress *core_mesh.ZoneIngressResource,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) *core_xds.ZoneIngressProxy {
	var meshResourceList []*core_xds.MeshIngressResources

	for _, mesh := range aggregatedMeshCtxs.Meshes {
		meshName := mesh.GetMeta().GetName()
		meshCtx := aggregatedMeshCtxs.MustGetMeshContext(meshName)

		meshResources := &core_xds.MeshIngressResources{
			Mesh:        mesh,
			EndpointMap: meshCtx.IngressEndpointMap,
			Resources:   meshCtx.Resources.MeshLocalResources,
		}

		meshResourceList = append(meshResourceList, meshResources)
	}

	return &core_xds.ZoneIngressProxy{
		ZoneIngressResource: zoneIngress,
		MeshResourceList:    meshResourceList,
	}
}

func (p *IngressProxyBuilder) getZoneIngress(
	ctx context.Context,
	key core_model.ResourceKey,
) (*core_mesh.ZoneIngressResource, error) {
	zoneIngress := core_mesh.NewZoneIngressResource()
	if err := p.ResManager.Get(ctx, zoneIngress, core_store.GetBy(key)); err != nil {
		return nil, err
	}
	return zoneIngress, nil
}
