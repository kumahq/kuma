package cache

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
)

type ResourceBuilder interface{}

type SnapshotBuilder interface {
	With(typ string, resources []envoy_types.Resource) SnapshotBuilder
	Build(version string) envoy_cache.ResourceSnapshot
}

type builder struct {
	resources map[string][]envoy_types.ResourceWithTTL
}

func (b *builder) With(typ string, resources []envoy_types.Resource) SnapshotBuilder {
	ttlResources := make([]envoy_types.ResourceWithTTL, len(resources))
	for i, res := range resources {
		ttlResources[i] = envoy_types.ResourceWithTTL{
			Resource: res,
			TTL:      nil,
		}
	}
	b.resources[typ] = ttlResources
	return b
}

func (b *builder) Build(version string) envoy_cache.ResourceSnapshot {
	snapshot := &Snapshot{Resources: map[string]envoy_cache.Resources{}}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		snapshot.Resources[typ] = envoy_cache.NewResources(version, nil)
	}
	for typ, items := range b.resources {
		snapshot.Resources[typ] = envoy_cache.Resources{Version: version, Items: IndexResourcesByName(items)}
	}
	return snapshot
}

func NewSnapshotBuilder() SnapshotBuilder {
	return &builder{resources: map[string][]envoy_types.ResourceWithTTL{}}
}