package zone

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"

	"github.com/kumahq/kuma/pkg/config/core/resources/store"
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
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
)

var (
	kdsZoneLog = core.Log.WithName("kds-zone")

	// ProvidedTypes lists the resource types provided by the Zone
	// CP to the Global CP.
	ProvidedTypes = []model.ResourceType{
		mesh.DataplaneInsightType,
		mesh.DataplaneType,
		mesh.ZoneIngressInsightType,
		mesh.ZoneIngressType,
	}

	// ConsumedTypes lists the resource types consumed from the
	// Global CP by the Zone CP.
	ConsumedTypes = []model.ResourceType{
		mesh.CircuitBreakerType,
		mesh.DataplaneType,
		mesh.ExternalServiceType,
		mesh.FaultInjectionType,
		mesh.HealthCheckType,
		mesh.MeshType,
		mesh.ProxyTemplateType,
		mesh.RateLimitType,
		mesh.RetryType,
		mesh.TimeoutType,
		mesh.TrafficLogType,
		mesh.TrafficPermissionType,
		mesh.TrafficRouteType,
		mesh.TrafficTraceType,
		mesh.ZoneIngressType,
		system.ConfigType,
		system.GlobalSecretType,
		system.SecretType,
	}
)

func Setup(rt core_runtime.Runtime) error {
	zone := rt.Config().Multizone.Zone.Name
	kdsServer, err := kds_server.New(kdsZoneLog, rt, ProvidedTypes,
		zone, rt.Config().Multizone.Zone.KDS.RefreshInterval,
		rt.KDSContext().ZoneProvidedFilter, false)
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsZoneLog, rt.ResourceStore())
	kubeFactory := resources_k8s.NewSimpleKubeFactory()
	onSessionStarted := mux.OnSessionStartedFunc(func(session mux.Session) error {
		log := kdsZoneLog.WithValues("peer-id", session.PeerID())
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
	muxClient := mux.NewClient(
		rt.Config().Multizone.Zone.GlobalAddress,
		zone,
		onSessionStarted,
		*rt.Config().Multizone.Zone.KDS,
		rt.Metrics(),
		rt.KDSContext().ZoneClientCtx,
	)
	return rt.Add(component.NewResilientComponent(kdsZoneLog.WithName("mux-client"), muxClient))
}

func Callbacks(rt core_runtime.Runtime, syncer sync_store.ResourceSyncer, k8sStore bool, localZone string, kubeFactory resources_k8s.KubeFactory) *kds_client.Callbacks {
	return &kds_client.Callbacks{
		OnResourcesReceived: func(clusterID string, rs model.ResourceList) error {
			if k8sStore && rs.GetItemType() != system.ConfigType && rs.GetItemType() != system.SecretType && rs.GetItemType() != system.GlobalSecretType {
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
					return r.(*mesh.DataplaneResource).Spec.IsZoneIngress(localZone)
				}))
			}
			if rs.GetItemType() == mesh.ZoneIngressType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.(*mesh.ZoneIngressResource).IsRemoteIngress(localZone)
				}))
			}
			if rs.GetItemType() == system.ConfigType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return rt.KDSContext().Configs[r.GetMeta().GetName()]
				}))
			}
			if rs.GetItemType() == system.GlobalSecretType {
				return syncer.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
					return r.GetMeta().GetName() == zoneingress.SigningKeyResourceKey().Name
				}))
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
