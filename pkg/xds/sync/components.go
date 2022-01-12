package sync

import (
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/dns/lookup"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	"github.com/kumahq/kuma/pkg/core/ratelimits"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
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
	cachingRm manager.ReadOnlyResourceManager,
	nonCachingRm manager.ResourceManager,
	lookupIP lookup.LookupIPFunc,
	dataSourceLoader datasource.Loader,
	config kuma_cp.Config,
	configManager config_manager.ConfigManager,
	metadataTracker DataplaneMetadataTracker,
	apiVersion envoy.APIVersion,
) *DataplaneProxyBuilder {
	return &DataplaneProxyBuilder{
		CachingResManager:     cachingRm,
		NonCachingResManager:  nonCachingRm,
		LookupIP:              lookupIP,
		DataSourceLoader:      dataSourceLoader,
		MetadataTracker:       metadataTracker,
		PermissionMatcher:     permissions.TrafficPermissionsMatcher{ResourceManager: cachingRm},
		LogsMatcher:           logs.TrafficLogsMatcher{ResourceManager: cachingRm},
		FaultInjectionMatcher: faultinjections.FaultInjectionMatcher{ResourceManager: cachingRm},
		RateLimitMatcher:      ratelimits.RateLimitMatcher{ResourceManager: cachingRm},
		Zone:                  config.Multizone.Zone.Name,
		APIVersion:            apiVersion,
		ConfigManager:         configManager,
		TopLevelDomain:        config.DNSServer.Domain,
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
	dataplaneReconciler SnapshotReconciler,
	ingressReconciler SnapshotReconciler,
	xdsMetrics *xds_metrics.Metrics,
	meshSnapshotCache *mesh.Cache,
	envoyCpCtx *xds_context.ControlPlaneContext,
	apiVersion envoy.APIVersion,
) (DataplaneWatchdogFactory, error) {
	dataplaneProxyBuilder := DefaultDataplaneProxyBuilder(
		rt.ReadOnlyResourceManager(),
		rt.ResourceManager(),
		rt.LookupIP(),
		rt.DataSourceLoader(),
		rt.Config(),
		rt.ConfigManager(),
		metadataTracker,
		apiVersion)
	ingressProxyBuilder := defaultIngressProxyBuilder(rt, metadataTracker, apiVersion)

	deps := DataplaneWatchdogDependencies{
		dataplaneProxyBuilder: dataplaneProxyBuilder,
		dataplaneReconciler:   dataplaneReconciler,
		ingressProxyBuilder:   ingressProxyBuilder,
		ingressReconciler:     ingressReconciler,
		envoyCpCtx:            envoyCpCtx,
		meshCache:             meshSnapshotCache,
		metadataTracker:       metadataTracker,
	}
	return NewDataplaneWatchdogFactory(
		xdsMetrics,
		rt.Config().XdsServer.DataplaneConfigurationRefreshInterval,
		deps,
	)
}
