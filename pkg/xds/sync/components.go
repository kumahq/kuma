package sync

import (
	"context"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func DefaultDataplaneProxyBuilder(
	config kuma_cp.Config,
	metadataTracker DataplaneMetadataTracker,
	apiVersion envoy.APIVersion,
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
	apiVersion envoy.APIVersion,
	meshCache *mesh.Cache,
) *IngressProxyBuilder {
	return &IngressProxyBuilder{
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
		apiVersion:         apiVersion,
		meshCache:          meshCache,
		zone:               rt.Config().Multizone.Zone.Name,
	}
}

func DefaultEgressProxyBuilder(
	ctx context.Context,
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	meshCache *mesh.Cache,
	apiVersion envoy.APIVersion,
) *EgressProxyBuilder {
	return &EgressProxyBuilder{
		ctx:                ctx,
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
		meshCache:          meshCache,
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
	meshSnapshotCache *mesh.Cache,
	envoyCpCtx *xds_context.ControlPlaneContext,
	apiVersion envoy.APIVersion,
) (DataplaneWatchdogFactory, error) {
	ctx := context.Background()
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
		meshSnapshotCache,
	)

	egressProxyBuilder := DefaultEgressProxyBuilder(
		ctx,
		rt,
		metadataTracker,
		meshSnapshotCache,
		apiVersion,
	)

	deps := DataplaneWatchdogDependencies{
		dataplaneProxyBuilder: dataplaneProxyBuilder,
		dataplaneReconciler:   dataplaneReconciler,
		ingressProxyBuilder:   ingressProxyBuilder,
		ingressReconciler:     ingressReconciler,
		egressProxyBuilder:    egressProxyBuilder,
		egressReconciler:      egressReconciler,
		envoyCpCtx:            envoyCpCtx,
		meshCache:             meshSnapshotCache,
		metadataTracker:       metadataTracker,
		resManager:            rt.ReadOnlyResourceManager(),
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		config.XdsServer.DataplaneConfigurationRefreshInterval,
		deps,
	)
}
