package v3

import (
	"context"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

type Snapshot interface {
	envoy_cache.ResourceSnapshot

	Consistent() error
	// WithVersion creates a new snapshot with a different version for a given resource type.
	WithVersion(typ string, version string) Snapshot
}

type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_config_core_v3.Node) (Snapshot, error)
}