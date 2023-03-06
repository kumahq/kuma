package cache

import (
	"fmt"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// Snapshot is an internally consistent snapshot of xDS resources.
type Snapshot struct {
	Resources map[string]envoy_cache.Resources

	// VersionMap holds the current hash map of all resources in the snapshot.
	// This field should remain nil until it is used, at which point should be
	// instantiated by calling ConstructVersionMap().
	// VersionMap is only to be used with delta xDS.
	VersionMap map[string]map[string]string
}

var _ envoy_cache.ResourceSnapshot = &Snapshot{}

func (s *Snapshot) GetResources(typ string) map[string]envoy_types.Resource {
	if s == nil {
		return nil
	}

	resources := s.GetResourcesAndTTL(typ)
	if resources == nil {
		return nil
	}

	withoutTtl := make(map[string]envoy_types.Resource, len(resources))
	for name, res := range resources {
		withoutTtl[name] = res.Resource
	}
	return withoutTtl
}

func (s *Snapshot) GetResourcesAndTTL(typ string) map[string]envoy_types.ResourceWithTTL {
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

func (s *Snapshot) GetVersionMap(typeURL string) map[string]string {
	return s.VersionMap[typeURL]
}

// ConstructVersionMap will construct a version map based on the current state of a snapshot
func (s *Snapshot) ConstructVersionMap() error {
	if s == nil {
		return fmt.Errorf("missing snapshot")
	}

	// The snapshot resources never change, so no need to ever rebuild.
	if s.VersionMap != nil {
		return nil
	}

	s.VersionMap = make(map[string]map[string]string)

	for typeURL, resources := range s.Resources {
		if _, ok := s.VersionMap[typeURL]; !ok {
			s.VersionMap[typeURL] = make(map[string]string)
		}

		for _, r := range resources.Items {
			// Hash our version in here and build the version map.
			marshaledResource, err := envoy_cache.MarshalResource(r.Resource)
			if err != nil {
				return err
			}
			v := envoy_cache.HashResource(marshaledResource)
			if v == "" {
				return fmt.Errorf("failed to build resource version: %w", err)
			}
			s.VersionMap[typeURL][GetResourceName(r.Resource)] = v
		}
	}

	return nil
}

func GetResourceName(res envoy_types.Resource) string {
	switch v := res.(type) {
	case *mesh_proto.KumaResource:
		return fmt.Sprintf("%s.%s", v.GetMeta().GetName(), v.GetMeta().GetMesh())
	default:
		return ""
	}
}

// IndexResourcesByName creates a map from the resource name to the resource. Name should be unique
// across meshes that's why Name is <name>.<mesh>
func IndexResourcesByName(items []envoy_types.ResourceWithTTL) map[string]envoy_types.ResourceWithTTL {
	indexed := make(map[string]envoy_types.ResourceWithTTL, len(items))
	for _, item := range items {
		key := GetResourceName(item.Resource)
		indexed[key] = item
	}
	return indexed
}
