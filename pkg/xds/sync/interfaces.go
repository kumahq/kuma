package sync

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type DataplaneMetadataTracker interface {
	Metadata(streamID core_xds.StreamID) *core_xds.DataplaneMetadata
}

type ConnectionInfoTracker interface {
	ConnectionInfo(streamID core_xds.StreamID) xds_context.ConnectionInfo
}

// SnapshotReconciler reconciles Envoy XDS configuration (Snapshot) by executing all generators (pkg/xds/generator)
type SnapshotReconciler interface {
	Reconcile(ctx xds_context.Context, proxy *core_xds.Proxy) error
	Clear(proxyId *core_xds.ProxyId) error
}

// DataplaneWatchdogFactory returns a Watchdog that creates a new XdsContext and Proxy and executes SnapshotReconciler if there is any change
type DataplaneWatchdogFactory interface {
	New(proxyId *core_xds.ProxyId, streamId int64) util_watchdog.Watchdog
}
