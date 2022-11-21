package sync

import (
	"context"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/user"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func DefaultDataplaneProxyBuilder(
	config kuma_cp.Config,
	metadataTracker DataplaneMetadataTracker,
	apiVersion core_xds.APIVersion,
) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		MetadataTracker: metadataTracker,
		Zone:            config.Multizone.Zone.Name,
		APIVersion:      apiVersion,
	}
}

func DefaultIngressProxyBuilder(
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	apiVersion core_xds.APIVersion,
) *IngressProxyBuilder {
	return &IngressProxyBuilder{
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
		apiVersion:         apiVersion,
		meshCache:          rt.MeshCache(),
		zone:               rt.Config().Multizone.Zone.Name,
	}
}

func DefaultEgressProxyBuilder(
	ctx context.Context,
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	apiVersion core_xds.APIVersion,
) *EgressProxyBuilder {
	return &EgressProxyBuilder{
		ctx:                ctx,
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
		meshCache:          rt.MeshCache(),
		apiVersion:         apiVersion,
		zone:               rt.Config().Multizone.Zone.Name,
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
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	config := rt.Config()

	dataplaneProxyBuilder := DefaultDataplaneProxyBuilder(
		config,
		metadataTracker,
		apiVersion,
	)

	ingressProxyBuilder := DefaultIngressProxyBuilder(
		rt,
		metadataTracker,
		apiVersion,
	)

	egressProxyBuilder := DefaultEgressProxyBuilder(
		ctx,
		rt,
		metadataTracker,
		apiVersion,
	)

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
