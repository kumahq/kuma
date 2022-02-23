package v1alpha1_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Zone Insights", func() {
	Context("UpdateSubscription", func() {
		t1, _ := time.Parse(time.RFC3339, "2018-07-17T16:05:36.995+00:00")

		It("should leave subscriptions in a valid state", func() {
			// given
			zoneInsight := &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:             "1",
						ConnectTime:    util_proto.MustTimestampProto(t1),
						DisconnectTime: util_proto.MustTimestampProto(t1.Add(1 * time.Hour)),
					},
					{
						Id:          "2",
						ConnectTime: util_proto.MustTimestampProto(t1.Add(2 * time.Hour)),
					},
				},
			}

			// when
			Expect(zoneInsight.UpdateSubscription(&system_proto.KDSSubscription{
				Id:          "3",
				ConnectTime: util_proto.MustTimestampProto(t1.Add(3 * time.Hour)),
			})).To(Succeed())

			// then
			_, subscription := zoneInsight.GetSubscription("2")
			Expect(subscription.DisconnectTime).ToNot(BeNil())
		})

		It("should return error for wrong subscription type", func() {
			// given
			zoneInsight := &system_proto.ZoneInsight{
				Subscriptions: []*system_proto.KDSSubscription{
					{
						Id:             "1",
						ConnectTime:    util_proto.MustTimestampProto(t1),
						DisconnectTime: util_proto.MustTimestampProto(t1.Add(1 * time.Hour)),
					},
					{
						Id:          "2",
						ConnectTime: util_proto.MustTimestampProto(t1.Add(2 * time.Hour)),
					},
				},
			}

			// when
			err := zoneInsight.UpdateSubscription(&mesh_proto.DiscoverySubscription{})

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid type *v1alpha1.DiscoverySubscription for ZoneInsight"))
		})
	})
})
