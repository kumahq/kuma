package sync

import (
	"context"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
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
	Reconcile(context.Context, xds_context.Context, *core_xds.Proxy) (bool, error)
	Clear(proxyId *core_xds.ProxyId) error
}

// DataplaneWatchdogFactory returns a Watchdog that creates a new XdsContext and Proxy and executes SnapshotReconciler if there is any change
type DataplaneWatchdogFactory interface {
<<<<<<< HEAD
	New(dpKey core_model.ResourceKey) util_xds_v3.Watchdog
=======
	New(dpKey core_model.ResourceKey, meta *core_xds.DataplaneMetadata) util_xds_v3.Watchdog
}

// DataplaneWatchdogFactoryWithStreamCtx extends DataplaneWatchdogFactory with stream context support.
// When the stream context is closed, the watchdog will skip ticks to prevent
// race conditions between gRPC stream closure and xDS snapshot updates.
type DataplaneWatchdogFactoryWithStreamCtx interface {
	DataplaneWatchdogFactory
	NewWithStreamCtx(dpKey core_model.ResourceKey, meta *core_xds.DataplaneMetadata, streamCtx context.Context) util_xds_v3.Watchdog
}

type DataplaneWatchdogFactoryFunc func(dpKey core_model.ResourceKey, meta *core_xds.DataplaneMetadata) util_xds_v3.Watchdog

func (f DataplaneWatchdogFactoryFunc) New(dpKey core_model.ResourceKey, meta *core_xds.DataplaneMetadata) util_xds_v3.Watchdog {
	return f(dpKey, meta)
>>>>>>> 42c3b352ba (fix(xds): prevent panic on send to closed channel during stream closure (#15511))
}
