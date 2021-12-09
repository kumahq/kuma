package cache

import (
	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/pkg/errors"

	v1 "github.com/kumahq/kuma/pkg/mads/v1"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

// NewSnapshot creates a snapshot from response types and a version.
func NewSnapshot(version string, assignments map[string]envoy_types.Resource) *Snapshot {
	withTtl := make(map[string]envoy_types.ResourceWithTTL, len(assignments))
	for name, res := range assignments {
		withTtl[name] = envoy_types.ResourceWithTTL{
			Resource: res,
		}
	}
	return &Snapshot{
		MonitoringAssignments: envoy_cache.Resources{Version: version, Items: withTtl},
	}
}

// Snapshot is an internally consistent snapshot of xDS resources.
type Snapshot struct {
	MonitoringAssignments envoy_cache.Resources
}

var _ util_xds_v3.Snapshot = &Snapshot{}

// GetSupportedTypes returns a list of xDS types supported by this snapshot.
func (s *Snapshot) GetSupportedTypes() []string {
	return []string{v1.MonitoringAssignmentType}
}

// Consistent check verifies that the dependent resources are exactly listed in the
// snapshot.
func (s *Snapshot) Consistent() error {
	if s == nil {
		return errors.New("nil snapshot")
	}
	return nil
}

// GetResources selects snapshot resources by type.
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

func (s *Snapshot) GetResourcesAndTtl(typ string) map[string]envoy_types.ResourceWithTTL {
	if s == nil {
		return nil
	}
	switch typ {
	case v1.MonitoringAssignmentType:
		return s.MonitoringAssignments.Items
	}
	return nil
}

// GetVersion returns the version for a resource type.
func (s *Snapshot) GetVersion(typ string) string {
	if s == nil {
		return ""
	}
	switch typ {
	case v1.MonitoringAssignmentType:
		return s.MonitoringAssignments.Version
	}
	return ""
}

// WithVersion creates a new snapshot with a different version for a given resource type.
func (s *Snapshot) WithVersion(typ string, version string) util_xds_v3.Snapshot {
	if s == nil {
		return nil
	}
	if s.GetVersion(typ) == version {
		return s
	}
	switch typ {
	case v1.MonitoringAssignmentType:
		return &Snapshot{
			MonitoringAssignments: envoy_cache.Resources{Version: version, Items: s.MonitoringAssignments.Items},
		}
	}
	return s
}
