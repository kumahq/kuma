package global

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"

	"github.com/pkg/errors"

	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"

	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds/client"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
)

var (
	kdsGlobalLog = core.Log.WithName("kds-global")

	// ProvidedTypes lists the resource types provided by the Global
	// CP to the Zone CP.
	ProvidedTypes = []model.ResourceType{
		core_mesh.CircuitBreakerType,
		core_mesh.DataplaneType,
		core_mesh.ExternalServiceType,
		core_mesh.FaultInjectionType,
		core_mesh.HealthCheckType,
		core_mesh.MeshType,
		core_mesh.ProxyTemplateType,
		core_mesh.RateLimitType,
		core_mesh.RetryType,
		core_mesh.TimeoutType,
		core_mesh.TrafficLogType,
		core_mesh.TrafficPermissionType,
		core_mesh.TrafficRouteType,
		core_mesh.TrafficTraceType,
		core_mesh.ZoneIngressType,
		system.ConfigType,
		system.GlobalSecretType,
		system.SecretType,
	}

	// ConsumedTypes lists the resource types consumed from the Zone
	// CP by the Global CP.
	ConsumedTypes = []model.ResourceType{
		core_mesh.DataplaneInsightType,
		core_mesh.DataplaneType,
		core_mesh.ZoneIngressInsightType,
		core_mesh.ZoneIngressType,
	}
)

func Setup(rt runtime.Runtime) (err error) {
	kdsServer, err := kds_server.New(kdsGlobalLog, rt, ProvidedTypes,
		"global", rt.Config().Multizone.Global.KDS.RefreshInterval,
		rt.KDSContext().GlobalProvidedFilter, true)
	if err != nil {
		return err
	}
	resourceSyncer := sync_store.NewResourceSyncer(kdsGlobalLog, rt.ResourceStore())
	kubeFactory := resources_k8s.NewSimpleKubeFactory()
	onSessionStarted := mux.OnSessionStartedFunc(func(session mux.Session) error {
		log := kdsGlobalLog.WithValues("peer-id", session.PeerID())
		log.Info("new session created")
		go func() {
			if err := kdsServer.StreamKumaResources(session.ServerStream()); err != nil {
				log.Error(err, "StreamKumaResources finished with an error")
			}
		}()
		kdsStream := client.NewKDSStream(session.ClientStream(), session.PeerID())
		if err := createZoneIfAbsent(session.PeerID(), rt.ResourceManager()); err != nil {
			log.Error(err, "Global CP could not create a zone")
			return errors.New("Global CP could not create a zone") // send back message without details. Zone CP will retry
		}
		sink := client.NewKDSSink(log, ConsumedTypes, kdsStream, Callbacks(resourceSyncer, rt.Config().Store.Type == store_config.KubernetesStore, kubeFactory))
		go func() {
			if err := sink.Start(session.Done()); err != nil {
				log.Error(err, "KDSSink finished with an error")
			}
		}()
		return nil
	})
	callbacks := append(rt.KDSContext().GlobalServerCallbacks, onSessionStarted)
	return rt.Add(mux.NewServer(callbacks, *rt.Config().Multizone.Global.KDS, rt.Metrics()))
}

func createZoneIfAbsent(name string, resManager manager.ResourceManager) error {
	if err := resManager.Get(context.Background(), system.NewZoneResource(), store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		kdsGlobalLog.Info("creating Zone", "name", name)
		zone := &system.ZoneResource{
			Spec: &system_proto.Zone{
				Enabled: &wrapperspb.BoolValue{Value: true},
			},
		}
		if err := resManager.Create(context.Background(), zone, store.CreateByKey(name, model.NoMesh)); err != nil {
			return err
		}
	}
	return nil
}

func Callbacks(s sync_store.ResourceSyncer, k8sStore bool, kubeFactory resources_k8s.KubeFactory) *client.Callbacks {
	return &client.Callbacks{
		OnResourcesReceived: func(clusterName string, rs model.ResourceList) error {
			util.AddPrefixToNames(rs.GetItems(), clusterName)
			if k8sStore {
				// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
				// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
				kubeObject, err := kubeFactory.NewObject(rs.NewItem())
				if err != nil {
					return errors.Wrap(err, "could not convert object")
				}
				if kubeObject.Scope() == k8s_model.ScopeNamespace {
					util.AddSuffixToNames(rs.GetItems(), "default")
				}
			}
			if rs.GetItemType() == core_mesh.ZoneIngressType {
				for _, zi := range rs.(*core_mesh.ZoneIngressResourceList).Items {
					zi.Spec.Zone = clusterName
				}
			}
			return s.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
				return strings.HasPrefix(r.GetMeta().GetName(), fmt.Sprintf("%s.", clusterName))
			}), sync_store.Zone(clusterName))
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
