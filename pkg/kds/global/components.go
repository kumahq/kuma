package global

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Kong/kuma/pkg/plugins/leader/memory"

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

func clusterTag(r model.Resource) string {
	return r.GetSpec().(*mesh_proto.Dataplane).GetNetworking().GetInbound()[0].GetTags()[mesh_proto.ClusterTag]
}

func Callbacks(s store.SyncResourceStore) *client.Callbacks {
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

func SetupServer(rt runtime.Runtime) error {
	syncStore := store.NewSyncResourceStore(rt.ResourceStore())
	resilient := component.NewResilientComponent(kdsDataplaneSinkLog, func(log logr.Logger) manager.Runnable {
		mgr := component.NewManager(memory.NewAlwaysLeaderElector())
		for _, cluster := range rt.Config().KumaClusters.Clusters {
			_ = mgr.Add(client.NewKDSSink(log, resourceTypes, func() (kdsClient client.KDSClient, err error) {
				return client.New(cluster.Local.Address)
			}, Callbacks(syncStore)))
		}
		return mgr
	})
	return rt.Add(resilient)
}
