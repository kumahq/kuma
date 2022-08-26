package gc_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/gc"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Subscription Finalizer", func() {
	sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")

	var rm *countingManager

	BeforeEach(func() {
		rm = &countingManager{ResourceManager: core_manager.NewResourceManager(memory.NewStore())}
	})

	startSubscriptionFinalizer := func(ticks chan time.Time, stop chan struct{}) {
		finalizer, err := gc.NewSubscriptionFinalizer(rm, func() *time.Ticker {
			return &time.Ticker{C: ticks}
		}, system.ZoneInsightType)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			_ = finalizer.Start(stop)
		}()
	}

	createZoneInsight := func() {
		Expect(rm.Create(context.Background(), &system.ZoneInsightResource{
			Spec: &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:               "stream-id-1",
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime),
						DisconnectTime:   proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(),
					},
					{
						Id:               "stream-id-2",
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(),
						Generation:       0,
					},
				},
			},
		}, store.CreateByKey("zone-1", core_model.NoMesh))).To(Succeed())
	}

	isOnline := func() bool {
		zoneInsight := system.NewZoneInsightResource()
		Expect(
			rm.Get(context.Background(), zoneInsight, store.GetByKey("zone-1", core_model.NoMesh)),
		).To(Succeed())
		return zoneInsight.Spec.IsOnline()
	}

	incGeneration := func() {
		zoneInsight := system.NewZoneInsightResource()
		Expect(
			rm.Get(context.Background(), zoneInsight, store.GetByKey("zone-1", core_model.NoMesh)),
		).To(Succeed())
		zoneInsight.Spec.GetLastSubscription().(*system_proto.KDSSubscription).Generation++
		Expect(rm.Update(context.Background(), zoneInsight)).To(Succeed())
	}

	addNewSubscription := func() {
		zoneInsight := system.NewZoneInsightResource()
		Expect(
			rm.Get(context.Background(), zoneInsight, store.GetByKey("zone-1", core_model.NoMesh)),
		).To(Succeed())
		zoneInsight.Spec.GetLastSubscription().SetDisconnectTime(sampleTime.Add(2 * time.Hour))
		zoneInsight.Spec.Subscriptions = append(zoneInsight.Spec.Subscriptions, &system_proto.KDSSubscription{
			Id:               "stream-id-3",
			GlobalInstanceId: "cp-1",
			ConnectTime:      proto.MustTimestampProto(sampleTime.Add(2 * time.Hour)),
			Status:           system_proto.NewSubscriptionStatus(),
			Generation:       0,
		})
		Expect(rm.Update(context.Background(), zoneInsight)).To(Succeed())
	}

	listCalled := func(times uint32) func() bool {
		return func() bool {
			rm.mtxList.Lock()
			defer rm.mtxList.Unlock()
			return rm.list == times
		}
	}

	It("should finalize subscription after idle timeout", func() {
		stop := make(chan struct{})
		defer close(stop)
		ticks := make(chan time.Time)
		defer close(ticks)
		startSubscriptionFinalizer(ticks, stop)

		createZoneInsight()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		incGeneration()
		// finalizer should memorize the new generation = 1
		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		// finalizer should observe the generation didn't change and set DisconnectTime
		ticks <- time.Time{}
		Eventually(listCalled(3), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())
	})

	It("should not finalize subscription if generation is the same, but subscription id was changed", func() {
		stop := make(chan struct{})
		defer close(stop)
		ticks := make(chan time.Time)
		defer close(ticks)
		startSubscriptionFinalizer(ticks, stop)

		createZoneInsight()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		addNewSubscription()

		// generation is the same, but last subscriptionId was changed
		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())
	})
})

type countingManager struct {
	core_manager.ResourceManager
	mtxList sync.Mutex
	list    uint32
}

func (c *countingManager) List(ctx context.Context, rl core_model.ResourceList, opts ...store.ListOptionsFunc) error {
	c.mtxList.Lock()
	defer c.mtxList.Unlock()

	c.list++
	return c.ResourceManager.List(ctx, rl, opts...)
}
