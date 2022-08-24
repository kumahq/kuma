package sync

import (
	"context"
	"sort"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_cache "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type EgressProxyBuilder struct {
	ctx context.Context

	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	MetadataTracker    DataplaneMetadataTracker
	meshCache          *xds_cache.Cache

	zone       string
	apiVersion envoy_common.APIVersion
}

func (p *EgressProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
) (*xds.Proxy, error) {
	zoneEgress := core_mesh.NewZoneEgressResource()

	if err := p.ReadOnlyResManager.Get(
		ctx,
		zoneEgress,
		core_store.GetBy(key),
	); err != nil {
		return nil, err
	}

	var meshList core_mesh.MeshResourceList
	if err := p.ReadOnlyResManager.List(ctx, &meshList); err != nil {
		return nil, err
	}

	// As egress is using SNI to identify the services, we need to filter out
	// meshes with no mTLS enabled and with ZoneEgress enabled
	var meshes []*core_mesh.MeshResource
	for _, mesh := range meshList.Items {
		if mesh.ZoneEgressEnabled() {
			meshes = append(meshes, mesh)
		}
	}

	var zoneIngressesList core_mesh.ZoneIngressResourceList
	if err := p.ReadOnlyResManager.List(ctx, &zoneIngressesList); err != nil {
		return nil, err
	}

	// We don't want to process services from our local zone ingress
	var zoneIngresses []*core_mesh.ZoneIngressResource
	for _, zoneIngress := range zoneIngressesList.Items {
		if zoneIngress.IsRemoteIngress(p.zone) {
			zoneIngresses = append(zoneIngresses, zoneIngress)
		}
	}

	// Resolve hostnames to ips in zoneIngresses.
	zoneIngresses = xds_topology.ResolveZoneIngressAddresses(xdsServerLog, p.LookupIP, zoneIngresses)

	// It's done for achieving stable xds config
	sort.Slice(zoneIngresses, func(a, b int) bool {
		return zoneIngresses[a].GetMeta().GetName() < zoneIngresses[b].GetMeta().GetName()
	})

	var meshResourcesList []*xds.MeshResources

	for _, mesh := range meshes {
		meshName := mesh.GetMeta().GetName()

		meshCtx, err := p.meshCache.GetMeshContext(ctx, syncLog, meshName)
		if err != nil {
			return nil, err
		}

		trafficPermissions := meshCtx.Resources.TrafficPermissions().Items
		trafficRoutes := meshCtx.Resources.TrafficRoutes().Items
		externalServices := meshCtx.Resources.ExternalServices().Items
		faultInjections := meshCtx.Resources.FaultInjections().Items
		rateLimits := meshCtx.Resources.RateLimits().Items

		// It's done for achieving stable xds config
		sort.Slice(externalServices, func(a, b int) bool {
			return externalServices[a].GetMeta().GetName() < externalServices[b].GetMeta().GetName()
		})

		meshResources := &xds.MeshResources{
			Mesh:             mesh,
			TrafficRoutes:    trafficRoutes,
			ExternalServices: externalServices,
			EndpointMap: xds_topology.BuildRemoteEndpointMap(
				ctx,
				mesh,
				p.zone,
				zoneIngresses,
				externalServices,
				meshCtx.DataSourceLoader,
			),
			ExternalServicePermissionMap: permissions.BuildExternalServicesPermissionsMapForZoneEgress(
				externalServices,
				trafficPermissions,
			),
			ExternalServiceFaultInjections: faultinjections.BuildExternalServiceFaultInjectionMapForZoneEgress(
				externalServices,
				faultInjections,
			),
			ExternalServiceRateLimits: ratelimits.BuildExternalServiceRateLimitMapForZoneEgress(
				externalServices,
				rateLimits,
			),
		}

		meshResourcesList = append(meshResourcesList, meshResources)
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		ZoneEgressProxy: &xds.ZoneEgressProxy{
			ZoneEgressResource: zoneEgress,
			ZoneIngresses:      zoneIngresses,
			MeshResourcesList:  meshResourcesList,
		},
		Metadata: p.MetadataTracker.Metadata(key),
	}

	return proxy, nil
}
