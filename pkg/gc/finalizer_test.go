package gc_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/gc"
	"github.com/kumahq/kuma/pkg/intercp/catalog"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Subscription Finalizer", func() {
	sampleTime, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")

	var rm core_manager.ResourceManager
	var cpCatalog catalog.Catalog

	BeforeEach(func() {
		rm = core_manager.NewResourceManager(memory.NewStore())
		cpCatalog = catalog.NewConfigCatalog(rm)
	})

	startSubscriptionFinalizer := func(ticks chan time.Time, stop chan struct{}) {
		metrics, err := core_metrics.NewMetrics("")
		Expect(err).ToNot(HaveOccurred())
		finalizer, err := gc.NewSubscriptionFinalizer(rm, multitenant.SingleTenant, func() *time.Ticker {
			return &time.Ticker{C: ticks}
		}, metrics, system.ZoneInsightType)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			_ = finalizer.Start(stop)
		}()
	}

	createZoneInsight := func(cpID string) {
		Expect(rm.Create(context.Background(), &system.ZoneInsightResource{
			Spec: &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:               "stream-id-1",
						GlobalInstanceId: cpID,
						ConnectTime:      proto.MustTimestampProto(sampleTime),
						DisconnectTime:   proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(),
					},
					{
						Id:               "stream-id-2",
						GlobalInstanceId: cpID,
						ConnectTime:      proto.MustTimestampProto(sampleTime.Add(1 * time.Hour)),
						Status:           system_proto.NewSubscriptionStatus(),
						Generation:       0,
					},
				},
			},
		}, store.CreateByKey("zone-1", core_model.NoMesh))).To(Succeed())
	}

	replaceCPInstances := func(cpIDs ...string) {
		instances := make([]catalog.Instance, 0, len(cpIDs))
		for _, id := range cpIDs {
			instances = append(instances, catalog.Instance{Id: id})
		}
		updated, err := cpCatalog.Replace(context.Background(), instances)
		Expect(err).ToNot(HaveOccurred())
		Expect(updated).To(BeTrue())
	}

	isOnline := func() bool {
		zoneInsight := system.NewZoneInsightResource()
		Expect(
			rm.Get(context.Background(), zoneInsight, store.GetByKey("zone-1", core_model.NoMesh)),
		).To(Succeed())
		return zoneInsight.Spec.IsOnline()
	}

	It("should finalize subscription when CP instance is absent", func() {
		stop := make(chan struct{})
		defer close(stop)
		ticks := make(chan time.Time)
		defer close(ticks)
		startSubscriptionFinalizer(ticks, stop)

		createZoneInsight("cp-1")

		replaceCPInstances("cp-1")

		// finalizer should not finalize subscription because CP is in catalog
		ticks <- time.Time{}
		Eventually(isOnline, "5s", "100ms").Should(BeTrue())

		replaceCPInstances("cp-2")

		// finalizer should finalize subscription because "cp-1" is absent from catalog
		ticks <- time.Time{}
		Eventually(isOnline, "5s", "100ms").Should(BeFalse())
	})
})
