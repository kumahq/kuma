package global

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/kds/mux"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/util"
	resources_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	k8s_model "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var (
	kdsGlobalLog = core.Log.WithName("kds-global")
)

func Setup(rt runtime.Runtime) (err error) {
	reg := registry.Global()
	kdsServer, err := kds_server.New(kdsGlobalLog, rt, reg.ObjectTypes(model.HasKDSFlag(model.ProvidedByGlobal)),
		"global", rt.Config().Multizone.Global.KDS.RefreshInterval,
		rt.KDSContext().GlobalProvidedFilter, rt.KDSContext().GlobalResourceMapper, true)
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
			} else {
				log.V(1).Info("StreamKumaResources finished gracefully")
			}
		}()
		kdsStream := client.NewKDSStream(session.ClientStream(), session.PeerID(), "") // we only care about Zone CP config. Zone CP should not receive Global CP config.
		if err := createZoneIfAbsent(session.PeerID(), rt.ResourceManager()); err != nil {
			log.Error(err, "Global CP could not create a zone")
			return errors.New("Global CP could not create a zone") // send back message without details. Zone CP will retry
		}
		sink := client.NewKDSSink(log, reg.ObjectTypes(model.HasKDSFlag(model.ConsumedByGlobal)), kdsStream, Callbacks(resourceSyncer, rt.Config().Store.Type == store_config.KubernetesStore, kubeFactory))
		go func() {
			if err := sink.Receive(); err != nil {
				log.Error(err, "KDSSink finished with an error")
			} else {
				log.V(1).Info("KDSSink finished gracefully")
			}
		}()
		return nil
	})
	return rt.Add(mux.NewServer(onSessionStarted, rt.KDSContext().GlobalServerFilters, *rt.Config().Multizone.Global.KDS, rt.Metrics()))
}

func createZoneIfAbsent(name string, resManager manager.ResourceManager) error {
	if err := resManager.Get(context.Background(), system.NewZoneResource(), store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		kdsGlobalLog.Info("creating Zone", "name", name)
		zone := &system.ZoneResource{
			Spec: &system_proto.Zone{
				Enabled: util_proto.Bool(true),
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
			} else if rs.GetItemType() == core_mesh.ZoneEgressType {
				for _, ze := range rs.(*core_mesh.ZoneEgressResourceList).Items {
					ze.Spec.Zone = clusterName
				}
			}

			return s.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
				return strings.HasPrefix(r.GetMeta().GetName(), fmt.Sprintf("%s.", clusterName))
			}), sync_store.Zone(clusterName))
		},
	}
}
