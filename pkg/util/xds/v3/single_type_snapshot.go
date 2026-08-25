package v3

import (
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/proto"
)

// Snapshot is an internally consistent snapshot of xDS resources.
// Consistency is important for the convergence as different resource types
// from the snapshot may be delivered to the proxy in arbitrary order.
type SingleTypeSnapshot struct {
	Resources envoy_cache.Resources
	TypeUrl   string

	// VersionMap holds the current hash map of all resources in the snapshot.
	// This field should remain nil until it is used, at which point should be
	// instantiated by calling ConstructVersionMap().
	// VersionMap is only to be used with delta xDS.
	VersionMap map[string]map[string]string
}

var _ envoy_cache.ResourceSnapshot = &SingleTypeSnapshot{}

// NewSingleTypeSnapshot creates a snapshot from response types and a version.
// The resources map is keyed off the type URL of a resource, followed by the slice of resource objects.
func NewSingleTypeSnapshot(version string, typeURL string, resources []types.Resource) *SingleTypeSnapshot {
	return &SingleTypeSnapshot{
		TypeUrl:   typeURL,
		Resources: envoy_cache.NewResources(version, resources),
	}
}

// GetResources selects snapshot resources by type, returning the map of resources.
func (s *SingleTypeSnapshot) GetResources(typeURL resource.Type) map[string]types.Resource {
	resources := s.GetResourcesAndTTL(typeURL)
	if resources == nil {
		return nil
	}

	withoutTTL := make(map[string]types.Resource, len(resources))

	for k, v := range resources {
		withoutTTL[k] = v.Resource
	}

	return withoutTTL
}

// GetResourcesAndTTL selects snapshot resources by type, returning the map of resources and the associated TTL.
func (s *SingleTypeSnapshot) GetResourcesAndTTL(typeURL resource.Type) map[string]types.ResourceWithTTL {
	if s == nil {
		return nil
	}
	if typeURL != s.TypeUrl {
		return nil
	}
	return s.Resources.Items
}

// GetVersion returns the version for a resource type.
func (s *SingleTypeSnapshot) GetVersion(typeURL resource.Type) string {
	if s == nil {
		return ""
	}
	if typeURL != s.TypeUrl {
		return ""
	}
	return s.Resources.Version
}

// GetVersionMap will return the internal version map of the currently applied snapshot.
func (s *SingleTypeSnapshot) GetVersionMap(typeURL string) map[string]string {
	return s.VersionMap[typeURL]
}

// ConstructVersionMap will construct a version map based on the current state of a snapshot
func (s *SingleTypeSnapshot) ConstructVersionMap() error {
	if s == nil {
		return fmt.Errorf("missing snapshot")
	}

	// The snapshot resources never change, so no need to ever rebuild.
	if s.VersionMap != nil {
		return nil
	}

	s.VersionMap = make(map[string]map[string]string)

	s.VersionMap[s.TypeUrl] = make(map[string]string, len(s.Resources.Items))
	for _, r := range s.Resources.Items {
		// Hash our version in here and build the version map.
		marshaledResource, err := envoy_cache.MarshalResource(r.Resource)
		if err != nil {
			return err
		}
		v := envoy_cache.HashResource(marshaledResource)
		if v == "" {
			return fmt.Errorf("failed to build resource version: %w", err)
		}

		s.VersionMap[s.TypeUrl][envoy_cache.GetResourceName(r.Resource)] = v
	}

	return nil
}

// SingleTypeSnapshotEqual checks value equality of 2 snapshots that contain a single type.
// Snapshots that aren't a *SingleTypeSnapshot are always reported as different, so the
// caller replaces them with the new one.
func SingleTypeSnapshotEqual(newSnap, oldSnap envoy_cache.ResourceSnapshot) bool {
	snap, ok := newSnap.(*SingleTypeSnapshot)
	if !ok {
		return false
	}
	return snap.Equal(oldSnap)
}

// Equal checks value equality against another snapshot, comparing the resources
// of this snapshot's single type.
func (s *SingleTypeSnapshot) Equal(other envoy_cache.ResourceSnapshot) bool {
	if other == nil {
		return false
	}
	// For now there's a single resourceType so the diff is easy
	newResources := s.GetResources(s.TypeUrl)
	oldResources := other.GetResources(s.TypeUrl)
	if len(newResources) != len(oldResources) {
		return false
	}
	for key, newValue := range newResources {
		if oldValue, hasOldValue := oldResources[key]; !hasOldValue || !proto.Equal(newValue, oldValue) {
			return false
		}
	}
	return true
}
