package cache

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

type ResourceBuilder any

type SnapshotBuilder interface {
	With(typ core_model.ResourceType, resources []envoy_types.Resource) SnapshotBuilder
	WithPrecomputedVersions(typ core_model.ResourceType, versions NameToVersion) SnapshotBuilder
	WithIndexedResources(typ core_model.ResourceType, items map[string]envoy_types.ResourceWithTTL) SnapshotBuilder
	Build(version string) envoy_cache.ResourceSnapshot
}

type builder struct {
	resources      map[core_model.ResourceType][]envoy_types.ResourceWithTTL
	indexed        map[core_model.ResourceType]map[string]envoy_types.ResourceWithTTL
	versions       ResourceVersionMap
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

func (b *builder) WithIndexedResources(typ core_model.ResourceType, items map[string]envoy_types.ResourceWithTTL) SnapshotBuilder {
	if len(items) > 0 {
		b.indexed[typ] = items
	}
	return b
}

func (b *builder) WithPrecomputedVersions(typ core_model.ResourceType, versions NameToVersion) SnapshotBuilder {
	if len(versions) > 0 {
		b.versions[typ] = versions
	}
	return b
}

func (b *builder) Build(version string) envoy_cache.ResourceSnapshot {
	snapshot := &Snapshot{Resources: map[core_model.ResourceType]envoy_cache.Resources{}}
	if len(b.versions) > 0 {
		snapshot.VersionMap = b.versions
	}
	for _, typ := range b.supportedTypes {
		if carried, ok := b.indexed[typ]; ok {
			snapshot.Resources[typ] = envoy_cache.Resources{Version: version, Items: carried}
			continue
		}
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
		indexed:        map[core_model.ResourceType]map[string]envoy_types.ResourceWithTTL{},
		versions:       ResourceVersionMap{},
		supportedTypes: supportedTypes,
	}
}
