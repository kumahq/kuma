package v2

import (
	"context"
	"time"

	envoy_service_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	v2 "github.com/kumahq/kuma/pkg/core/xds/v2"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	"github.com/kumahq/kuma/pkg/xds/auth"
	auth_components "github.com/kumahq/kuma/pkg/xds/auth/components"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
	xds_callbacks "github.com/kumahq/kuma/pkg/xds/server/callbacks"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
)

var xdsServerLog = core.Log.WithName("xds-server")

func RegisterXDS(
	statsCallbacks util_xds.Callbacks,
	xdsMetrics *xds_metrics.Metrics,
	meshSnapshotCache *mesh.Cache,
	envoyCpCtx *xds_context.ControlPlaneContext,
	rt core_runtime.Runtime,
	server *grpc.Server,
) error {
	xdsContext := v2.NewXdsContext()

	authenticator, err := auth_components.DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := auth.NewCallbacks(rt.ResourceManager(), authenticator, auth.DPNotFoundRetry{}) // no need to retry on DP Not Found because we are creating DP in DataplaneLifecycle callback

	metadataTracker := xds_callbacks.NewDataplaneMetadataTracker()
	connectionInfoTracker := xds_callbacks.NewConnectionInfoTracker()
	reconciler := DefaultReconciler(rt, xdsContext)
	ingressReconciler := DefaultIngressReconciler(rt, xdsContext)
	watchdogFactory, err := xds_sync.DefaultDataplaneWatchdogFactory(rt, metadataTracker, connectionInfoTracker, reconciler, ingressReconciler, xdsMetrics, meshSnapshotCache, envoyCpCtx, envoy_common.APIV2)
	if err != nil {
		return err
	}

	callbacks := util_xds_v2.CallbacksChain{
		util_xds_v2.NewControlPlaneIdCallbacks(rt.GetInstanceId()),
		util_xds_v2.AdaptCallbacks(statsCallbacks),
		util_xds_v2.AdaptCallbacks(connectionInfoTracker),
		util_xds_v2.AdaptCallbacks(authCallbacks),
		util_xds_v2.AdaptCallbacks(xds_callbacks.NewDataplaneSyncTracker(watchdogFactory.New)),
		util_xds_v2.AdaptCallbacks(metadataTracker),
		util_xds_v2.AdaptCallbacks(xds_callbacks.NewDataplaneLifecycle(rt.ResourceManager())),
		util_xds_v2.AdaptCallbacks(DefaultDataplaneStatusTracker(rt)),
		newResourceWarmingForcer(xdsContext.Cache(), xdsContext.Hasher()),
	}

	srv := envoy_server.NewServer(context.Background(), xdsContext.Cache(), callbacks)

	xdsServerLog.Info("registering Aggregated Discovery Service V2 in Dataplane Server")
	envoy_service_discovery.RegisterAggregatedDiscoveryServiceServer(server, srv)
	return nil
}

func DefaultReconciler(rt core_runtime.Runtime, xdsContext v2.XdsContext) xds_sync.SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &xds_template.SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: rt.ReadOnlyResourceManager(),
				DefaultProxyTemplate:    xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
	}
}

func DefaultIngressReconciler(rt core_runtime.Runtime, xdsContext v2.XdsContext) xds_sync.SnapshotReconciler {
	return &reconciler{
		generator: &templateSnapshotGenerator{
			ProxyTemplateResolver: &xds_template.StaticProxyTemplateResolver{
				Template: xds_template.IngressProxyTemplate,
			},
		},
		cacher: &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
	}
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) xds_callbacks.DataplaneStatusTracker {
	return xds_callbacks.NewDataplaneStatusTracker(rt, func(accessor xds_callbacks.SubscriptionStatusAccessor) xds_callbacks.DataplaneInsightSink {
		return xds_callbacks.NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			rt.Config().XdsServer.DataplaneStatusFlushInterval/10,
			xds_callbacks.NewDataplaneInsightStore(rt.ResourceManager()),
		)
	})
}
