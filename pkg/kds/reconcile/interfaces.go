package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(context.Context, *envoy_core.Node) error
	Clear(context.Context, *envoy_core.Node)
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_core.Node) (util_xds_v3.Snapshot, error)
}
