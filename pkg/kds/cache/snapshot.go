package cache

import (
	"fmt"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/kds"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type ResourceBuilder interface {
}

type SnapshotBuilder interface {
	With(typ string, resources []envoy_types.Resource) SnapshotBuilder
	Build(version string) util_xds.Snapshot
}

type builder struct {
	resources map[string][]envoy_types.ResourceWithTtl
}

func (b *builder) With(typ string, resources []envoy_types.Resource) SnapshotBuilder {
	ttlResources := make([]envoy_types.ResourceWithTtl, len(resources))
	for i, res := range resources {
		ttlResources[i] = envoy_types.ResourceWithTtl{
			Resource: res,
			Ttl:      nil,
		}
	}
	b.resources[typ] = ttlResources
	return b
}

func (b *builder) Build(version string) util_xds.Snapshot {
	snapshot := &Snapshot{Resources: map[string]envoy_cache.Resources{}}
	for _, typ := range snapshot.GetSupportedTypes() {
		snapshot.Resources[typ] = envoy_cache.NewResources(version, nil)
	}
	for typ, items := range b.resources {
		snapshot.Resources[typ] = envoy_cache.Resources{Version: version, Items: IndexResourcesByName(items)}
	}
	return snapshot
}

func NewSnapshotBuilder() SnapshotBuilder {
	return &builder{resources: map[string][]envoy_types.ResourceWithTtl{}}
}

// Snapshot is an internally consistent snapshot of xDS resources.
type Snapshot struct {
	Resources map[string]envoy_cache.Resources
}

var _ util_xds.Snapshot = &Snapshot{}

func (s *Snapshot) GetSupportedTypes() (types []string) {
	for _, typ := range kds.SupportedTypes {
		types = append(types, string(typ))
	}
	return
}

func (s *Snapshot) Consistent() error {
	if s == nil {
		return errors.New("nil snapshot")
	}
	return nil
}

func (s *Snapshot) GetResources(typ string) map[string]envoy_types.Resource {
	if s == nil {
		return nil
	}

	resources := s.GetResourcesAndTtl(typ)
	if resources == nil {
		return nil
	}

	withoutTtl := make(map[string]envoy_types.Resource, len(resources))
	for name, res := range resources {
		withoutTtl[name] = res.Resource
	}
	return withoutTtl
}

func (s *Snapshot) GetResourcesAndTtl(typ string) map[string]envoy_types.ResourceWithTtl {
	if s == nil {
		return nil
	}
	if r, ok := s.Resources[typ]; ok {
		return r.Items
	}
	return nil
}

func (s *Snapshot) GetVersion(typ string) string {
	if s == nil {
		return ""
	}
	if r, ok := s.Resources[typ]; ok {
		return r.Version
	}
	return ""
}

func (s *Snapshot) WithVersion(typ string, version string) util_xds.Snapshot {
	if s == nil {
		return nil
	}
	if s.GetVersion(typ) == version {
		return s
	}
	n := map[string]envoy_cache.Resources{}
	for k, v := range s.Resources {
		n[k] = v
	}
	if r, ok := n[typ]; ok {
		n[typ] = envoy_cache.Resources{Version: version, Items: r.Items}
		return &Snapshot{
			Resources: n,
		}
	}
	return s
}

// IndexResourcesByName creates a map from the resource name to the resource. Name should be unique
// across meshes that's why Name is <name>.<mesh>
func IndexResourcesByName(items []envoy_types.ResourceWithTtl) map[string]envoy_types.ResourceWithTtl {
	indexed := make(map[string]envoy_types.ResourceWithTtl, len(items))
	for _, item := range items {
		key := fmt.Sprintf("%s.%s", item.Resource.(*mesh_proto.KumaResource).GetMeta().GetName(), item.Resource.(*mesh_proto.KumaResource).GetMeta().GetMesh())
		indexed[key] = item
	}
	return indexed
}
