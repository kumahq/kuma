package sync

import (
	"context"
	"sort"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

type EgressProxyBuilder struct {
	zone       string
	apiVersion core_xds.APIVersion
}

func (p *EgressProxyBuilder) Build(
	ctx context.Context,
	key core_model.ResourceKey,
	aggregatedMeshCtxs xds_context.AggregatedMeshContexts,
) (*core_xds.Proxy, error) {
	zoneEgress, ok := aggregatedMeshCtxs.ZoneEgressByName[key.Name]
	if !ok {
		return nil, core_store.ErrorResourceNotFound(core_mesh.DataplaneType, key.Name, key.Mesh)
	}

	// As egress is using SNI to identify the services, we need to filter out
	// meshes with no mTLS enabled and with ZoneEgress enabled
	var meshes []*core_mesh.MeshResource
	for _, mesh := range aggregatedMeshCtxs.Meshes {
		if mesh.ZoneEgressEnabled() {
			meshes = append(meshes, mesh)
		}
	}

	// We don't want to process services from our local zone ingress
	var zoneIngresses []*core_mesh.ZoneIngressResource
	for _, zoneIngress := range aggregatedMeshCtxs.ZoneIngresses() {
		if zoneIngress.IsRemoteIngress(p.zone) {
			zoneIngresses = append(zoneIngresses, zoneIngress)
		}
	}

	// It's done for achieving stable xds config
	sort.Slice(zoneIngresses, func(a, b int) bool {
		return zoneIngresses[a].GetMeta().GetName() < zoneIngresses[b].GetMeta().GetName()
	})

	var meshResourcesList []*core_xds.MeshResources

	for _, mesh := range meshes {
		meshName := mesh.GetMeta().GetName()
		meshCtx := aggregatedMeshCtxs.MustGetMeshContext(meshName)

		trafficPermissions := meshCtx.Resources.TrafficPermissions().Items
		trafficRoutes := meshCtx.Resources.TrafficRoutes()
		externalServices := meshCtx.Resources.ExternalServices().Items
		faultInjections := meshCtx.Resources.FaultInjections().Items
		rateLimits := meshCtx.Resources.RateLimits().Items

		resourceMap := map[core_model.ResourceType]core_model.ResourceList{
			core_mesh.TrafficRouteType: trafficRoutes,
		}
		meshResources := &core_xds.MeshResources{
			Mesh:             mesh,
			Resources:        resourceMap,
			ExternalServices: externalServices,
			EndpointMap: xds_topology.BuildEgressEndpointMap(
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
					return nil, errors.Errorf("matched policy didn't set type for policy plugin %s", name)
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
