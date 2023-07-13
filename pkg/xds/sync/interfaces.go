package sync

import (
	"context"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type DataplaneMetadataTracker interface {
	Metadata(dpKey core_model.ResourceKey) *core_xds.DataplaneMetadata
}

type ConnectionInfoTracker interface {
	ConnectionInfo(dpKey core_model.ResourceKey) *xds_context.ConnectionInfo
}

// SnapshotReconciler reconciles Envoy XDS configuration (Snapshot) by executing all generators (pkg/xds/generator)
type SnapshotReconciler interface {
<<<<<<< HEAD
	Reconcile(ctx xds_context.Context, proxy *core_xds.Proxy) error
=======
	Reconcile(context.Context, xds_context.Context, *core_xds.Proxy) (bool, error)
>>>>>>> df9c5f925 (fix(kuma-cp): pass context via snapshot reconciler to generateCerts (#7231))
	Clear(proxyId *core_xds.ProxyId) error
}

// DataplaneWatchdogFactory returns a Watchdog that creates a new XdsContext and Proxy and executes SnapshotReconciler if there is any change
type DataplaneWatchdogFactory interface {
	New(dpKey core_model.ResourceKey) util_watchdog.Watchdog
}
