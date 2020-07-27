package remote

import (
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
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

func Setup(rt core_runtime.Runtime) error {
	zone := rt.Config().Mode.Remote.Zone
	kdsServer, err := kds_server.New(kdsRemoteLog, rt, providedTypes, zone, providedFilter(zone))
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsRemoteLog, rt.ResourceStore())
	onSessionStarted := mux.OnSessionStartedFunc(func(session mux.Session) error {
		log := kdsRemoteLog.WithValues("peer-id", session.PeerID())
		log.Info("new session created")
		go func() {
			if err := kdsServer.StreamKumaResources(session.ServerStream()); err != nil {
				log.Error(err, "StreamKumaResources finished with an error")
			}
		}()
		sink := kds_client.NewKDSSink(log, consumedTypes, kds_client.NewKDSStream(session.ClientStream(), zone),
			Callbacks(resourceSyncer, rt.Config().Store.Type == store.KubernetesStore, zone),
		)
		go func() {
			if err := sink.Start(session.Done()); err != nil {
				log.Error(err, "KDSSink finished with an error")
			}
		}()
		return nil
	})
	muxClient := mux.NewClient(rt.Config().KDS.Client.GlobalAddress, zone, onSessionStarted, *rt.Config().KDS.Client)
	return rt.Add(component.NewResilientComponent(kdsRemoteLog.WithName("mux-client"), muxClient))
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

func ConsumesType(typ model.ResourceType) bool {
	for _, consumedTyp := range consumedTypes {
		if consumedTyp == typ {
			return true
		}
	}
	return false
}
