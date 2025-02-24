package sync

import (
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var xdsServerLog = core.Log.WithName("xds").WithName("server")

func DefaultDataplaneProxyBuilder(
	config kuma_cp.Config,
	apiVersion core_xds.APIVersion,
) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		Zone:       config.Multizone.Zone.Name,
		APIVersion: apiVersion,
	}
}

func DefaultIngressProxyBuilder(
	rt core_runtime.Runtime,
	apiVersion core_xds.APIVersion,
) *IngressProxyBuilder {
	return &IngressProxyBuilder{
		ResManager:        rt.ResourceManager(),
		LookupIP:          rt.LookupIP(),
		apiVersion:        apiVersion,
		zone:              rt.Config().Multizone.Zone.Name,
		ingressTagFilters: rt.Config().Experimental.IngressTagFilters,
	}
}

func DefaultEgressProxyBuilder(rt core_runtime.Runtime, apiVersion core_xds.APIVersion) *EgressProxyBuilder {
	return &EgressProxyBuilder{
		apiVersion: apiVersion,
		zone:       rt.Config().Multizone.Zone.Name,
	}
}

// DataplaneWatchdogFactory returns a Watchdog that creates a new XdsContext and Proxy and executes SnapshotReconciler if there is any change
func DefaultDataplaneWatchdogFactory(
	rt core_runtime.Runtime,
	dataplaneReconciler SnapshotReconciler,
	ingressReconciler SnapshotReconciler,
	egressReconciler SnapshotReconciler,
	xdsMetrics *xds_metrics.Metrics,
	envoyCpCtx *xds_context.ControlPlaneContext,
	apiVersion core_xds.APIVersion,
) DataplaneWatchdogFactory {
	config := rt.Config()

	dataplaneProxyBuilder := DefaultDataplaneProxyBuilder(
		config,
		apiVersion,
	)

	ingressProxyBuilder := DefaultIngressProxyBuilder(
		rt,
		apiVersion,
	)

	egressProxyBuilder := DefaultEgressProxyBuilder(rt, apiVersion)

	deps := DataplaneWatchdogDependencies{
		DataplaneProxyBuilder: dataplaneProxyBuilder,
		DataplaneReconciler:   dataplaneReconciler,
		IngressProxyBuilder:   ingressProxyBuilder,
		IngressReconciler:     ingressReconciler,
		EgressProxyBuilder:    egressProxyBuilder,
		EgressReconciler:      egressReconciler,
		EnvoyCpCtx:            envoyCpCtx,
		MeshCache:             rt.MeshCache(),
		ResManager:            rt.ReadOnlyResourceManager(),
	}
	return NewDataplaneWatchdogFactory(
		deps,
		config.XdsServer.DataplaneConfigurationRefreshInterval.Duration,
		xdsMetrics,
	)
}
