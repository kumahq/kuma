package sync

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/permissions"
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
	DataSourceLoader   datasource.Loader

	zone       string
	apiVersion envoy_common.APIVersion
}

func (p *EgressProxyBuilder) Build(key core_model.ResourceKey) (*xds.Proxy, error) {
	ze := core_mesh.NewZoneEgressResource()
	if err := p.ReadOnlyResManager.Get(p.ctx, ze, core_store.GetBy(key)); err != nil {
		return nil, err
	}

	var meshList core_mesh.MeshResourceList
	if err := p.ReadOnlyResManager.List(
		p.ctx,
		&meshList,
	); err != nil {
		return nil, err
	}

	var meshes []*core_mesh.MeshResource
	for _, mesh := range meshList.Items {
		if mesh.MTLSEnabled() {
			meshes = append(meshes, mesh)
		}
	}

	var zoneIngressesList core_mesh.ZoneIngressResourceList
	if err := p.ReadOnlyResManager.List(
		p.ctx,
		&zoneIngressesList,
	); err != nil {
		return nil, err
	}

	var zoneIngresses []*core_mesh.ZoneIngressResource
	for _, zoneIngress := range zoneIngressesList.Items {
		if zoneIngress.IsRemoteIngress(p.zone) {
			zoneIngresses = append(zoneIngresses, zoneIngress)
		}
	}

	externalServiceMap := map[string][]*core_mesh.ExternalServiceResource{}
	meshEndpointMap := map[string]xds.EndpointMap{}
	trafficRouteMap := map[string][]*core_mesh.TrafficRouteResource{}

	for _, mesh := range meshes {
		meshName := mesh.GetMeta().GetName()

		meshExternalServiceMap := map[string]*core_mesh.ExternalServiceResource{}

		meshCtx, err := p.meshCache.GetMeshContext(p.ctx, syncLog, meshName)
		if err != nil {
			return nil, err
		}

		meshResources := meshCtx.Resources
		trafficPermissions := meshResources.TrafficPermissions()
		trafficRoutes := meshResources.TrafficRoutes().Items
		externalServices := meshResources.ExternalServices()

		trafficRouteMap[meshName] = trafficRoutes

		for _, dp := range meshCtx.DataplanesByName {
			if !dp.Spec.Matches(map[string]string{
				mesh_proto.ZoneTag: p.zone,
			}) {
				continue
			}

			allowedExternalServices, err := permissions.MatchExternalServices(
				dp,
				externalServices,
				trafficPermissions,
			)
			if err != nil {
				return nil, err
			}

			for _, es := range allowedExternalServices {
				esName := es.GetMeta().GetName()

				if _, ok := meshExternalServiceMap[esName]; !ok {
					meshExternalServiceMap[esName] = es
				}
			}
		}

		var meshExternalServices []*core_mesh.ExternalServiceResource
		for _, es := range meshExternalServiceMap {
			meshExternalServices = append(meshExternalServices, es)
		}

		externalServiceMap[meshName] = meshExternalServices

		meshEndpointMap[meshName] = xds_topology.BuildRemoteEndpointMap(
			mesh,
			p.zone,
			zoneIngresses,
			externalServices.Items,
			p.DataSourceLoader,
		)
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		ZoneEgressProxy: &xds.ZoneEgressProxy{
			ZoneEgressResource: ze,
			DataSourceLoader:   p.DataSourceLoader,
			ExternalServiceMap: externalServiceMap,
			Zone:               p.zone,
			Meshes:             meshes,
			MeshEndpointMap:    meshEndpointMap,
			TrafficRouteMap:    trafficRouteMap,
			ZoneIngresses:      zoneIngresses,
		},
		Metadata: p.MetadataTracker.Metadata(key),
	}

	return proxy, nil
}
