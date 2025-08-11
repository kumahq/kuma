package sync

import (
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var xdsServerLog = core.Log.WithName("xds-server")

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
		InternalAddresses: core_xds.InternalAddressesFromCIDRs(rt.Config().IPAM.KnownInternalCIDRs),
	}
}

func DefaultEgressProxyBuilder(rt core_runtime.Runtime, apiVersion core_xds.APIVersion) *EgressProxyBuilder {
	return &EgressProxyBuilder{
		apiVersion:        apiVersion,
		zone:              rt.Config().Multizone.Zone.Name,
		InternalAddresses: core_xds.InternalAddressesFromCIDRs(rt.Config().IPAM.KnownInternalCIDRs),
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
		MetadataTracker:       metadataTracker,
		ResManager:            rt.ReadOnlyResourceManager(),
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		config.XdsServer.DataplaneConfigurationRefreshInterval.Duration,
		deps,
	)
}
