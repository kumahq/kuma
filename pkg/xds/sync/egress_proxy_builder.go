package sync

import (
	"context"
	"sort"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/ordered"
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
		return nil, core_store.ErrorResourceNotFound(core_mesh.ZoneEgressType, key.Name, key.Mesh)
	}

	// As egress is using SNI to identify the services, we need to filter out meshes with no mTLS enabled
	// We don't check egress enabled to take into account MeshExternalServices
	var meshes []*core_mesh.MeshResource
	for _, mesh := range aggregatedMeshCtxs.Meshes {
		if mesh.MTLSEnabled() {
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
		externalServices := meshCtx.Resources.ExternalServices().Items
		faultInjections := meshCtx.Resources.FaultInjections().Items
		rateLimits := meshCtx.Resources.RateLimits().Items
		mes := meshCtx.Resources.MeshExternalServices().Items

		
		meshResources := &core_xds.MeshResources{
			Mesh:             mesh,
			ExternalServices: externalServices,
			EndpointMap: xds_topology.BuildEgressEndpointMap(
				ctx,
				mesh,
				p.zone,
				zoneIngresses,
				externalServices,
				mes,
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
			Dynamic:   core_xds.ExternalServiceDynamicPolicies{},
			Resources: meshCtx.Resources.MeshLocalResources,
		}

		core.Log.Info("EgressProxyBuilder", "mes", mes, "meshResources", meshResources)

		for _, es := range externalServices {
			policies, err := matchEgressPolicies(es.Spec.GetTags(), meshCtx.Resources)
			if err != nil {
				return nil, err
			}
			meshResources.Dynamic[es.Spec.GetService()] = policies
		}

		for serviceName := range meshResources.EndpointMap {
			policies, err := matchEgressPolicies(map[string]string{mesh_proto.ServiceTag: serviceName}, meshCtx.Resources)
			if err != nil {
				return nil, err
			}
			meshResources.Dynamic[serviceName] = policies
		}

		meshResourcesList = append(meshResourcesList, meshResources)
	}

	proxy := &core_xds.Proxy{
		Id:         core_xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		Zone:       p.zone,
		ZoneEgressProxy: &core_xds.ZoneEgressProxy{
			ZoneEgressResource: zoneEgress,
			ZoneIngresses:      zoneIngresses,
			MeshResourcesList:  meshResourcesList,
		},
	}
	for k, pl := range core_plugins.Plugins().ProxyPlugins() {
		err := pl.Apply(ctx, xds_context.MeshContext{}, proxy) // No mesh context for zone proxies
		if err != nil {
			return nil, errors.Wrapf(err, "Failed applying proxy plugin: %s", k)
		}
	}

	return proxy, nil
}

func matchEgressPolicies(tags map[string]string, resources xds_context.Resources) (core_xds.PluginOriginatedPolicies, error) {
	pluginPolicies := core_xds.PluginOriginatedPolicies{}
	for _, plugin := range core_plugins.Plugins().PolicyPlugins(ordered.Policies) {
		egressPlugin, ok := plugin.Plugin.(core_plugins.EgressPolicyPlugin)
		if !ok {
			continue
		}
		res, err := egressPlugin.EgressMatchedPolicies(tags, resources)
		if err != nil {
			return nil, errors.Wrapf(err, "could not apply policy plugin %s", plugin.Name)
		}
		if res.Type == "" {
			return nil, errors.Errorf("matched policy didn't set type for policy plugin %s", plugin.Name)
		}
		pluginPolicies[res.Type] = res
	}

	return pluginPolicies, nil
}
