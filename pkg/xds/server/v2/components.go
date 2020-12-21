package v2

import (
	"context"

	envoy_service_discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"google.golang.org/grpc"

	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	"github.com/kumahq/kuma/pkg/xds/auth"
	auth_components "github.com/kumahq/kuma/pkg/xds/auth/components"
	xds_server "github.com/kumahq/kuma/pkg/xds/server"
	xds_callbacks "github.com/kumahq/kuma/pkg/xds/server/callbacks"
	xds_sync "github.com/kumahq/kuma/pkg/xds/sync"
	xds_template "github.com/kumahq/kuma/pkg/xds/template"
)

var xdsServerLog = core.Log.WithName("xds-server")

func RegisterXDS(rt core_runtime.Runtime, server *grpc.Server) error {
	callbacks, err := DefaultCallbacks(rt)
	if err != nil {
		return err
	}
	srv := envoy_server.NewServer(context.Background(), rt.XDS().Cache(), callbacks)

	xdsServerLog.Info("registering Aggregated Discovery Service in Dataplane Server")
	envoy_service_discovery.RegisterAggregatedDiscoveryServiceServer(server, srv)
	return nil
}

func DefaultCallbacks(rt core_runtime.Runtime) (envoy_server.Callbacks, error) {
	authenticator, err := auth_components.DefaultAuthenticator(rt)
	if err != nil {
		return nil, err
	}
	authCallbacks := auth.NewCallbacks(rt.ResourceManager(), authenticator)

	metadataTracker := xds_callbacks.NewDataplaneMetadataTracker()
	connectionInfoTracker := xds_callbacks.NewConnectionInfoTracker()

	reconciler := DefaultReconciler(rt)
	ingressReconciler := DefaultIngressReconciler(rt)
	watchdogFactory, err := xds_sync.DefaultDataplaneWatchdogFactory(rt, metadataTracker, connectionInfoTracker, reconciler, ingressReconciler)
	if err != nil {
		return nil, err
	}

	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "xds")
	if err != nil {
		return nil, err
	}
	return util_xds_v2.CallbacksChain{
		util_xds_v2.AdaptCallbacks(statsCallbacks),
		util_xds_v2.AdaptCallbacks(connectionInfoTracker),
		util_xds_v2.AdaptCallbacks(authCallbacks),
		util_xds_v2.AdaptCallbacks(xds_callbacks.NewDataplaneSyncTracker(watchdogFactory.New)),
		util_xds_v2.AdaptCallbacks(metadataTracker),
		util_xds_v2.AdaptCallbacks(xds_callbacks.NewDataplaneLifecycle(rt.ResourceManager())),
		util_xds_v2.AdaptCallbacks(xds_server.DefaultDataplaneStatusTracker(rt)),
		newResourceWarmingForcer(rt.XDS().Cache(), rt.XDS().Hasher()),
	}, nil
}

func DefaultReconciler(rt core_runtime.Runtime) xds_sync.SnapshotReconciler {
	return &reconciler{
		&templateSnapshotGenerator{
			ProxyTemplateResolver: &xds_template.SimpleProxyTemplateResolver{
				ReadOnlyResourceManager: rt.ReadOnlyResourceManager(),
				DefaultProxyTemplate:    xds_template.DefaultProxyTemplate,
			},
		},
		&simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}

func DefaultIngressReconciler(rt core_runtime.Runtime) xds_sync.SnapshotReconciler {
	return &reconciler{
		generator: &templateSnapshotGenerator{
			ProxyTemplateResolver: &xds_template.StaticProxyTemplateResolver{
				Template: xds_template.IngressProxyTemplate,
			},
		},
		cacher: &simpleSnapshotCacher{rt.XDS().Hasher(), rt.XDS().Cache()},
	}
}
