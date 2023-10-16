package sync

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type IngressProxyBuilder struct {
	ResManager manager.ResourceManager
	LookupIP   lookup.LookupIPFunc

	apiVersion        core_xds.APIVersion
	zone              string
	ingressTagFilters []string
}

func (p *IngressProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) (*core_xds.Proxy, error) {
	zoneIngress, err := p.getZoneIngress(ctx, key, aggregatedMeshCtxs)
	if err != nil {
		return nil, err
	}

	zoneIngress, err = xds_topology.ResolveZoneIngressPublicAddress(p.LookupIP, zoneIngress)
	if err != nil {
		return nil, err
	}

	proxy := &core_xds.Proxy{
		Id:               core_xds.FromResourceKey(key),
		APIVersion:       p.apiVersion,
		Zone:             p.zone,
		ZoneIngressProxy: p.buildZoneIngressProxy(zoneIngress, aggregatedMeshCtxs),
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
	zoneEgressesList := aggregatedMeshCtxs.ZoneEgresses()
	var meshResourceList []*core_xds.MeshIngressResources

	for _, mesh := range aggregatedMeshCtxs.Meshes {
		meshName := mesh.GetMeta().GetName()
		meshCtx := aggregatedMeshCtxs.MustGetMeshContext(meshName)

		meshResources := &core_xds.MeshIngressResources{
			Mesh: mesh,
			EndpointMap: ingress.BuildEndpointMap(
				ingress.BuildDestinationMap(meshName, zoneIngress),
				meshCtx.Resources.Dataplanes().Items,
				meshCtx.Resources.ExternalServices().Items,
				zoneEgressesList,
				meshCtx.Resources.Gateways().Items,
			),
			Resources: meshCtx.Resources.MeshLocalResources,
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
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) (*core_mesh.ZoneIngressResource, error) {
	zoneIngress := core_mesh.NewZoneIngressResource()
	if err := p.ResManager.Get(ctx, zoneIngress, core_store.GetBy(key)); err != nil {
		return nil, err
	}
	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	if err := p.updateIngress(ctx, zoneIngress, aggregatedMeshCtxs); err != nil {
		return nil, err
	}
	return zoneIngress, nil
}

func (p *IngressProxyBuilder) updateIngress(
	ctx context.Context, zoneIngress *core_mesh.ZoneIngressResource,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) error {
	// Update Ingress' Available Services
	// This was placed as an operation of DataplaneWatchdog out of the convenience.
	// Consider moving to the outside of this component (follow the pattern of updating VIP outbounds)
	return ingress.UpdateAvailableServices(
		ctx,
		p.ResManager,
		zoneIngress,
		aggregatedMeshCtxs.AllDataplanes(),
		aggregatedMeshCtxs.AllMeshGateways(),
		p.getIngressExternalServices(aggregatedMeshCtxs).Items,
		p.ingressTagFilters,
	)
}

func (p *IngressProxyBuilder) getIngressExternalServices(
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) *core_mesh.ExternalServiceResourceList {
	allMeshExternalServices := &core_mesh.ExternalServiceResourceList{}
	var externalServices []*core_mesh.ExternalServiceResource
	for _, mesh := range aggregatedMeshCtxs.Meshes {
		if !mesh.ZoneEgressEnabled() {
			continue
		}

		meshCtx := aggregatedMeshCtxs.MustGetMeshContext(mesh.GetMeta().GetName())

		// look for external services that are only available in my zone and expose them
		for _, es := range meshCtx.Resources.ExternalServices().Items {
			if es.Spec.Tags[mesh_proto.ZoneTag] == p.zone {
				externalServices = append(externalServices, es)
			}
		}
	}

	allMeshExternalServices.Items = externalServices
	return allMeshExternalServices
}
