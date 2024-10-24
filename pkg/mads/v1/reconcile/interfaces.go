package reconcile

import (
	"context"
)

// Reconciler re-computes configuration for a given node.
type Reconciler interface {
	Reconcile(context.Context) error
	KnownClientIds() []string
}
