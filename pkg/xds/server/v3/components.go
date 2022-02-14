package v3

import (
	"context"
	"time"

	envoy_service_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/auth"
	auth_components "github.com/kumahq/kuma/pkg/xds/auth/components"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
	xds_metrics "github.com/kumahq/kuma/pkg/xds/metrics"
	"github.com/kumahq/kuma/pkg/xds/secrets"
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
) error {
	xdsContext := NewXdsContext()

	authenticator, err := auth_components.DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := auth.NewCallbacks(rt.ReadOnlyResourceManager(), authenticator, auth.DPNotFoundRetry{}) // no need to retry on DP Not Found because we are creating DP in DataplaneLifecycle callback

	metadataTracker := xds_callbacks.NewDataplaneMetadataTracker()
	reconciler := DefaultReconciler(rt, xdsContext)
	ingressReconciler := DefaultIngressReconciler(rt, xdsContext)
	egressReconciler := DefaultEgressReconciler(rt, xdsContext)
	watchdogFactory, err := xds_sync.DefaultDataplaneWatchdogFactory(rt, metadataTracker, reconciler, ingressReconciler, egressReconciler, xdsMetrics, meshSnapshotCache, envoyCpCtx, envoy_common.APIV3)
	if err != nil {
		return err
	}

	callbacks := util_xds_v3.CallbacksChain{
		util_xds_v3.NewControlPlaneIdCallbacks(rt.GetInstanceId()),
		util_xds_v3.AdaptCallbacks(statsCallbacks),
		util_xds_v3.AdaptCallbacks(authCallbacks),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(xds_callbacks.NewDataplaneSyncTracker(watchdogFactory.New))),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(metadataTracker)),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(xds_callbacks.NewDataplaneLifecycle(rt.AppContext(), rt.ResourceManager(), authenticator))),
		util_xds_v3.AdaptCallbacks(DefaultDataplaneStatusTracker(rt, envoyCpCtx.Secrets)),
		util_xds_v3.AdaptCallbacks(xds_callbacks.NewNackBackoff(rt.Config().XdsServer.NACKBackoff)),
		newResourceWarmingForcer(xdsContext.Cache(), xdsContext.Hasher()),
	}

	srv := envoy_server.NewServer(context.Background(), xdsContext.Cache(), callbacks)

	xdsServerLog.Info("registering Aggregated Discovery Service V3 in Dataplane Server")
	envoy_service_discovery.RegisterAggregatedDiscoveryServiceServer(rt.DpServer().GrpcServer(), srv)
	return nil
}

func DefaultReconciler(rt core_runtime.Runtime, xdsContext XdsContext) xds_sync.SnapshotReconciler {
	resolver := xds_template.SequentialResolver(
		&xds_template.SimpleProxyTemplateResolver{
			ReadOnlyResourceManager: rt.ReadOnlyResourceManager(),
		},
		generator.DefaultTemplateResolver,
	)

	return &reconciler{
		&templateSnapshotGenerator{
			ResourceSetHooks:      rt.XDSHooks().ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		&simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
	}
}

func DefaultIngressReconciler(rt core_runtime.Runtime, xdsContext XdsContext) xds_sync.SnapshotReconciler {
	resolver := &xds_template.StaticProxyTemplateResolver{
		Template: &mesh_proto.ProxyTemplate{
			Conf: &mesh_proto.ProxyTemplate_Conf{
				Imports: []string{
					generator.IngressProxy,
				},
			},
		},
	}

	return &reconciler{
		generator: &templateSnapshotGenerator{
			ResourceSetHooks:      rt.XDSHooks().ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		cacher: &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
	}
}

func DefaultEgressReconciler(rt core_runtime.Runtime, xdsContext XdsContext) xds_sync.SnapshotReconciler {
	resolver := &xds_template.StaticProxyTemplateResolver{
		Template: &mesh_proto.ProxyTemplate{
			Conf: &mesh_proto.ProxyTemplate_Conf{
				Imports: []string{
					egress.EgressProxy,
				},
			},
		},
	}

	return &reconciler{
		generator: &templateSnapshotGenerator{
			ResourceSetHooks:      rt.XDSHooks().ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		cacher: &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
	}
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime, secrets secrets.Secrets) xds_callbacks.DataplaneStatusTracker {
	return xds_callbacks.NewDataplaneStatusTracker(rt,
		func(dataplaneType core_model.ResourceType, accessor xds_callbacks.SubscriptionStatusAccessor) xds_callbacks.DataplaneInsightSink {
			return xds_callbacks.NewDataplaneInsightSink(
				dataplaneType,
				accessor,
				secrets,
				func() *time.Ticker {
					return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
				},
				func() *time.Ticker {
					return time.NewTicker(rt.Config().Metrics.Dataplane.IdleTimeout / 2)
				},
				rt.Config().XdsServer.DataplaneStatusFlushInterval/10,
				xds_callbacks.NewDataplaneInsightStore(rt.ResourceManager()),
			)
		})
}
