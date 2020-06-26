package global

import (
	"fmt"

	"github.com/Kong/kuma/pkg/config/core/resources/store"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds/client"
	sync_store "github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/kds/util"
)

var (
	kdsDataplaneSinkLog = core.Log.WithName("KDS Dataplane Sink")
	resourceTypes       = []model.ResourceType{mesh.DataplaneType}
)

func SetupComponent(rt runtime.Runtime) error {
	syncStore := sync_store.NewResourceSyncer(kdsDataplaneSinkLog, rt.ResourceStore())

	clientFactory := func(clusterIP string) client.ClientFactory {
		return func() (kdsClient client.KDSClient, err error) {
			return client.New(clusterIP)
		}
	}

	for _, cluster := range rt.Config().KumaClusters.Clusters {
		log := kdsDataplaneSinkLog.WithValues("clusterIP", cluster.Remote.Address)
		dataplaneSink := client.NewKDSSink(log, rt.Config().General.ClusterName, resourceTypes,
			clientFactory(cluster.Remote.Address), Callbacks(syncStore, rt.Config().Store.Type == store.KubernetesStore))
		if err := rt.Add(component.NewResilientComponent(log, dataplaneSink)); err != nil {
			return err
		}
	}
	return nil
}

func Callbacks(s sync_store.ResourceSyncer, k8sStore bool) *client.Callbacks {
	clusterTag := func(r model.Resource) string {
		return r.GetSpec().(*mesh_proto.Dataplane).GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ClusterTag]
	}

	return &client.Callbacks{
		OnResourcesReceived: func(rs model.ResourceList) error {
			if len(rs.GetItems()) == 0 {
				return nil
			}
			cluster := clusterTag(rs.GetItems()[0])
			addPrefixToNames(rs.GetItems(), cluster)
			// if type of Store is Kubernetes then we want to store upstream resources in dedicated Namespace.
			// KubernetesStore parses Name and considers substring after the last dot as a Namespace's Name.
			if k8sStore {
				addSuffixToNames(rs.GetItems(), "default")
			}
			return s.Sync(rs, sync_store.PrefilterBy(func(r model.Resource) bool {
				return cluster == clusterTag(r)
			}))
		},
	}
}

func addPrefixToNames(rs []model.Resource, prefix string) {
	for _, r := range rs {
		newName := fmt.Sprintf("%s.%s", prefix, r.GetMeta().GetName())
		// method Sync takes into account only 'Name' and 'Mesh'. Another ResourceMeta's fields like
		// 'Version', 'CreationTime' and 'ModificationTime' will be taken from downstream store.
		// That's why we can set 'Name' like this
		m := util.ResourceKeyToMeta(newName, r.GetMeta().GetMesh())
		r.SetMeta(m)
	}
}

func addSuffixToNames(rs []model.Resource, suffix string) {
	for _, r := range rs {
		newName := fmt.Sprintf("%s.%s", r.GetMeta().GetName(), suffix)
		// method Sync takes into account only 'Name' and 'Mesh'. Another ResourceMeta's fields like
		// 'Version', 'CreationTime' and 'ModificationTime' will be taken from downstream store.
		// That's why we can set 'Name' like this
		m := util.ResourceKeyToMeta(newName, r.GetMeta().GetMesh())
		r.SetMeta(m)
	}
}
