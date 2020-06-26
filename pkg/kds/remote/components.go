package remote

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/go-logr/logr"

	"github.com/Kong/kuma/pkg/config/core/resources/store"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	kds_client "github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/kds/reconcile"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	sync_store "github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/kds/util"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	kdsRemoteLog  = core.Log.WithName("kds-remote")
	providedTypes = []model.ResourceType{mesh.DataplaneType}
	consumedTypes = []model.ResourceType{
		mesh.MeshType,
		mesh.DataplaneType,
		mesh.CircuitBreakerType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ProxyTemplateType,
	}
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := kds_server.NewXdsContext(kdsRemoteLog)
	generator := kds_server.NewSnapshotGenerator(rt, providedTypes, makeFilter(rt.Config().General.ClusterName))
	versioner := kds_server.NewVersioner()
	reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := kds_server.NewSyncTracker(kdsRemoteLog, reconciler, rt.Config().KDSServer.RefreshInterval)
	resourceSyncer := sync_store.NewResourceSyncer(kdsRemoteLog, rt.ResourceStore())

	clientFactory := func(clusterAddress string) kds_client.ClientFactory {
		return func() (kdsClient kds_client.KDSClient, err error) {
			return kds_client.New(clusterAddress)
		}
	}

	componentFactory := func(log logr.Logger, req *envoy_api_v2.DiscoveryRequest) component.Component {
		globalAddress := req.Node.Id
		policiesSink := kds_client.NewKDSSink(kdsRemoteLog, rt.Config().General.ClusterName, consumedTypes,
			clientFactory(globalAddress), Callbacks(resourceSyncer, rt.Config().Store.Type == store.KubernetesStore))
		return component.NewResilientComponent(kdsRemoteLog, policiesSink)
	}

	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsRemoteLog},
		syncTracker,
		NewComponentSpawner(kdsRemoteLog.WithName("policy-sink-spawner"), componentFactory),
	}
	srv := kds_server.NewServer(cache, callbacks, kdsRemoteLog)
	return rt.Add(kds_server.NewKDSServer(srv, *rt.Config().KDSServer))
}

// makeFilter creates filter that exclude Ingresses from another cluster
func makeFilter(clusterID string) reconcile.ResourceFilter {
	return func(_ string, r model.Resource) bool {
		if r.GetType() != mesh.DataplaneType {
			return false
		}
		return clusterID == util.ClusterTag(r)
	}
}

func Callbacks(syncer sync_store.ResourceSyncer, k8sStore bool) *kds_client.Callbacks {
	return &kds_client.Callbacks{
		OnResourcesReceived: func(rs model.ResourceList) error {
			if k8sStore && rs.GetItemType() != mesh.MeshType {
				util.AddSuffixToNames(rs.GetItems(), "default")
			}
			return syncer.Sync(rs)
		},
	}
}
