package remote

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

var (
	kdsRemoteLog  = core.Log.WithName("kds-remote")
	providedTypes = []model.ResourceType{
		mesh.DataplaneType,
		mesh.DataplaneInsightType,
	}
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
		system.SecretType,
	}
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := kds_server.NewXdsContext(kdsRemoteLog)
	generator := kds_server.NewSnapshotGenerator(rt, providedTypes, providedFilter(rt.Config().Mode.Remote.Zone))
	versioner := kds_server.NewVersioner()
	reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := kds_server.NewSyncTracker(kdsRemoteLog, reconciler, rt.Config().KDS.Server.RefreshInterval)
	resourceSyncer := sync_store.NewResourceSyncer(kdsRemoteLog, rt.ResourceStore())

	clientFactory := func(clusterAddress string) kds_client.ClientFactory {
		return func() (kdsClient kds_client.KDSClient, err error) {
			return kds_client.New(clusterAddress, rt.Config().KDS.Client)
		}
	}

	componentFactory := func(log logr.Logger, req *envoy_api_v2.DiscoveryRequest) component.Component {
		globalAddress := req.Node.Id
		policiesSink := kds_client.NewKDSSink(kdsRemoteLog, rt.Config().Mode.Remote.Zone, consumedTypes,
			clientFactory(globalAddress), Callbacks(resourceSyncer, rt.Config().Store.Type == store.KubernetesStore, rt.Config().Mode.Remote.Zone))
		return component.NewResilientComponent(kdsRemoteLog, policiesSink)
	}

	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsRemoteLog},
		syncTracker,
		NewComponentSpawner(kdsRemoteLog.WithName("policy-sink-spawner"), componentFactory),
	}
	srv := kds_server.NewServer(cache, callbacks, kdsRemoteLog, rt.Config().Mode.Remote.Zone)
	return rt.Add(kds_server.NewKDSServer(srv, *rt.Config().KDS.Server))
}

// providedFilter filter Resources provided by Remote, specifically Ingresses that belongs to another zones
func providedFilter(clusterName string) reconcile.ResourceFilter {
	return func(_ string, r model.Resource) bool {
		if r.GetType() == mesh.DataplaneType {
			return clusterName == util.ZoneTag(r)
		}
		return r.GetType() == mesh.DataplaneInsightType
	}
}

func Callbacks(syncer sync_store.ResourceSyncer, k8sStore bool, localZone string) *kds_client.Callbacks {
	return &kds_client.Callbacks{
		OnResourcesReceived: func(clusterID string, rs model.ResourceList) error {
			if k8sStore && rs.GetItemType() != mesh.MeshType && rs.GetItemType() != system.SecretType {
				util.AddSuffixToNames(rs.GetItems(), "default")
			}
			if rs.GetItemType() == mesh.DataplaneType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.(*mesh.DataplaneResource).Spec.IsIngress() && localZone != util.ZoneTag(r)
				}))
			}
			return syncer.Sync(rs)
		},
	}
}
