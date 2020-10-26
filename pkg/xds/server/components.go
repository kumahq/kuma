package server

import (
	"context"
	"time"

	"google.golang.org/grpc"

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
	config_core "github.com/kumahq/kuma/pkg/config/core"
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
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	xdsServerLog  = core.Log.WithName("xds-server")
	meshResources = meshResourceTypes(map[core_model.ResourceType]bool{
		core_mesh.DataplaneInsightType:  true,
		core_mesh.DataplaneOverviewType: true,
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

	syncTracker, err := DefaultDataplaneSyncTracker(rt, reconciler, ingressReconciler, metadataTracker)
	if err != nil {
		return err
	}
	statusTracker, err := DefaultDataplaneStatusTracker(rt)
	if err != nil {
		return err
	}

	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		statsCallbacks,
		authCallbacks,
		syncTracker,
		metadataTracker,
		lifecycle,
		statusTracker,
	}

	srv := NewServer(rt.XDS().Cache(), callbacks)

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
	if !rt.Config().AdminServer.Apis.DataplaneToken.Enabled {
		return universal_auth.NewNoopAuthenticator(), nil
	}
	issuer, err := builtin.NewDataplaneTokenIssuer(rt)
	if err != nil {
		return nil, err
	}
	return universal_auth.NewAuthenticator(issuer), nil
}

func DefaultAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	switch env := rt.Config().Environment; env {
	case config_core.KubernetesEnvironment:
		return NewKubeAuthenticator(rt)
	case config_core.UniversalEnvironment:
		return NewUniversalAuthenticator(rt)
	default:
		return nil, errors.Errorf("unable to choose SDS authenticator for environment type %q", env)
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

func DefaultDataplaneSyncTracker(rt core_runtime.Runtime, reconciler, ingressReconciler SnapshotReconciler, metadataTracker *DataplaneMetadataTracker) (envoy_xds.Callbacks, error) {
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
	claCache, err := cla.NewCache(rt.ReadOnlyResourceManager(), rt.Config().Multicluster.Remote.Zone,
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
				dataplane := &core_mesh.DataplaneResource{}
				proxyID := xds.FromResourceKey(key)

				if err := rt.ReadOnlyResourceManager().Get(ctx, dataplane, core_store.GetBy(key)); err != nil {
					if core_store.IsResourceNotFound(err) {
						return reconciler.Clear(&proxyID)
					}
					return err
				}

				if err := xds_topology.ResolveAddress(rt.LookupIP(), dataplane); err != nil {
					return err
				}

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
					proxy := xds.Proxy{
						Id:              proxyID,
						Dataplane:       dataplane,
						OutboundTargets: endpoints,
						Metadata:        metadataTracker.Metadata(streamId),
					}
					envoyCtx := xds_context.Context{
						ControlPlane: envoyCpCtx,
					}
					return ingressReconciler.Reconcile(envoyCtx, &proxy)
				}

				snapshotHash, err := meshSnapshotCache.GetHash(ctx, dataplane.GetMeta().GetMesh())
				if err != nil {
					return err
				}
				if prevHash != "" && snapshotHash == prevHash {
					return nil
				}
				log.V(1).Info("snapshot hash updated, reconcile", "prev", prevHash, "current", snapshotHash)
				prevHash = snapshotHash

				mesh := &core_mesh.MeshResource{}
				if err := rt.ReadOnlyResourceManager().Get(ctx, mesh, core_store.GetByKey(proxyID.Mesh, proxyID.Mesh)); err != nil {
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
						Resource:   mesh,
						Dataplanes: dataplanes,
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
				outbound := xds_topology.BuildEndpointMap(dataplanes.Items, rt.Config().Multicluster.Remote.Zone, mesh, externalServices.Items)

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
					tracingBackend = mesh.GetTracingBackend(trafficTrace.Spec.GetConf().GetBackend())
				}

				matchedPermissions, err := permissionsMatcher.Match(ctx, dataplane, mesh)
				if err != nil {
					return err
				}

				matchedLogs, err := logsMatcher.Match(ctx, dataplane)
				if err != nil {
					return err
				}

				faultInjection, err := faultInjectionMatcher.Match(ctx, dataplane, mesh)
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
				return reconciler.Reconcile(envoyCtx, &proxy)
			},
			OnError: func(err error) {
				xdsGenerationsErrors.Inc()
				log.Error(err, "OnTick() failed")
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
			NewDataplaneInsightStore(rt.ResourceManager()))
	})
	return tracker, nil
}
