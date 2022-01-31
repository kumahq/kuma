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
)

type EgressProxyBuilder struct {
	ResManager         manager.ResourceManager
	ReadOnlyResManager manager.ReadOnlyResourceManager
	LookupIP           lookup.LookupIPFunc
	MetadataTracker    DataplaneMetadataTracker

	apiVersion envoy.APIVersion
}

func filterMeshesByMTLS(rs core_model.Resource) bool {
	return rs.(*core_mesh.MeshResource).Spec.GetMtls().GetEnabledBackend() != ""
}

func (p *EgressProxyBuilder) build(key core_model.ResourceKey) (*xds.Proxy, error) {
	ctx := context.Background()

	zoneEgress := core_mesh.NewZoneEgressResource()
	if err := p.ReadOnlyResManager.Get(ctx, zoneEgress, core_store.GetBy(key)); err != nil {
		return nil, err
	}

	var meshes core_mesh.MeshResourceList
	if err := p.ReadOnlyResManager.List(ctx, &meshes, core_store.ListByFilterFunc(filterMeshesByMTLS)); err != nil {
		return nil, err
	}

	var externalServices []*core_mesh.ExternalServiceResource
	for _, mesh := range meshes.Items {
		meshName := mesh.GetMeta().GetName()

		var es core_mesh.ExternalServiceResourceList
		if err := p.ReadOnlyResManager.List(ctx, &es, core_store.ListByMesh(meshName)); err != nil {
			return nil, err
		}

		externalServices = append(externalServices, es.Items...)
	}

	proxy := &xds.Proxy{
		Id:         xds.FromResourceKey(key),
		APIVersion: p.apiVersion,
		ZoneEgressProxy: &xds.ZoneEgressProxy{
			ZoneEgressResource: zoneEgress,
			Meshes:             meshes.Items,
			ExternalServices:   externalServices,
		},
		Metadata: p.MetadataTracker.Metadata(key),
	}

	return proxy, nil
}
