package server

import (
	"context"
	"time"

	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/xds/ingress"

	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/xds/cache/cla"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"

	envoy_service_discovery_v2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_auth "k8s.io/api/authentication/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/faultinjections"
	"github.com/kumahq/kuma/pkg/core/logs"
	"github.com/kumahq/kuma/pkg/core/permissions"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	"github.com/kumahq/kuma/pkg/xds/auth"
	k8s_auth "github.com/kumahq/kuma/pkg/xds/auth/k8s"
	universal_auth "github.com/kumahq/kuma/pkg/xds/auth/universal"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	xdsServerLog  = core.Log.WithName("xds-server")
	meshResources = meshResourceTypes(map[core_model.ResourceType]bool{
		core_mesh.DataplaneInsightType:  true,
		core_mesh.DataplaneOverviewType: true,
		core_mesh.ServiceInsightType:    true,
		core_system.ConfigType:          true,
	})
)

func meshResourceTypes(exclude map[core_model.ResourceType]bool) []core_model.ResourceType {
	types := []core_model.ResourceType{}
	for _, typ := range registry.Global().ListTypes() {
		r, err := registry.Global().NewObject(typ)
		if err != nil {
			panic(err)
		}
		if r.Scope() == core_model.ScopeMesh && !exclude[typ] {
			types = append(types, typ)
		}
	}
	return types
}

func RegisterXDS(rt core_runtime.Runtime, server *grpc.Server) error {
	reconciler := DefaultReconciler(rt)

	authenticator, err := DefaultAuthenticator(rt)
	if err != nil {
		return err
	}
	authCallbacks := auth.NewCallbacks(rt.ResourceManager(), authenticator)

	metadataTracker := NewDataplaneMetadataTracker()
	lifecycle := NewDataplaneLifecycle(rt.ResourceManager())

	ingressReconciler := DefaultIngressReconciler(rt)

	connectionInfoTracker := NewConnectionInfoTracker()

	syncTracker, err := DefaultDataplaneSyncTracker(rt, reconciler, ingressReconciler, metadataTracker, connectionInfoTracker)
	if err != nil {
		return err
	}
	statusTracker, err := DefaultDataplaneStatusTracker(rt)
	if err != nil {
		return err
	}

	forcer := newResourceWarmingForcer(rt.XDS().Cache(), rt.XDS().Hasher())

	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		statsCallbacks,
		connectionInfoTracker,
		authCallbacks,
		syncTracker,
		metadataTracker,
		lifecycle,
		statusTracker,
		forcer,
	}

	srv := envoy_server.NewServer(context.Background(), rt.XDS().Cache(), callbacks)

	xdsServerLog.Info("registering Aggregated Discovery Service in Dataplane Server")
	envoy_service_discovery_v2.RegisterAggregatedDiscoveryServiceServer(server, srv)
	return nil
}

func NewKubeAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	mgr, ok := k8s_extensions.FromManagerContext(rt.Extensions())
	if !ok {
		return nil, errors.Errorf("k8s controller runtime Manager hasn't been configured")
	}
	if err := kube_auth.AddToScheme(mgr.GetScheme()); err != nil {
		return nil, errors.Wrapf(err, "could not add %q to scheme", kube_auth.SchemeGroupVersion)
	}
	return k8s_auth.New(mgr.GetClient()), nil
}

func NewUniversalAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	issuer, err := builtin.NewDataplaneTokenIssuer(rt.ReadOnlyResourceManager())
	if err != nil {
		return nil, err
	}
	return universal_auth.NewAuthenticator(issuer), nil
}

func DefaultAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	switch rt.Config().DpServer.Auth.Type {
	case dp_server.DpServerAuthServiceAccountToken:
		return NewKubeAuthenticator(rt)
	case dp_server.DpServerAuthDpToken:
		return NewUniversalAuthenticator(rt)
	case dp_server.DpServerAuthNone:
		return universal_auth.NewNoopAuthenticator(), nil
	default:
		return nil, errors.Errorf("unable to choose authenticator of %q", rt.Config().DpServer.Auth.Type)
	}
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

