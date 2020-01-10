package reconcile

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(*envoy_core.Node) error
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(*envoy_core.Node) (util_xds.Snapshot, error)
}
