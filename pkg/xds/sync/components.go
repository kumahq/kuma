package sync

import (
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var xdsServerLog = core.Log.WithName("xds").WithName("server")

func DefaultDataplaneProxyBuilder(
	zoneName string,
	apiVersion core_xds.APIVersion,
) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		Zone:       zoneName,
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

func DefaultDataplaneWatchdogFactory(
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	dataplaneReconciler SnapshotReconciler,
	ingressReconciler SnapshotReconciler,
	egressReconciler SnapshotReconciler,
	xdsMetrics *xds_metrics.Metrics,
	envoyCpCtx *xds_context.ControlPlaneContext,
	apiVersion core_xds.APIVersion,
) (DataplaneWatchdogFactory, error) {
	config := rt.Config()

	dataplaneProxyBuilder := DefaultDataplaneProxyBuilder(
		config.Multizone.Zone.Name,
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
		MetadataTracker:       metadataTracker,
		ResManager:            rt.ReadOnlyResourceManager(),
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		config.XdsServer.DataplaneConfigurationRefreshInterval.Duration,
		deps,
	)
}
