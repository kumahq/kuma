package v3

import (
	"context"
	"sync"
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

var xdsServerLog = core.Log.WithName("xds").WithName("server")

func RegisterXDS(
	statsCallbacks util_xds.StatsCallbacks,
	xdsMetrics *xds_metrics.Metrics,
	envoyCpCtx *xds_context.ControlPlaneContext,
	rt core_runtime.Runtime,
) error {
	xdsContext := NewXdsContext()

	authenticator := rt.XDS().PerProxyTypeAuthenticator()
	authCallbacks := auth.NewCallbacks(rt.ReadOnlyResourceManager(), authenticator, auth.DPNotFoundRetry{}) // no need to retry on DP Not Found because we are creating DP in DataplaneLifecycle callback

	metadataTracker := xds_callbacks.NewDataplaneMetadataTracker()
	snapshotCacheMux := &sync.Mutex{}
	reconciler := DefaultReconciler(rt, xdsContext, statsCallbacks, snapshotCacheMux)
	ingressReconciler := DefaultIngressReconciler(rt, xdsContext, statsCallbacks, snapshotCacheMux)
	egressReconciler := DefaultEgressReconciler(rt, xdsContext, statsCallbacks, snapshotCacheMux)
	watchdogFactory, err := xds_sync.DefaultDataplaneWatchdogFactory(rt, metadataTracker, reconciler, ingressReconciler, egressReconciler, xdsMetrics, envoyCpCtx, envoy_common.APIV3)
	if err != nil {
		return err
	}

	callbacks := util_xds_v3.CallbacksChain{
		util_xds_v3.NewControlPlaneIdCallbacks(rt.GetInstanceId()),
		util_xds_v3.AdaptCallbacks(statsCallbacks),
		util_xds_v3.AdaptCallbacks(authCallbacks),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(metadataTracker)),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(
			xds_callbacks.NewDataplaneLifecycle(rt.AppContext(), rt.ResourceManager(), authenticator, rt.Config().XdsServer.DataplaneDeregistrationDelay.Duration, rt.GetInstanceId(), rt.Config().Store.Cache.ExpirationTime.Duration)),
		),
		util_xds_v3.AdaptCallbacks(xds_callbacks.DataplaneCallbacksToXdsCallbacks(xds_callbacks.NewDataplaneSyncTracker(watchdogFactory.New))),
		util_xds_v3.AdaptCallbacks(DefaultDataplaneStatusTracker(rt, envoyCpCtx.Secrets)),
		util_xds_v3.AdaptCallbacks(xds_callbacks.NewNackBackoff(rt.Config().XdsServer.NACKBackoff.Duration)),
		newResourceWarmingForcer(xdsContext.Cache(), xdsContext.Hasher(), snapshotCacheMux),
	}

	if cb := rt.XDS().ServerCallbacks; cb != nil {
		callbacks = append(callbacks, util_xds_v3.AdaptCallbacks(cb))
	}

	srv := envoy_server.NewServer(context.Background(), xdsContext.Cache(), callbacks)

	xdsServerLog.Info("registering Aggregated Discovery Service V3 in Dataplane Server")
	envoy_service_discovery.RegisterAggregatedDiscoveryServiceServer(rt.DpServer().GrpcServer(), srv)
	return nil
}

func DefaultReconciler(
	rt core_runtime.Runtime,
	xdsContext XdsContext,
	statsCallbacks util_xds.StatsCallbacks,
	snapshotCacheMux *sync.Mutex,
) xds_sync.SnapshotReconciler {
	resolver := xds_template.SequentialResolver(
		&xds_template.SimpleProxyTemplateResolver{
			ReadOnlyResourceManager: rt.ReadOnlyResourceManager(),
		},
		generator.DefaultTemplateResolver,
	)

	return &reconciler{
		generator: &TemplateSnapshotGenerator{
			ResourceSetHooks:      rt.XDS().Hooks.ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		cacher:           &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
		statsCallbacks:   statsCallbacks,
		snapshotCacheMux: snapshotCacheMux,
	}
}

func DefaultIngressReconciler(
	rt core_runtime.Runtime,
	xdsContext XdsContext,
	statsCallbacks util_xds.StatsCallbacks,
	snapshotCacheMux *sync.Mutex,
) xds_sync.SnapshotReconciler {
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
		generator: &TemplateSnapshotGenerator{
			ResourceSetHooks:      rt.XDS().Hooks.ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		cacher:           &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
		statsCallbacks:   statsCallbacks,
		snapshotCacheMux: snapshotCacheMux,
	}
}

func DefaultEgressReconciler(
	rt core_runtime.Runtime,
	xdsContext XdsContext,
	statsCallbacks util_xds.StatsCallbacks,
	snapshotCacheMux *sync.Mutex,
) xds_sync.SnapshotReconciler {
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
		generator: &TemplateSnapshotGenerator{
			ResourceSetHooks:      rt.XDS().Hooks.ResourceSetHooks(),
			ProxyTemplateResolver: resolver,
		},
		cacher:           &simpleSnapshotCacher{xdsContext.Hasher(), xdsContext.Cache()},
		statsCallbacks:   statsCallbacks,
		snapshotCacheMux: snapshotCacheMux,
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
					return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval.Duration)
				},
				func() *time.Ticker {
					return time.NewTicker(rt.Config().Metrics.Dataplane.IdleTimeout.Duration / 2)
				},
				rt.Config().XdsServer.DataplaneStatusFlushInterval.Duration/10,
				xds_callbacks.NewDataplaneInsightStore(rt.ResourceManager()),
			)
		})
}
