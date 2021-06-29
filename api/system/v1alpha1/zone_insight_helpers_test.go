package v1alpha1_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/api/internal/util/proto"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
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
			zoneInsight.UpdateSubscription(&system_proto.KDSSubscription{
				Id:          "3",
				ConnectTime: util_proto.MustTimestampProto(t1.Add(3 * time.Hour)),
			})

			// then
			_, subscription := zoneInsight.GetSubscription("2")
			Expect(subscription.DisconnectTime).ToNot(BeNil())
		})
	})
})
