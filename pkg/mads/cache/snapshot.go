package cache

import (
	"github.com/pkg/errors"

	envoy_types "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"

	"github.com/Kong/kuma/pkg/mads"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

// NewSnapshot creates a snapshot from response types and a version.
func NewSnapshot(version string, assignments map[string]envoy_types.Resource) *Snapshot {
	return &Snapshot{
		MonitoringAssignments: envoy_cache.Resources{Version: version, Items: assignments},
	}
}

// Snapshot is an internally consistent snapshot of xDS resources.
type Snapshot struct {
	MonitoringAssignments envoy_cache.Resources
}

var _ util_xds.Snapshot = &Snapshot{}

// GetSupportedTypes returns a list of xDS types supported by this snapshot.
func (s *Snapshot) GetSupportedTypes() []string {
	return []string{mads.MonitoringAssignmentType}
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
	switch typ {
	case mads.MonitoringAssignmentType:
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
	case mads.MonitoringAssignmentType:
		return s.MonitoringAssignments.Version
	}
	return ""
}

// WithVersion creates a new snapshot with a different version for a given resource type.
func (s *Snapshot) WithVersion(typ string, version string) util_xds.Snapshot {
	if s == nil {
		return nil
	}
	if s.GetVersion(typ) == version {
		return s
	}
	switch typ {
	case mads.MonitoringAssignmentType:
		return &Snapshot{
			MonitoringAssignments: envoy_cache.Resources{Version: version, Items: s.MonitoringAssignments.Items},
		}
	}
	return s
}
