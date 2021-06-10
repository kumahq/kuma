package v2

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	sds_ca "github.com/kumahq/kuma/pkg/sds/ca"
	sds_identity "github.com/kumahq/kuma/pkg/sds/identity"
	sds_metrics "github.com/kumahq/kuma/pkg/sds/metrics"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
	auth_components "github.com/kumahq/kuma/pkg/xds/auth/components"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	xds_callbacks "github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func RegisterSDS(rt core_runtime.Runtime, sdsMetrics *sds_metrics.Metrics) error {
	hasher := hasher{sdsServerLog}
	logger := util_xds.NewLogger(sdsServerLog)
	cache := envoy_cache.NewSnapshotCache(false, hasher, logger)

	caProvider := sds_ca.NewProvider(rt.ResourceManager(), rt.CaManagers())
	identityProvider := sds_identity.NewProvider(rt.ResourceManager(), rt.CaManagers())
	authenticator, err := auth_components.DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := xds_auth.NewCallbacks(rt.ResourceManager(), authenticator, xds_auth.DPNotFoundRetry{}) // no need to retry on DP Not Found because we are creating DP in ADS before we initiate SDS

	reconciler := DataplaneReconciler{
		resManager:         rt.ResourceManager(),
		readOnlyResManager: rt.ReadOnlyResourceManager(),
		meshCaProvider:     caProvider,
		identityProvider:   identityProvider,
		cache:              cache,
		upsertConfig:       rt.Config().Store.Upsert,
		sdsMetrics:         sdsMetrics,
		proxySnapshotInfo:  map[string]snapshotInfo{},
	}

	syncTracker, err := syncTracker(&reconciler, rt.Config().SdsServer.DataplaneConfigurationRefreshInterval, sdsMetrics)
	if err != nil {
		return err
	}
	callbacks := util_xds_v2.CallbacksChain{
		util_xds_v2.AdaptCallbacks(sdsMetrics.Callbacks),
		util_xds_v2.AdaptCallbacks(util_xds.LoggingCallbacks{Log: sdsServerLog}),
		util_xds_v2.AdaptCallbacks(authCallbacks),
		util_xds_v2.AdaptCallbacks(syncTracker),
	}

	srv := envoy_server.NewServer(context.Background(), cache, callbacks)

	sdsServerLog.Info("registering Secret Discovery Service V2 in Dataplane Server")
	envoy_discovery.RegisterSecretDiscoveryServiceServer(rt.DpServer().GrpcServer(), srv)
	return nil
}

func syncTracker(reconciler *DataplaneReconciler, refresh time.Duration, sdsMetrics *sds_metrics.Metrics) (util_xds.Callbacks, error) {
	return xds_callbacks.NewDataplaneSyncTracker(func(dataplaneId core_model.ResourceKey, streamId int64, proxyType mesh_proto.DpType) util_watchdog.Watchdog {
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				sdsMetrics.IncrementActiveWatchdogs(envoy_common.APIV2)
				return time.NewTicker(refresh)
			},
			OnTick: func() error {
				start := core.Now()
				defer func() {
					sdsMetrics.SdsGeneration(envoy_common.APIV2).Observe(float64(core.Now().Sub(start).Milliseconds()))
				}()
				return reconciler.Reconcile(dataplaneId, proxyType)
			},
			OnError: func(err error) {
				sdsMetrics.SdsGenerationsErrors(envoy_common.APIV2).Inc()
				sdsServerLog.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				if err := reconciler.Cleanup(dataplaneId, proxyType); err != nil {
					sdsServerLog.Error(err, "could not cleanup sync", "dataplane", dataplaneId)
				}
				sdsMetrics.DecrementActiveWatchdogs(envoy_common.APIV2)
			},
		}
	}), nil
}

type hasher struct {
	log logr.Logger
}

func (h hasher) ID(node *envoy_core.Node) string {
	if node == nil {
		return "unknown"
	}
	proxyId, err := core_xds.ParseProxyIdFromString(node.GetId())
	if err != nil {
		h.log.Error(err, "failed to parse Proxy ID", "node", node)
		return "unknown"
	}
	return proxyId.String()
}
