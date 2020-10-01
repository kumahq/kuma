package server

import (
	"context"
	"time"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	kube_auth "k8s.io/api/authentication/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/metrics"
	k8s_runtime "github.com/kumahq/kuma/pkg/runtime/k8s"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	"github.com/kumahq/kuma/pkg/xds/auth"
	k8s_auth "github.com/kumahq/kuma/pkg/xds/auth/k8s"
	universal_auth "github.com/kumahq/kuma/pkg/xds/auth/universal"
	xds_bootstrap "github.com/kumahq/kuma/pkg/xds/bootstrap"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/ingress"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

var (
	xdsServerLog  = core.Log.WithName("xds-server")
	meshResources = []core_model.ResourceType{
		mesh_core.DataplaneType,
		mesh_core.DataplaneInsightType,
		mesh_core.CircuitBreakerType,
		mesh_core.FaultInjectionType,
		mesh_core.HealthCheckType,
		mesh_core.TrafficLogType,
		mesh_core.TrafficPermissionType,
		mesh_core.TrafficRouteType,
		mesh_core.TrafficTraceType,
		mesh_core.ProxyTemplateType,
	}
)

func SetupServer(rt core_runtime.Runtime) error {
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
	return rt.Add(
		// xDS gRPC API
		&grpcServer{
			server:      srv,
			port:        rt.Config().XdsServer.GrpcPort,
			tlsCertFile: rt.Config().XdsServer.TlsCertFile,
			tlsKeyFile:  rt.Config().XdsServer.TlsKeyFile,
			metrics:     rt.Metrics(),
		},
		// bootstrap server
		&xds_bootstrap.BootstrapServer{
			Config:    rt.Config().BootstrapServer,
			Generator: xds_bootstrap.NewDefaultBootstrapGenerator(rt.ResourceManager(), rt.Config().BootstrapServer.Params, rt.Config().XdsServer.TlsCertFile),
			Metrics:   rt.Metrics(),
		},
	)
}

func NewKubeAuthenticator(rt core_runtime.Runtime) (auth.Authenticator, error) {
	mgr, ok := k8s_runtime.FromManagerContext(rt.Extensions())
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

				dataplane := &mesh_core.DataplaneResource{}
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
					allMeshDataplanes := &mesh_core.DataplaneResourceList{}
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

				snapshot, err := GetMeshSnapshot(ctx, dataplane.GetMeta().GetMesh(), rt.ReadOnlyResourceManager(), meshResources, rt.LookupIP())
				if err != nil {
					return err
				}

				snapshotHash := snapshot.Hash()
				if prevHash != "" && snapshotHash == prevHash {
					log.V(1).Info("snapshot hashes are equal, no need to reconcile", "hash", prevHash)
					return nil
				}
				prevHash = snapshotHash
				log.V(1).Info("snapshot hash updated, reconcile", "hash", prevHash)

				envoyCtx := xds_context.Context{
					ControlPlane: envoyCpCtx,
					Mesh: xds_context.MeshContext{
						Resource:   snapshot.Mesh,
						Dataplanes: snapshot.Resources[mesh_core.DataplaneType].(*mesh_core.DataplaneResourceList),
					},
				}
				proxy, err := ToDataplaneProxy(dataplane, metadataTracker.Metadata(streamId), snapshot, rt.Config().Multicluster.Remote.Zone)
				if err != nil {
					return err
				}
				return reconciler.Reconcile(envoyCtx, proxy)
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