func DefaultIngressReconciler(rt core_runtime.Runtime) SnapshotReconciler {
	return &reconciler{
		generator: &templateSnapshotGenerator{
			ProxyTemplateResolver: &staticProxyTemplateResolver{
				template: xds_template.IngressProxyTemplate,
			},
		},
		cacher: &simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler, ingressReconciler SnapshotReconciler, metadataTracker *DataplaneMetadataTracker, connectionInfoTracker *ConnectionInfoTracker) (envoy_xds.Callbacks, error) {
	permissionsMatcher := permissions.TrafficPermissionsMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	logsMatcher := logs.TrafficLogsMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	faultInjectionMatcher := faultinjections.FaultInjectionMatcher{ResourceManager: rt.ReadOnlyResourceManager()}
	envoyCpCtx, err := xds_context.BuildControlPlaneContext(rt.Config())
	if err != nil {
		return nil, err
	}
	xdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "xds_generation",
		Help:       "Summary of XDS Snapshot generation",
		Objectives: metrics.DefaultObjectives,
	})
	if err := rt.Metrics().Register(xdsGenerations); err != nil {
		return nil, err
	}
	xdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "xds_generation_errors",
		Help: "Counter of errors during XDS generation",
	})
	if err := rt.Metrics().Register(xdsGenerationsErrors); err != nil {
		return nil, err
	}
	meshSnapshotCache, err := mesh.NewCache(rt.ReadOnlyResourceManager(),
		rt.Config().Store.Cache.ExpirationTime, meshResources, rt.LookupIP(), rt.Metrics())
	if err != nil {
		return nil, err
	}
	claCache, err := cla.NewCache(
		rt.ReadOnlyResourceManager(), rt.DataSourceLoader(), rt.Config().Multizone.Remote.Zone,
		rt.Config().Store.Cache.ExpirationTime, rt.LookupIP(), rt.Metrics())
	if err != nil {
		return nil, err
	}
	return xds_sync.NewDataplaneSyncTracker(func(key core_model.ResourceKey, streamId int64) util_watchdog.Watchdog {
		log := xdsServerLog.WithName("dataplane-sync-watchdog").WithValues("dataplaneKey", key)
		prevHash := ""
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneConfigurationRefreshInterval)
			},
			OnTick: func() error {
				start := core.Now()
				defer func() {
					xdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
				}()

				ctx := context.Background()
				proxyID := xds.FromResourceKey(key)

				// first of all calculate Hash before fetching any resources,
				// otherwise we can have lost updates
				snapshotHash, err := meshSnapshotCache.GetHash(ctx, proxyID.Mesh)
				if err != nil {
					return err
				}

				dataplane := core_mesh.NewDataplaneResource()
				if err := rt.ResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyID)
					}
					return err
				}
				resolvedDp, err := xds_topology.ResolveAddress(rt.LookupIP(), dataplane)
				if err != nil {
					return err
				}
				dataplane = resolvedDp

				// hash for Ingress should be calculated based on all dataplanes in all meshes,
				// we don't do that now, so just ignore existing `snapshotHash` and always reconcile Ingress
				if dataplane.Spec.IsIngress() {
					// update Ingress
					allMeshDataplanes := &core_mesh.DataplaneResourceList{}
					if err := rt.ReadOnlyResourceManager().List(ctx, allMeshDataplanes); err != nil {
						return err
					}
					allMeshDataplanes.Items = xds_topology.ResolveAddresses(log, rt.LookupIP(), allMeshDataplanes.Items)
					if err := ingress.UpdateAvailableServices(ctx, rt.ResourceManager(), dataplane, allMeshDataplanes.Items); err != nil {
						return err
					}
					destinations := ingress.BuildDestinationMap(dataplane)
					endpoints := ingress.BuildEndpointMap(destinations, allMeshDataplanes.Items)

					routes := &core_mesh.TrafficRouteResourceList{}
					if err := rt.ReadOnlyResourceManager().List(ctx, routes); err != nil {
						return err
					}

					proxy := xds.Proxy{
						Id:               proxyID,
						Dataplane:        dataplane,
						OutboundTargets:  endpoints,
						Metadata:         metadataTracker.Metadata(streamId),
						TrafficRouteList: routes,
					}
					envoyCtx := xds_context.Context{
						ControlPlane: envoyCpCtx,
					}
					return ingressReconciler.Reconcile(envoyCtx, &proxy)
				}

				// if previous reconciliation was without an error AND current hash is equal to previous hash
				// then we don't reconcile
				if prevHash != "" && snapshotHash == prevHash {
					return nil
				}
				log.V(1).Info("snapshot hash updated, reconcile", "prev", prevHash, "current", snapshotHash)

				meshRes := core_mesh.NewMeshResource()
				if err := rt.ReadOnlyResourceManager().Get(ctx, meshRes, core_store.GetByKey(proxyID.Mesh, core_model.NoMesh)); err != nil {
					return err
				}

				dataplanes, err := xds_topology.GetDataplanes(log, ctx, rt.ReadOnlyResourceManager(), rt.LookupIP(), dataplane.Meta.GetMesh())
				if err != nil {
					return err
				}
				externalServices := &core_mesh.ExternalServiceResourceList{}
				if err := rt.ReadOnlyResourceManager().List(ctx, externalServices, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
					return err
				}
				envoyCtx := xds_context.Context{
					ControlPlane: envoyCpCtx,
					Mesh: xds_context.MeshContext{
						Resource:   meshRes,
						Dataplanes: dataplanes,
						Hash:       snapshotHash,
					},
					ConnectionInfo: connectionInfoTracker.ConnectionInfo(streamId),
				}

				// pick a single the most specific route for each outbound interface
				routes, err := xds_topology.GetRoutes(ctx, dataplane, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				// create creates a map of selectors to match other dataplanes reachable via given routes
				destinations := xds_topology.BuildDestinationMap(dataplane, routes)

				// resolve all endpoints that match given selectors
				outbound := xds_topology.BuildEndpointMap(
					meshRes, rt.Config().Multizone.Remote.Zone,
					dataplanes.Items, externalServices.Items, rt.DataSourceLoader())

				healthChecks, err := xds_topology.GetHealthChecks(ctx, dataplane, destinations, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				circuitBreakers, err := xds_topology.GetCircuitBreakers(ctx, dataplane, destinations, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}

				trafficTrace, err := xds_topology.GetTrafficTrace(ctx, dataplane, rt.ReadOnlyResourceManager())
				if err != nil {
					return err
				}
				var tracingBackend *mesh_proto.TracingBackend
				if trafficTrace != nil {
					tracingBackend = meshRes.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
				}

				matchedPermissions, err := permissionsMatcher.Match(ctx, dataplane, meshRes)
				if err != nil {
					return err
				}

				matchedLogs, err := logsMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				faultInjection, err := faultInjectionMatcher.Match(ctx, dataplane, meshRes)
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
					CircuitBreakers:    circuitBreakers,
					Logs:               matchedLogs,
					TrafficTrace:       trafficTrace,
					TracingBackend:     tracingBackend,
					Metadata:           metadataTracker.Metadata(streamId),
					FaultInjections:    faultInjection,
					CLACache:           claCache,
				}
				err = reconciler.Reconcile(envoyCtx, &proxy)
				if err != nil {
					return err
				}
				prevHash = snapshotHash
				return nil
			},
			OnError: func(err error) {
				xdsGenerationsErrors.Inc()
				log.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				proxyID := xds.FromResourceKey(key)
				if err := reconciler.Clear(&proxyID); err != nil {
					log.Error(err, "OnStop() failed")
				}
			},
		}
	}), nil
}

func DefaultDataplaneStatusTracker(rt core_runtime.Runtime) (DataplaneStatusTracker, error) {
	tracker := NewDataplaneStatusTracker(rt, func(accessor SubscriptionStatusAccessor) DataplaneInsightSink {
		return NewDataplaneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(rt.Config().XdsServer.DataplaneStatusFlushInterval)
			},
			rt.Config().XdsServer.DataplaneStatusFlushInterval/10,
			NewDataplaneInsightStore(rt.ResourceManager()),
		)
	})
	return tracker, nil
}
