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
	New(dpKey core_model.ResourceKey) util_xds_v3.Watchdog
}
