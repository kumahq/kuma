package global

import (
	"fmt"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/kds/util"
)

var (
	kdsDataplaneSinkLog = core.Log.WithName("KDS Dataplane Sink")
	resourceTypes       = []model.ResourceType{mesh.DataplaneType}
)

func SetupComponent(rt runtime.Runtime) error {
	syncStore := store.NewSyncResourceStore(kdsDataplaneSinkLog, rt.ResourceStore())

	clientFactory := func(clusterIP string) client.ClientFactory {
		return func() (kdsClient client.KDSClient, err error) {
			return client.New(clusterIP)
		}
	}

	for _, cluster := range rt.Config().KumaClusters.Clusters {
		log := kdsDataplaneSinkLog.WithValues("clusterIP", cluster.Local.Address)
		dataplaneSink := client.NewKDSSink(log, rt.Config().General.ClusterName, resourceTypes, clientFactory(cluster.Local.Address), Callbacks(syncStore))
		if err := rt.Add(component.NewResilientComponent(log, dataplaneSink)); err != nil {
			return err
		}
	}
	return nil
}

func Callbacks(s store.ResourceSyncer) *client.Callbacks {
	clusterTag := func(r model.Resource) string {
		return r.GetSpec().(*mesh_proto.Dataplane).GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ClusterTag]
	}

	addPrefixToName := func(prefix string) func(r model.Resource) {
		return func(r model.Resource) {
			newName := fmt.Sprintf("%s.%s", prefix, r.GetMeta().GetName())
			// method Sync takes into account only 'Name' and 'Mesh' that why we can set name like this
			m := util.ResourceKeyToMeta(newName, r.GetMeta().GetMesh())
			r.SetMeta(m)
		}
	}

	return &client.Callbacks{
		OnResourcesReceived: func(rs model.ResourceList) error {
			if len(rs.GetItems()) == 0 {
				return nil
			}
			cluster := clusterTag(rs.GetItems()[0])
			forEach(rs.GetItems()).apply(addPrefixToName(cluster))
			return s.Sync(rs, store.PrefilterBy(func(r model.Resource) bool {
				return cluster == clusterTag(r)
			}))
		},
	}
}

type forEach []model.Resource

func (f forEach) apply(fn func(model.Resource)) {
	for _, item := range f {
		fn(item)
	}
}
