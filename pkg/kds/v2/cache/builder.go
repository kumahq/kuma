package cache

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceBuilder interface{}

type SnapshotBuilder interface {
	With(typ core_model.ResourceType, resources []envoy_types.Resource) SnapshotBuilder
	Build(version string) envoy_cache.ResourceSnapshot
}

type builder struct {
	resources      map[core_model.ResourceType][]envoy_types.ResourceWithTTL
	supportedTypes []core_model.ResourceType
}

func (b *builder) With(typ core_model.ResourceType, resources []envoy_types.Resource) SnapshotBuilder {
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
	snapshot := &Snapshot{Resources: map[core_model.ResourceType]envoy_cache.Resources{}}
	for _, typ := range b.supportedTypes {
		items, exists := b.resources[typ]
		if exists {
			snapshot.Resources[typ] = envoy_cache.Resources{Version: version, Items: IndexResourcesByName(items)}
		} else {
			snapshot.Resources[typ] = envoy_cache.NewResources(version, nil)
		}
	}
	return snapshot
}

func NewSnapshotBuilder(supportedTypes []core_model.ResourceType) SnapshotBuilder {
	return &builder{
		resources:      map[core_model.ResourceType][]envoy_types.ResourceWithTTL{},
		supportedTypes: supportedTypes,
	}
}
