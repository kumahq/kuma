package zoneinsight_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/managers/apis/zoneinsight"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("ZoneInsight Manager", func() {

	It("should limit the number of subscription", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.ZoneMetrics{
			SubscriptionLimit: 3,
		}
		manager := zoneinsight.NewZoneInsightManager(s, cfg)

		err := s.Create(context.Background(), system.NewZoneResource(), store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := system.NewZoneInsightResource()
		for i := 0; i < 10; i++ {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &v1alpha1.KDSSubscription{
				Id: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), input, store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		actual := system.NewZoneInsightResource()
		err = s.Get(context.Background(), actual, store.GetByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(HaveLen(3))
		Expect(actual.Spec.Subscriptions[0].Id).To(Equal("7"))
		Expect(actual.Spec.Subscriptions[1].Id).To(Equal("8"))
		Expect(actual.Spec.Subscriptions[2].Id).To(Equal("9"))
	})

	It("should cleanup subscriptions if limit is 0", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.ZoneMetrics{
			SubscriptionLimit: 0,
		}
		manager := zoneinsight.NewZoneInsightManager(s, cfg)

		err := s.Create(context.Background(), system.NewZoneResource(), store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := system.NewZoneInsightResource()
		for i := 0; i < 10; i++ {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &v1alpha1.KDSSubscription{
				Id: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), input, store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		actual := system.NewZoneInsightResource()
		err = s.Get(context.Background(), actual, store.GetByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(HaveLen(0))
		Expect(actual.Spec.Subscriptions).To(BeNil())
	})
})
