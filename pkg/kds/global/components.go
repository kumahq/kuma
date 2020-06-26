package global

import (
	"net/url"

	"github.com/Kong/kuma/pkg/config/clusters"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds/client"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	sync_store "github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/kds/util"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	kdsGlobalLog  = core.Log.WithName("kds-global")
	providedTypes = []model.ResourceType{
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
	consumedTypes = []model.ResourceType{mesh.DataplaneType}
)

func SetupComponent(rt runtime.Runtime) error {
	syncStore := sync_store.NewResourceSyncer(kdsGlobalLog, rt.ResourceStore())

	clientFactory := func(clusterIP string) client.ClientFactory {
		return func() (kdsClient client.KDSClient, err error) {
			return client.New(clusterIP)
		}
	}

	for _, cluster := range rt.Config().KumaClusters.Clusters {
		log := kdsGlobalLog.WithValues("clusterIP", cluster.Remote.Address)
		dataplaneSink := client.NewKDSSink(log, rt.Config().KumaClusters.LBConfig.Address, consumedTypes,
			clientFactory(cluster.Remote.Address), Callbacks(syncStore, rt.Config().Store.Type == store.KubernetesStore, cluster))
		if err := rt.Add(component.NewResilientComponent(log, dataplaneSink)); err != nil {
			return err
		}
	}
	return nil
}

func SetupServer(rt runtime.Runtime) error {
	hasher, cache := kds_server.NewXdsContext(kdsGlobalLog)
	generator := kds_server.NewSnapshotGenerator(rt, providedTypes, filter)
	versioner := kds_server.NewVersioner()
	reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := kds_server.NewSyncTracker(kdsGlobalLog, reconciler, rt.Config().KDSServer.RefreshInterval)
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsGlobalLog},
		syncTracker,
	}
	srv := kds_server.NewServer(cache, callbacks, kdsGlobalLog)
	return rt.Add(kds_server.NewKDSServer(srv, *rt.Config().KDSServer))
}

// filter excludes Dataplanes and Ingresses from 'clusterID' cluster
func filter(clusterID string, r model.Resource) bool {
	if r.GetType() != mesh.DataplaneType {
		return true
	}
	if !r.(*mesh.DataplaneResource).Spec.IsIngress() {
		return false
	}
	return clusterID != util.ClusterTag(r)
}

func Callbacks(s sync_store.ResourceSyncer, k8sStore bool, cfg *clusters.ClusterConfig) *client.Callbacks {
	return &client.Callbacks{
		OnResourcesReceived: func(rs model.ResourceList) error {
			if len(rs.GetItems()) == 0 {
				return nil
			}
			cluster := util.ClusterTag(rs.GetItems()[0])
			util.AddPrefixToNames(rs.GetItems(), cluster)
			// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
			// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
			if k8sStore {
				util.AddSuffixToNames(rs.GetItems(), "default")
			}
			adjustIngressNetworking(cfg, rs)
			return s.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
				return cluster == util.ClusterTag(r)
			}))
		},
	}
}

func adjustIngressNetworking(cfg *clusters.ClusterConfig, rs model.ResourceList) {
	if rs.GetItemType() != mesh.DataplaneType {
		return
	}
	u, _ := url.Parse(cfg.Ingress.Address)
	for _, r := range rs.GetItems() {
		if !r.(*mesh.DataplaneResource).Spec.IsIngress() {
			continue
		}
		r.(*mesh.DataplaneResource).Spec.Networking.Address = u.Hostname()
	}
}
