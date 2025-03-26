package gc_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/gc"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Subscription Finalizer", func() {
	sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")

	var rm *countingManager

	BeforeEach(func() {
		rm = &countingManager{ResourceManager: core_manager.NewResourceManager(memory.NewStore())}
	})

	startSubscriptionFinalizer := func(ticks chan time.Time, stop chan struct{}, stopped chan<- error) {
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		finalizer, err := gc.NewSubscriptionFinalizer(rm, multitenant.SingleTenant, func() *time.Ticker {
			return &time.Ticker{C: ticks}
		}, metrics, context.Background(), store_config.DefaultUpsertConfig(), system.ZoneInsightType)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			stopped <- finalizer.Start(stop)
		}()
	}

	alreadyOfflineSub := "stream-id-1"
	onlineSub := "stream-id-2"
	createZoneInsight := func() {
		Expect(rm.Create(context.Background(), &system.ZoneInsightResource{
			Spec: &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:               alreadyOfflineSub,
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime),
						DisconnectTime:   proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(sampleTime),
					},
					{
						Id:               onlineSub,
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(sampleTime.Add(time.Hour)),
						Generation:       0,
					},
				},
			},
		}, store.CreateByKey("zone-1", core_model.NoMesh))).To(Succeed())
	}

	onlineSub1 := "stream-id-2"
	onlineSub2 := "stream-id-3"
	createZoneInsightWithMultipleOnlineSubs := func() {
		Expect(rm.Create(context.Background(), &system.ZoneInsightResource{
			Spec: &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:               "stream-id-1",
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime),
						DisconnectTime:   proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(sampleTime),
					},
					{
						Id:               onlineSub1,
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(sampleTime.Add(time.Hour)),
						Generation:       0,
					},
					{
						Id:               onlineSub2,
						GlobalInstanceId: "cp-1",
						ConnectTime:      proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(sampleTime.Add(time.Hour)),
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

	incGeneration := func(id string) {
		zoneInsight := system.NewZoneInsightResource()
		key := core_model.ResourceKey{Name: "zone-1"}
		Expect(core_manager.Upsert(context.Background(), rm, key, zoneInsight, func(r core_model.Resource) error {
			zoneInsight.Spec.GetSubscription(id).(*system_proto.KDSSubscription).Generation++
			return nil
		}, core_manager.WithConflictRetry(5*time.Millisecond, 5, 10))).To(Succeed())
	}

	disconnectAndAddNewSubscription := func() {
		zoneInsight := system.NewZoneInsightResource()
		key := core_model.ResourceKey{Name: "zone-1"}
		Expect(core_manager.Upsert(context.Background(), rm, key, zoneInsight, func(r core_model.Resource) error {
			zoneInsight.Spec.GetSubscription(onlineSub).SetDisconnectTime(sampleTime.Add(2 * time.Hour))
			zoneInsight.Spec.Subscriptions = append(zoneInsight.Spec.Subscriptions, &system_proto.KDSSubscription{
				Id:               "stream-id-3",
				GlobalInstanceId: "cp-1",
				ConnectTime:      proto.MustTimestampProto(sampleTime.Add(2 * time.Hour)),
				Status:           system_proto.NewSubscriptionStatus(sampleTime.Add(2 * time.Hour)),
				Generation:       0,
			})
			return nil
		}, core_manager.WithConflictRetry(5*time.Millisecond, 5, 10))).To(Succeed())
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
		ticks := make(chan time.Time)
		defer close(ticks)
		stopped := make(chan error)
		startSubscriptionFinalizer(ticks, stop, stopped)

		createZoneInsight()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		incGeneration(onlineSub)
		// finalizer should memorize the new generation = 1
		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		// finalizer should observe the generation didn't change and set DisconnectTime
		ticks <- time.Time{}
		Eventually(listCalled(3), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())

		close(stop)
		Eventually(stopped).Should(Receive(Succeed()))
	})

	It("should not finalize subscription if generation is the same, but subscription id was changed", func() {
		stop := make(chan struct{})
		ticks := make(chan time.Time)
		defer close(ticks)
		stopped := make(chan error)
		startSubscriptionFinalizer(ticks, stop, stopped)

		createZoneInsight()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		disconnectAndAddNewSubscription()

		// generation is the same, but last subscriptionId was changed
		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		close(stop)
		Eventually(stopped).Should(Receive(Succeed()))
	})

	It("should finalize multiple subscriptions if generations haven't changed after timeout", func() {
		stop := make(chan struct{})
		ticks := make(chan time.Time)
		defer close(ticks)
		stopped := make(chan error)
		startSubscriptionFinalizer(ticks, stop, stopped)

		createZoneInsightWithMultipleOnlineSubs()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		// finalizer should observe the generation didn't change and set DisconnectTime
		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())

		close(stop)
		Eventually(stopped).Should(Receive(Succeed()))
	})

	It("should finalize only one of multiple online subscriptions if only its generation changes", func() {
		stop := make(chan struct{})
		ticks := make(chan time.Time)
		defer close(ticks)
		stopped := make(chan error)
		startSubscriptionFinalizer(ticks, stop, stopped)

		createZoneInsightWithMultipleOnlineSubs()

		ticks <- time.Time{}
		Eventually(listCalled(1), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		incGeneration(onlineSub2)

		ticks <- time.Time{}
		Eventually(listCalled(2), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		ticks <- time.Time{}
		Eventually(listCalled(3), "5s", "100ms").Should(BeTrue())
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())

		close(stop)
		Eventually(stopped).Should(Receive(Succeed()))
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
