package reconcile

import (
	"context"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/go-logr/logr"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(context.Context, logr.Logger) error
	// NeedsReconciliation checks if there is a valid configuration snapshot already present
	// for a given node
	NeedsReconciliation(node *envoy_core.Node) bool
	KnownClientIds() map[string]bool
}
