package runtime

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Component of the Control Plane, i.e. gRPC Server, HTTP server, reconciliation loop.
type Component = manager.Runnable
