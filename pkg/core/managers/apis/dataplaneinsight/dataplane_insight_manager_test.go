package dataplaneinsight_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core/managers/apis/dataplaneinsight"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("DataplaneInsight Manager", func() {

	It("should limit the number of subscription", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.DataplaneMetrics{
			Enabled:           true,
			SubscriptionLimit: 3,
		}
		manager := dataplaneinsight.NewDataplaneInsightManager(s, cfg)

		err := s.Create(context.Background(), &mesh_core.DataplaneResource{}, store.CreateByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		input := mesh_core.DataplaneInsightResource{}
		for i := 0; i < 10; i++ {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
				Id: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), &input, store.CreateByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		actual := mesh_core.DataplaneInsightResource{}
		err = s.Get(context.Background(), &actual, store.GetByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(HaveLen(3))
		Expect(actual.Spec.Subscriptions[0].Id).To(Equal("7"))
		Expect(actual.Spec.Subscriptions[1].Id).To(Equal("8"))
		Expect(actual.Spec.Subscriptions[2].Id).To(Equal("9"))
	})

	It("should cleanup subscriptions if disabled", func() {
		// setup
		s := memory.NewStore()
		cfg := &kuma_cp.DataplaneMetrics{
			Enabled: false,
		}
		manager := dataplaneinsight.NewDataplaneInsightManager(s, cfg)

		err := s.Create(context.Background(), &mesh_core.DataplaneResource{}, store.CreateByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		input := mesh_core.DataplaneInsightResource{}
		for i := 0; i < 10; i++ {
			input.Spec.Subscriptions = append(input.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
				Id: fmt.Sprintf("%d", i),
			})
		}

		// when
		err = manager.Create(context.Background(), &input, store.CreateByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		actual := mesh_core.DataplaneInsightResource{}
		err = s.Get(context.Background(), &actual, store.GetByKey("di1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Subscriptions).To(HaveLen(0))
		Expect(actual.Spec.Subscriptions).To(BeNil())
	})
})
