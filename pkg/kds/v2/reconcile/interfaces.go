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
	Clear(*envoy_core.Node)
=======
	// Reconcile reconciles state of node given changed resource types.
	// Returns error and bool which is true if any resource was changed.
	Reconcile(context.Context, *envoy_core.Node, map[model.ResourceType]struct{}, logr.Logger) (error, bool)
	// ForceVersion marks that resource type for a node ID will obtain a new version even if nothing changes.
	// Note that it does not change snapshot, for this to actually apply on Envoy, we need to call Reconcile.
	// It's not called immediately to avoid parallel Reconcile calls for the same node.
	ForceVersion(node *envoy_core.Node, resourceType model.ResourceType)
	Clear(context.Context, *envoy_core.Node) error
>>>>>>> 4752f7b82 (fix(kds): fix retry on NACK and add backoff (#9736))
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_core.Node) (envoy_cache.ResourceSnapshot, error)
}
