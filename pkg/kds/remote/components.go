package remote

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"

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
	ProvidedTypes = []model.ResourceType{
		mesh.DataplaneType,
		mesh.DataplaneInsightType,
	}
	ConsumedTypes = []model.ResourceType{
		mesh.MeshType,
		mesh.DataplaneType,
		mesh.ExternalServiceType,
		mesh.CircuitBreakerType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ProxyTemplateType,
		mesh.RetryType,
		system.SecretType,
		system.ConfigType,
	}
)

func Setup(rt core_runtime.Runtime) error {
	zone := rt.Config().Multizone.Remote.Zone
	kdsServer, err := kds_server.New(kdsRemoteLog, rt, ProvidedTypes,
		zone, rt.Config().Multizone.Remote.KDS.RefreshInterval,
		providedFilter(zone), false)
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsRemoteLog, rt.ResourceStore())
	kubeFactory := resources_k8s.NewSimpleKubeFactory()
	onSessionStarted := mux.OnSessionStartedFunc(func(session mux.Session) error {
		log := kdsRemoteLog.WithValues("peer-id", session.PeerID())
		log.Info("new session created")
		go func() {
			if err := kdsServer.StreamKumaResources(session.ServerStream()); err != nil {
				log.Error(err, "StreamKumaResources finished with an error")
			}
		}()
		sink := kds_client.NewKDSSink(log, ConsumedTypes, kds_client.NewKDSStream(session.ClientStream(), zone),
			Callbacks(rt, resourceSyncer, rt.Config().Store.Type == store.KubernetesStore, zone, kubeFactory),
		)
		go func() {
			if err := sink.Start(session.Done()); err != nil {
				log.Error(err, "KDSSink finished with an error")
			}
		}()
		return nil
	})
	muxClient := mux.NewClient(rt.Config().Multizone.Remote.GlobalAddress, zone, onSessionStarted, *rt.Config().Multizone.Remote.KDS, rt.Metrics())
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

func Callbacks(rt core_runtime.Runtime, syncer sync_store.ResourceSyncer, k8sStore bool, localZone string, kubeFactory resources_k8s.KubeFactory) *kds_client.Callbacks {
	return &kds_client.Callbacks{
		OnResourcesReceived: func(clusterID string, rs model.ResourceList) error {
			if k8sStore && rs.GetItemType() != system.ConfigType && rs.GetItemType() != system.SecretType {
				// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
				// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
				// System resources are not in the kubeFactory therefore we need explicit ifs for them
				kubeObject, err := kubeFactory.NewObject(rs.NewItem())
				if err != nil {
					return errors.Wrap(err, "could not convert object")
				}
				if kubeObject.Scope() == k8s_model.ScopeNamespace {
					util.AddSuffixToNames(rs.GetItems(), "default")
				}
			}
			if rs.GetItemType() == mesh.DataplaneType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.(*mesh.DataplaneResource).Spec.IsRemoteIngress(localZone)
				}))
			}
			if rs.GetItemType() == system.ConfigType {
				for _, resource := range rs.GetItems() {
					if resource.GetMeta().GetName() == config_manager.ClusterIdConfigKey {
						if trr, ok := resource.(*system.ConfigResource); ok {
							clusterId := trr.Spec.Config
							rt.SetClusterId(clusterId)
							return nil
						} else {
							return model.ErrorInvalidItemType((*system.ConfigResource)(nil), resource)
						}
					}
				}
			}
			return syncer.Sync(rs)
		},
	}
}

func ConsumesType(typ model.ResourceType) bool {
	for _, consumedTyp := range ConsumedTypes {
		if consumedTyp == typ {
			return true
		}
	}
	return false
}
