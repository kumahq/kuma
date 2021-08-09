package gc_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/gc"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Subscription Finalizer", func() {
	var rm core_manager.ResourceManager
	var finalizer component.Component

	BeforeEach(func() {
		rm = core_manager.NewResourceManager(memory.NewStore())
	})

	startSubscriptionFinalizer := func(ticks chan time.Time, stop chan struct{}) {
		var err error
		finalizer, err = gc.NewSubscriptionFinalizer(rm, func() *time.Ticker {
			return &time.Ticker{C: ticks}
		}, system.ZoneInsightType)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			_ = finalizer.Start(stop)
		}()
	}

	createZoneInsight := func() {
		sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")
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

	It("should finalize subscription after idle timeout", func() {
		stop := make(chan struct{})
		defer close(stop)
		ticks := make(chan time.Time)
		defer close(ticks)
		startSubscriptionFinalizer(ticks, stop)

		createZoneInsight()

		// finalizer should memorize the current generation = 0
		ticks <- time.Time{}
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		incGeneration()
		// finalizer should memorize the new generation = 1
		ticks <- time.Time{}
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		// finalizer should observe the generation didn't change and set DisconnectTime
		ticks <- time.Time{}
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())
	})
})
