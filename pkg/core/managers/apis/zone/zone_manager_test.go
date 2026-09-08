package zone_test

import (
	"context"
	"time"

	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/zone"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("Zone Manager", func() {
	var validator zone.Validator
	var resStore store.ResourceStore

	BeforeEach(func() {
		resStore = memory.NewStore()
		validator = zone.Validator{Store: resStore}
	})

	It("should not delete zone if it's online", func() {
		// given zone and zoneInsight
		err := resStore.Create(context.Background(), zone_api.NewZoneResource(), store.CreateByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = resStore.Create(context.Background(), &zoneinsight_api.ZoneInsightResource{
			Spec: &zoneinsight_api.ZoneInsight{
				Subscriptions: []*zoneinsight_api.KDSSubscription{
					{
						ConnectTime: zoneinsight_api.NewTime(time.Now()),
					},
				},
			},
		}, store.CreateByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		zoneManager := zone.NewZoneManager(resStore, validator, false)

		zone := zone_api.NewZoneResource()
		err = resStore.Get(context.Background(), zone, store.GetByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = zoneManager.Delete(context.Background(), zone, store.DeleteByKey("zone-1", model.NoMesh))

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("zone: unable to delete Zone, Zone CP is still connected, please shut it down first"))
	})

	It("should delete if zone is online when unsafe delete is enabled", func() {
		// given zone and zoneInsight
		err := resStore.Create(context.Background(), zone_api.NewZoneResource(), store.CreateByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = resStore.Create(context.Background(), &zoneinsight_api.ZoneInsightResource{
			Spec: &zoneinsight_api.ZoneInsight{
				Subscriptions: []*zoneinsight_api.KDSSubscription{
					{
						ConnectTime: zoneinsight_api.NewTime(time.Now()),
					},
				},
			},
		}, store.CreateByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		zoneManager := zone.NewZoneManager(resStore, validator, true)

		zone := zone_api.NewZoneResource()
		err = resStore.Get(context.Background(), zone, store.GetByKey("zone-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = zoneManager.Delete(context.Background(), zone, store.DeleteByKey("zone-1", model.NoMesh))

		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
