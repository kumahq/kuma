package global

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	"github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/kds/store"
)

var (
	kdsDataplaneSinkLog = core.Log.WithName("KDS Dataplane Sink")
	resourceTypes       = []model.ResourceType{mesh.DataplaneType}
)

func SetupComponent(rt runtime.Runtime) error {
	syncStore := store.NewSyncResourceStore(kdsDataplaneSinkLog, rt.ResourceStore())
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
	return &client.Callbacks{
		OnResourcesReceived: func(rs model.ResourceList) error {
			if len(rs.GetItems()) == 0 {
				return nil
			}
			cluster := clusterTag(rs.GetItems()[0])
			return s.Sync(rs, store.PrefilterBy(func(r model.Resource) bool {
				return cluster == clusterTag(r)
			}))
		},
	}
}

func clusterTag(r model.Resource) string {
	return r.GetSpec().(*mesh_proto.Dataplane).GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ClusterTag]
}

func clientFactory(clusterIP string) client.ClientFactory {
	return func() (kdsClient client.KDSClient, err error) {
		return client.New(clusterIP)
	}
}
