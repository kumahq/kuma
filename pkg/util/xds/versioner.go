package xds

import (
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/proto"
)

// SnapshotVersioner assigns versions to xDS resources in a new Snapshot.
type SnapshotVersioner interface {
	Version(new, old Snapshot)
}

// SnapshotAutoVersioner assigns versions to xDS resources in a new Snapshot
// by reusing if possible a version from the old snapshot and
// generating a new version (UUID) otherwise.
type SnapshotAutoVersioner struct {
	UUID func() string
}

func (v SnapshotAutoVersioner) Version(new, old Snapshot) {
	for _, typ := range new.GetSupportedTypes() {
		version := new.GetVersion(typ)
		if version != "" {
			// favour a version assigned by resource generator
			continue
		}
		if old != nil && v.equal(new.GetResources(typ), old.GetResources(typ)) {
			version = old.GetVersion(typ)
		}
		if version == "" {
			version = v.UUID()
		}
		new.SetVersion(typ, version)
	}
}

func (_ SnapshotAutoVersioner) equal(new, old map[string]envoy_cache.Resource) bool {
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
