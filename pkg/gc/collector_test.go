package gc_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/gc"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/util/proto"
)

// The problem is SubscriptionFinalizer in the same package uses core.Now
// variable to set DisconnectTime when finalizes subscription.
// Test in the current file replaces core.Now for testing purposes. When running
// the whole suite with `-race` flag it constantly fails with `WARNING: DATA RACE`.
// In order to avoid that we protect core.Now with mtxNow mutex in the scope
// of this gc_test package.
var mtxNow sync.RWMutex

var _ = Describe("Dataplane Collector", func() {
	var rm manager.ResourceManager
	now := time.Now()
	var backupNow func() time.Time

	createDpAndDpInsight := func(name, mesh string) {
		dp := &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{Name: name, Mesh: mesh},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 8080,
						Tags: map[string]string{
							"kuma.io/service": "db",
						},
					}},
				},
			},
		}
		dpInsight := &core_mesh.DataplaneInsightResource{
			Meta: &model.ResourceMeta{Name: name, Mesh: mesh},
			Spec: &mesh_proto.DataplaneInsight{
				Subscriptions: []*mesh_proto.DiscoverySubscription{
					{
						DisconnectTime: proto.MustTimestampProto(core.Now()),
					},
				},
			},
		}
		err := rm.Create(context.Background(), dp, store.CreateByKey(name, mesh))
		Expect(err).ToNot(HaveOccurred())
		err = rm.Create(context.Background(), dpInsight, store.CreateByKey(name, mesh))
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		mtxNow.Lock()
		backupNow = core.Now
		core.Now = func() time.Time {
			return now
		}
		mtxNow.Unlock()
	})

	AfterEach(func() {
		mtxNow.Lock()
		core.Now = backupNow
		mtxNow.Unlock()
	})

	It("should cleanup old dataplanes", func() {
		// given 5 dataplanes now
		for i := 0; i < 5; i++ {
			createDpAndDpInsight(fmt.Sprintf("dp-%d", i), "default")
		}
		now = now.Add(1 * time.Hour)
		// given 5 dataplanes after an hour
		for i := 5; i < 10; i++ {
			createDpAndDpInsight(fmt.Sprintf("dp-%d", i), "default")
		}

		now = now.Add(30 * time.Minute)
		// when dataplane collector is run after 1.5 hours
		collector := gc.NewCollector(rm, 100*time.Millisecond, 1*time.Hour)

		stop := make(chan struct{})
		defer close(stop)
		go func() {
			_ = collector.Start(stop)
		}()

		// then first 5 dataplanes that are offline for more than 1 hour are deleted
		Eventually(func() (int, error) {
			dataplanes := &core_mesh.DataplaneResourceList{}
			if err := rm.List(context.Background(), dataplanes); err != nil {
				return 0, err
			}
			return len(dataplanes.Items), nil
		}).Should(Equal(5))

		actual := &core_mesh.DataplaneResourceList{}
		err := rm.List(context.Background(), actual)
		Expect(err).ToNot(HaveOccurred())
		names := []string{}
		for _, dp := range actual.Items {
			names = append(names, dp.Meta.GetName())
		}
		Expect(names).To(Equal([]string{"dp-5", "dp-6", "dp-7", "dp-8", "dp-9"}))
	})
})
