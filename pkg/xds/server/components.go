package server

import (
	"context"
	"time"

	"github.com/Kong/kuma/pkg/core/faultinjections"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/logs"
	"github.com/Kong/kuma/pkg/core/permissions"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/xds"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	xds_bootstrap "github.com/Kong/kuma/pkg/xds/bootstrap"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	xds_sync "github.com/Kong/kuma/pkg/xds/sync"
	xds_template "github.com/Kong/kuma/pkg/xds/template"
	xds_topology "github.com/Kong/kuma/pkg/xds/topology"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	reconciler := DefaultReconciler(rt)

	metadataTracker := NewDataplaneMetadataTracker()

	tracker, err := DefaultDataplaneSyncTracker(rt, reconciler, metadataTracker)
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		tracker,
		metadataTracker,
		DefaultDataplaneStatusTracker(rt),
	}

	srv := NewServer(rt.XDS().Cache(), callbacks)
	return rt.Add(
		// xDS gRPC API
		&grpcServer{srv, rt.Config().XdsServer.GrpcPort},
		// diagnostics server
		&diagnosticsServer{rt.Config().XdsServer.DiagnosticsPort},
		// bootstrap server
		&xds_bootstrap.BootstrapServer{
			Port:      rt.Config().BootstrapServer.Port,
			Generator: xds_bootstrap.NewDefaultBootstrapGenerator(rt.ResourceManager(), rt.Config().BootstrapServer.Params),
		},
	)
}

func DefaultReconciler(rt core_runtime.Runtime) SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &simpleProxyTemplateResolver{
				ReadOnlyResourceManager: rt.ReadOnlyResourceManager(),
				DefaultProxyTemplate:    xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler SnapshotReconciler, metadataTracker *DataplaneMetadataTracker) (envoy_xds.Callbacks, error) {
	permissionsMatcher := permissions.TrafficPermissionsMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	logsMatcher := logs.TrafficLogsMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	faultInjectionMatcher := faultinjections.FaultInjectionMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	envoyCpCtx, err := xds_context.BuildControlPlaneContext(rt.Config())
	if err != nil {
		return nil, err
	}
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				ctx := context.Background()
				dataplane := &mesh_core.DataplaneResource{}
				proxyID := xds.FromResourceKey(key)

				if err := rt.ReadOnlyResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyID)
					}
					return err
				}

				mesh := &mesh_core.MeshResource{}
				if err := rt.ReadOnlyResourceManager().Get(ctx, mesh, core_store.GetByKey(proxyID.Mesh, proxyID.Mesh)); err != nil {
					return err
				}
				envoyCtx := xds_context.Context{
					ControlPlane: envoyCpCtx,
					Mesh: xds_context.MeshContext{
						Resource: mesh,
					},
				}

				// pick a single the most specific route for each outbound interface
				routes, err := xds_topology.GetRoutes(ctx, dataplane, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				// create creates a map of selectors to match other dataplanes reachable via given routes
				destinations := xds_topology.BuildDestinationMap(dataplane, routes)

				// resolve all endpoints that match given selectors
				outbound, err := xds_topology.GetOutboundTargets(ctx, dataplane, destinations, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				healthChecks, err := xds_topology.GetHealthChecks(ctx, dataplane, destinations, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				trafficTrace, err := xds_topology.GetTrafficTrace(ctx, dataplane, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}
				var tracingBackend *mesh_proto.TracingBackend
				if trafficTrace != nil {
					tracingBackend = mesh.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
				}

				matchedPermissions, err := permissionsMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				matchedLogs, err := logsMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				faultInjection, err := faultInjectionMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				proxy := xds.Proxy{
					Id:                 proxyID,
					Dataplane:          dataplane,
					TrafficPermissions: matchedPermissions,
					TrafficRoutes:      routes,
					OutboundSelectors:  destinations,
					OutboundTargets:    outbound,
					HealthChecks:       healthChecks,
					Logs:               matchedLogs,
					TrafficTrace:       trafficTrace,
					TracingBackend:     tracingBackend,
					Metadata:           metadataTracker.Metadata(streamId),
					FaultInjection:     faultInjection,
				}
				return reconciler.Reconcile(envoyCtx, &proxy)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}
	}), nil
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
