package zoneinsight_test

import (
	"context"
	"fmt"

	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v3/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/zoneinsight"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("ZoneInsight Manager", func() {
	It("should limit the number of subscription", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.ZoneMetrics{
			SubscriptionLimit: 3,
		}
		manager := zoneinsight.NewZoneInsightManager(s, cfg)

		err := s.Create(context.Background(), zone_api.NewZoneResource(), store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := zoneinsight_api.NewZoneInsightResource()
		for i := range 10 {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &zoneinsight_api.KDSSubscription{
				ID: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), input, store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		actual := zoneinsight_api.NewZoneInsightResource()
		err = s.Get(context.Background(), actual, store.GetByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(HaveLen(3))
		Expect(actual.Spec.Subscriptions[0].ID).To(Equal("7"))
		Expect(actual.Spec.Subscriptions[1].ID).To(Equal("8"))
		Expect(actual.Spec.Subscriptions[2].ID).To(Equal("9"))
	})

	It("should cleanup subscriptions if limit is 0", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.ZoneMetrics{
			SubscriptionLimit: 0,
		}
		manager := zoneinsight.NewZoneInsightManager(s, cfg)

		err := s.Create(context.Background(), zone_api.NewZoneResource(), store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := zoneinsight_api.NewZoneInsightResource()
		for i := range 10 {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &zoneinsight_api.KDSSubscription{
				ID: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), input, store.CreateByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		actual := zoneinsight_api.NewZoneInsightResource()
		err = s.Get(context.Background(), actual, store.GetByKey("di1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(BeEmpty())
		Expect(actual.Spec.Subscriptions).To(BeNil())
	})
})
