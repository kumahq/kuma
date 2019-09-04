package server

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	"time"

	//core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	util_watchdog "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/watchdog"
	xds_sync "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/sync"
	xds_template "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/template"

	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func DefaultReconciler(rt core_runtime.Runtime) SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ResourceManager:      rt.ResourceManager(),
				DefaultProxyTemplate: xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler SnapshotReconciler) envoy_xds.Callbacks {
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				ctx := context.Background()
				dataplane := &mesh_core.DataplaneResource{}
				proxyId := xds.ProxyId{
					Name:      key.Name,
					Namespace: key.Namespace,
					Mesh:      key.Mesh,
				}
				if err := rt.ResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyId)
					}
					return err
				}
				proxy := xds.Proxy{
					Id:        proxyId,
					Dataplane: dataplane,
				}
				return reconciler.Reconcile(&proxy)
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
			NewDataplaneInsightStore(rt.ResourceManager()))
	})
}
