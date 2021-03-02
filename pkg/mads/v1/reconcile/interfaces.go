package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(context.Context, *envoy_core.Node) error
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_core.Node) (util_xds.Snapshot, error)
}
