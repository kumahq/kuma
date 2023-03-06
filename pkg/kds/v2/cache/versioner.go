package cache

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/proto"

	util_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/util"
)

// SnapshotVersioner assigns versions to xDS resources in a new Snapshot.
type SnapshotVersioner interface {
	Version(new, old envoy_cache.ResourceSnapshot) envoy_cache.ResourceSnapshot
}

// SnapshotAutoVersioner assigns versions to xDS resources in a new Snapshot
// by reusing if possible a version from the old snapshot and
// generating a new version (UUID) otherwise.
type SnapshotAutoVersioner struct {
	UUID func() string
}

func (v SnapshotAutoVersioner) Version(new, old envoy_cache.ResourceSnapshot) envoy_cache.ResourceSnapshot {
	if new == nil {
		return nil
	}
	newResources := map[string]envoy_cache.Resources{}
	for _, typ := range util_kds_v2.GetSupportedTypes() {
		version := new.GetVersion(typ)
		if version != "" {
			// favor a version assigned by resource generator
			continue
		}

		if old != nil && v.equal(new.GetResources(typ), old.GetResources(typ)) {
			version = old.GetVersion(typ)
		}
		if version == "" {
			version = v.UUID()
		}
		if new == nil {
			continue
		}
		if new.GetVersion(typ) == version {
			continue
		}

		n := map[string]envoy_types.ResourceWithTTL{}
		for k, v := range new.GetResourcesAndTTL(typ) {
			n[k] = v
		}
		newResources[typ] = envoy_cache.Resources{Version: version, Items: n}
	}
	return &Snapshot{
		Resources: newResources,
	}
}

func (_ SnapshotAutoVersioner) equal(new, old map[string]envoy_types.Resource) bool {
	if len(new) != len(old) {
		return false
	}
	for key, newValue := range new {
		if oldValue, hasOldValue := old[key]; !hasOldValue || !proto.Equal(newValue, oldValue) {
			return false
		}
	}
	return true
}
