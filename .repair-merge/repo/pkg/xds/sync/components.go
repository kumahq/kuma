package sync

import (
	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/v3/pkg/xds/metrics"
	otelstatus "github.com/kumahq/kuma/v3/pkg/xds/otel/status"
)

var xdsServerLog = core.Log.WithName("xds").WithName("server")

func DefaultDataplaneProxyBuilder(
	config kuma_cp.Config,
	apiVersion core_xds.APIVersion,
) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		Zone:              config.Multizone.Zone.Name,
		APIVersion:        apiVersion,
		InternalAddresses: core_xds.InternalAddressesFromCIDRs(config.IPAM.KnownInternalCIDRs),
	}
}

// DataplaneWatchdogFactory returns a Watchdog that creates a new XdsContext and Proxy and executes SnapshotReconciler if there is any change
func DefaultDataplaneWatchdogFactory(
	rt core_runtime.Runtime,
	dataplaneReconciler SnapshotReconciler,
	xdsMetrics *xds_metrics.Metrics,
	envoyCpCtx *xds_context.ControlPlaneContext,
	otelStatusCache *otelstatus.Cache,
	apiVersion core_xds.APIVersion,
) (DataplaneWatchdogFactory, error) {
	config := rt.Config()

	dataplaneProxyBuilder := DefaultDataplaneProxyBuilder(
		config,
		apiVersion,
	)
	if config.XdsServer.PolicyMatchingCacheSize > 0 {
		dataplaneProxyBuilder = dataplaneProxyBuilder.WithPolicyMatchingCache(matchers.NewPolicyMatchingCache(
			xdsMetrics.PolicyMatchingCache,
			config.XdsServer.PolicyMatchingCacheSize,
		))
	}

	deps := DataplaneWatchdogDependencies{
		DataplaneProxyBuilder: dataplaneProxyBuilder,
		DataplaneReconciler:   dataplaneReconciler,
		EnvoyCpCtx:            envoyCpCtx,
		MeshCache:             rt.MeshCache(),
		ResManager:            rt.ReadOnlyResourceManager(),
		OtelStatusCache:       otelStatusCache,
		XdsMetrics:            xdsMetrics,
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		config.XdsServer.DataplaneConfigurationRefreshInterval.Duration,
		deps,
	)
}
