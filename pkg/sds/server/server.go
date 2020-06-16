package server

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	"github.com/Kong/kuma/pkg/core"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	xds_sync "github.com/Kong/kuma/pkg/xds/sync"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher := hasher{sdsServerLog}
	logger := util_xds.NewLogger(sdsServerLog)
	cache := envoy_cache.NewSnapshotCache(false, hasher, logger)

	caProvider := DefaultMeshCaProvider(rt)
	identityProvider := DefaultIdentityCertProvider(rt)
	authenticator, err := DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := newAuthCallbacks(authenticator)

	reconciler := DataplaneReconciler{
		resManager:         rt.ResourceManager(),
		readOnlyResManager: rt.ReadOnlyResourceManager(),
		meshCaProvider:     caProvider,
		identityProvider:   identityProvider,
		cache:              cache,
	}
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: sdsServerLog},
		authCallbacks,
		syncTracker(reconciler, rt.Config().SdsServer.DataplaneConfigurationRefreshInterval),
	}

	srv := envoy_server.NewServer(context.Background(), cache, callbacks)

	return rt.Add(&grpcServer{srv, *rt.Config().SdsServer})
}

func syncTracker(reconciler DataplaneReconciler, refresh time.Duration) envoy_server.Callbacks {
	return xds_sync.NewDataplaneSyncTracker(func(dataplaneId core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func() error {
				return reconciler.Reconcile(dataplaneId)
			},
			OnError: func(err error) {
				sdsServerLog.Error(err, "OnTick() failed")
			},
		}
	})
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
