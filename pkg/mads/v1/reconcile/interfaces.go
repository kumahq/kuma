package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(context.Context) error
	// NeedsReconciliation checks if there is a valid configuration snapshot already present
	// for a given node
	NeedsReconciliation(node *envoy_core.Node) bool
	KnownClientIds() map[string]bool
	ReconcileIfNeeded(ctx context.Context, node *envoy_core.Node) error
}
