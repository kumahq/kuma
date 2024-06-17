package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
<<<<<<< HEAD
	Reconcile(context.Context, *envoy_core.Node) error
=======
	// Reconcile reconciles state of node given changed resource types.
	// Returns error and bool which is true if any resource was changed.
	Reconcile(context.Context, *envoy_core.Node, map[model.ResourceType]struct{}, logr.Logger) (error, bool)
>>>>>>> bc8adb233 (fix(kds): send NACK only when resource is invalid and do not retry (#10480))
	Clear(context.Context, *envoy_core.Node) error
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_core.Node) (envoy_cache.ResourceSnapshot, error)
}
