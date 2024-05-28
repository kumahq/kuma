package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	cache_kds_v2 "github.com/kumahq/kuma/pkg/kds/v2/cache"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	// Reconcile reconciles state of node given changed resource types.
	// Returns error and bool which is true if any resource was changed.
	Reconcile(context.Context, *envoy_core.Node, map[model.ResourceType]struct{}, logr.Logger) (error, bool)
	// ForceVersion marks that resource type for a node ID will obtain a new version even if nothing changes.
	// Note that it does not change snapshot, for this to actually apply on Envoy, we need to call Reconcile.
	// It's not called immediately to avoid parallel Reconcile calls for the same node.
	ForceVersion(node *envoy_core.Node, resourceType model.ResourceType)
	Clear(context.Context, *envoy_core.Node) error
}

// Generates a snapshot of xDS resources for a given node.
type SnapshotGenerator interface {
	GenerateSnapshot(context.Context, *envoy_core.Node, cache_kds_v2.SnapshotBuilder, map[model.ResourceType]struct{}) (envoy_cache.ResourceSnapshot, error)
}
