package sync

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func defaultDataplaneProxyBuilder(rt core_runtime.Runtime, metadataTracker DataplaneMetadataTracker, apiVersion envoy.APIVersion) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		CachingResManager:     rt.ReadOnlyResourceManager(),
		NonCachingResManager:  rt.ResourceManager(),
		LookupIP:              rt.LookupIP(),
		DataSourceLoader:      rt.DataSourceLoader(),
		MetadataTracker:       metadataTracker,
		PermissionMatcher:     permissions.TrafficPermissionsMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		LogsMatcher:           logs.TrafficLogsMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		FaultInjectionMatcher: faultinjections.FaultInjectionMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		RateLimitMatcher:      ratelimits.RateLimitMatcher{ResourceManager: rt.ReadOnlyResourceManager()},
		Zone:                  rt.Config().Multizone.Zone.Name,
		APIVersion:            apiVersion,
		ConfigManager:         rt.ConfigManager(),
		TopLevelDomain:        rt.Config().DNSServer.Domain,
	}
}

func defaultIngressProxyBuilder(rt core_runtime.Runtime, metadataTracker DataplaneMetadataTracker, apiVersion envoy.APIVersion) *IngressProxyBuilder {
	return &IngressProxyBuilder{
		ResManager:         rt.ResourceManager(),
		ReadOnlyResManager: rt.ReadOnlyResourceManager(),
		LookupIP:           rt.LookupIP(),
		MetadataTracker:    metadataTracker,
		apiVersion:         apiVersion,
	}
}

func DefaultDataplaneWatchdogFactory(
	rt core_runtime.Runtime,
	metadataTracker DataplaneMetadataTracker,
	connectionInfoTracker ConnectionInfoTracker,
	dataplaneReconciler SnapshotReconciler,
	ingressReconciler SnapshotReconciler,
	xdsMetrics *xds_metrics.Metrics,
	meshSnapshotCache *mesh.Cache,
	envoyCpCtx *xds_context.ControlPlaneContext,
	apiVersion envoy.APIVersion,
) (DataplaneWatchdogFactory, error) {
	dataplaneProxyBuilder := defaultDataplaneProxyBuilder(rt, metadataTracker, apiVersion)
	ingressProxyBuilder := defaultIngressProxyBuilder(rt, metadataTracker, apiVersion)
	xdsContextBuilder := newXDSContextBuilder(envoyCpCtx, connectionInfoTracker, rt.ReadOnlyResourceManager(), rt.LookupIP(), rt.EnvoyAdminClient())

	deps := DataplaneWatchdogDependencies{
		dataplaneProxyBuilder: dataplaneProxyBuilder,
		dataplaneReconciler:   dataplaneReconciler,
		ingressProxyBuilder:   ingressProxyBuilder,
		ingressReconciler:     ingressReconciler,
		xdsContextBuilder:     xdsContextBuilder,
		meshCache:             meshSnapshotCache,
		metadataTracker:       metadataTracker,
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		rt.Config().XdsServer.DataplaneConfigurationRefreshInterval,
		deps,
	)
}
