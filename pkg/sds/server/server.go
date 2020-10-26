package server

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	xds_auth "github.com/kumahq/kuma/pkg/xds/auth"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func RegisterSDS(rt core_runtime.Runtime, server *grpc.Server) error {
	hasher := hasher{sdsServerLog}
	logger := util_xds.NewLogger(sdsServerLog)
	cache := envoy_cache.NewSnapshotCache(false, hasher, logger)

	caProvider := DefaultMeshCaProvider(rt)
	identityProvider := DefaultIdentityCertProvider(rt)
	authenticator, err := xds_server.DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := xds_auth.NewCallbacks(rt.ResourceManager(), authenticator)

	reconciler := DataplaneReconciler{
		resManager:         rt.ResourceManager(),
		readOnlyResManager: rt.ReadOnlyResourceManager(),
		meshCaProvider:     caProvider,
		identityProvider:   identityProvider,
		cache:              cache,
	}
	certGenerationsMetric := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Help: "Number of generated certificates",
		Name: "sds_cert_generation",
	}, func() float64 {
		return float64(reconciler.GeneratedCerts())
	})
	if err := rt.Metrics().Register(certGenerationsMetric); err != nil {
		return err
	}

	syncTracker, err := syncTracker(&reconciler, rt.Config().SdsServer.DataplaneConfigurationRefreshInterval, rt.Metrics())
	if err != nil {
		return err
	}
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "sds")
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		statsCallbacks,
		util_xds.LoggingCallbacks{Log: sdsServerLog},
		authCallbacks,
		syncTracker,
	}

	srv := envoy_server.NewServer(context.Background(), cache, callbacks)

	sdsServerLog.Info("registering Secret Discovery Service in Dataplane Server")
	envoy_discovery.RegisterSecretDiscoveryServiceServer(server, srv)
	return nil
}

func syncTracker(reconciler *DataplaneReconciler, refresh time.Duration, metrics core_metrics.Metrics) (envoy_server.Callbacks, error) {
	sdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "sds_generation",
		Help:       "Summary of SDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(sdsGenerations); err != nil {
		return nil, err
	}
	sdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Help: "Counter of errors during SDS generation",
		Name: "sds_generation_errors",
	})
	if err := metrics.Register(sdsGenerationsErrors); err != nil {
		return nil, err
	}
	return xds_sync.NewDataplaneSyncTracker(func(dataplaneId core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func() error {
				start := core.Now()
				defer func() {
					sdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
				}()
				return reconciler.Reconcile(dataplaneId)
			},
			OnError: func(err error) {
				sdsGenerationsErrors.Inc()
				sdsServerLog.Error(err, "OnTick() failed")
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
	proxyId, err := core_xds.ParseProxyId(node)
	if err != nil {
		h.log.Error(err, "failed to parse Proxy ID", "node", node)
		return "unknown"
	}
	return proxyId.String()
}
