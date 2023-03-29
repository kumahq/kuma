package sync

import (
	"context"
	"sort"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_cache "github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type EgressProxyBuilder struct {
	ctx context.Context

	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	meshCache          *xds_cache.Cache

	zone       string
	apiVersion core_xds.APIVersion
}

func (p *EgressProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
) (*core_xds.Proxy, error) {
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

	var meshResourcesList []*core_xds.MeshResources

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

		meshResources := &core_xds.MeshResources{
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
			Dynamic: core_xds.ExternalServiceDynamicPolicies{},
		}

		for _, es := range externalServices {
			policies := core_xds.PluginOriginatedPolicies{}
			for name, plugin := range plugins.Plugins().PolicyPlugins() {
				egressPlugin, ok := plugin.(plugins.EgressPolicyPlugin)
				if !ok {
					continue
				}
				res, err := egressPlugin.EgressMatchedPolicies(es, meshCtx.Resources)
				if err != nil {
					return nil, errors.Wrapf(err, "could not apply policy plugin %s", name)
				}
				if res.Type == "" {
					return nil, errors.Wrapf(err, "matched policy didn't set type for policy plugin %s", name)
				}
				policies[res.Type] = res
			}
			meshResources.Dynamic[es.Spec.GetService()] = policies
		}

		meshResourcesList = append(meshResourcesList, meshResources)
	}

	proxy := &core_xds.Proxy{
		Id:         core_xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		ZoneEgressProxy: &core_xds.ZoneEgressProxy{
			ZoneEgressResource: zoneEgress,
			ZoneIngresses:      zoneIngresses,
			MeshResourcesList:  meshResourcesList,
		},
	}

	return proxy, nil
}
