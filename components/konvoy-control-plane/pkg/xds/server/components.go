package server

import (
	"context"
	"time"

	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	util_watchdog "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/watchdog"
	xds_sync "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/sync"
	xds_template "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"

	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func DefaultReconciler(rt core_runtime.Runtime) *core_discovery.DiscoverySink {
	return &core_discovery.DiscoverySink{
		DataplaneConsumer: &reconciler{
			&templateSnapshotGenerator{
				ProxyTemplateResolver: &simpleProxyTemplateResolver{
					ResourceStore:        rt.ResourceStore(),
					DefaultProxyTemplate: xds_template.DefaultProxyTemplate,
				},
			},
			&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
		},
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, ds *core_discovery.DiscoverySink) envoy_xds.Callbacks {
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				ctx := context.Background()
				dataplane := &mesh_core.DataplaneResource{}
				if err := rt.ResourceStore().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return ds.OnDataplaneDelete(key)
					}
					return err
				}
				return ds.OnDataplaneUpdate(dataplane)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}
	})
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) DataplaneStatusTracker {
	return NewDataplaneStatusTracker(rt, func(accessor SubscriptionStatusAccessor) DataplaneInsightSink {
		return NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			NewDataplaneInsightStore(rt.ResourceStore()))
	})
}
