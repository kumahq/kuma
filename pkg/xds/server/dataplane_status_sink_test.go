package server

import (
	"context"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	test_resources "github.com/Kong/kuma/pkg/test/resources"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/gogo/protobuf/proto"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"

	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("DataplaneInsightSink", func() {

	t0, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")

	Describe("DataplaneInsightSink", func() {

		var recorder *DataplaneInsightStoreRecorder
		var stop chan struct{}

		BeforeEach(func() {
			recorder = &DataplaneInsightStoreRecorder{Upserts: make(chan DataplaneInsightUpsert)}
			stop = make(chan struct{})
		})

		AfterEach(func() {
			close(stop)
		})

		It("should periodically flush DataplaneInsight into a store", func() {
			// setup
			key := core_model.ResourceKey{Mesh: "default", Namespace: "demo", Name: "example-001"}
			subscription := &mesh_proto.DiscoverySubscription{
				Id:                     "3287995C-7E11-41FB-9479-7D39337F845D",
				ControlPlaneInstanceId: "control-plane-01",
				ConnectTime:            util_proto.MustTimestampProto(t0),
			}
			accessor := &SubscriptionStatusHolder{key, subscription}
			ticks := make(chan time.Time)
			ticker := &time.Ticker{
				C: ticks,
			}
			var latestUpsert *DataplaneInsightUpsert

			// given
			sink := NewDataplaneInsightSink(accessor, func() *time.Ticker { return ticker }, recorder)
			go sink.Start(stop)

			// when
			ticks <- t0.Add(1 * time.Second)
			// then
			Eventually(func() bool {
				select {
				case upsert, ok := <-recorder.Upserts:
					latestUpsert = &upsert
					return ok
				default:
					return false
				}
			}, "1s", "1ms").Should(BeTrue())
			// and
			Expect(util_proto.ToYAML(latestUpsert.DiscoverySubscription)).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: control-plane-01
            id: 3287995C-7E11-41FB-9479-7D39337F845D
            status:
              cds: {}
              eds: {}
              lds: {}
              rds: {}
              total: {}
`))

			// when - time tick after changes
			subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(t0.Add(2 * time.Second))
			subscription.Status.Lds.ResponsesSent += 1
			subscription.Status.Total.ResponsesSent += 1
			// and
			ticks <- t0.Add(2 * time.Second)
			// then
			Eventually(func() bool {
				select {
				case upsert, ok := <-recorder.Upserts:
					latestUpsert = &upsert
					return ok
				default:
					return false
				}
			}, "1s", "1ms").Should(BeTrue())
			// and
			Expect(util_proto.ToYAML(latestUpsert.DiscoverySubscription)).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: control-plane-01
            id: 3287995C-7E11-41FB-9479-7D39337F845D
            status:
              lastUpdateTime: "2019-07-01T00:00:02Z"
              cds: {}
              eds: {}
              lds:
                responsesSent: "1"
              rds: {}
              total:
                responsesSent: "1"
`))

			// when - time tick without changes
			ticks <- t0.Add(3 * time.Second)
			// then
			select {
			case _, _ = <-recorder.Upserts:
				Fail("time tick should not lead to status update")
			case <-time.After(100 * time.Millisecond):
				// no update is good
			}
		})
	})

	Describe("DataplaneInsightStore", func() {

		var store core_store.ResourceStore

		BeforeEach(func() {
			store = memory_resources.NewStore()
			err := store.Create(context.Background(), &mesh_core.MeshResource{}, core_store.CreateByKey("default", "default", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create/update DataplaneInsight resource", func() {
			// setup
			key := core_model.ResourceKey{Mesh: "default", Namespace: "demo", Name: "example-001"}
			subscription := &mesh_proto.DiscoverySubscription{
				Id:                     "3287995C-7E11-41FB-9479-7D39337F845D",
				ControlPlaneInstanceId: "control-plane-01",
				ConnectTime:            util_proto.MustTimestampProto(t0),
			}
			dataplaneInsight := &mesh_core.DataplaneInsightResource{}
			lastSeenVersion := ""

			// given
			statusStore := NewDataplaneInsightStore(manager.NewResourceManager(store, test_resources.Global()))

			// when
			err := statusStore.Upsert(key, proto.Clone(subscription).(*mesh_proto.DiscoverySubscription))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Eventually(func() bool {
				err := store.Get(context.Background(), dataplaneInsight, core_store.GetBy(key))
				if err != nil {
					return false
				}
				if dataplaneInsight.Meta.GetVersion() == lastSeenVersion {
					return false
				}
				lastSeenVersion = dataplaneInsight.Meta.GetVersion()
				return true
			}, "1s", "1ms").Should(BeTrue())
			// and
			Expect(util_proto.ToYAML(dataplaneInsight.GetSpec())).To(MatchYAML(`
            subscriptions:
            - connectTime: "2019-07-01T00:00:00Z"
              controlPlaneInstanceId: control-plane-01
              id: 3287995C-7E11-41FB-9479-7D39337F845D
              status:
                cds: {}
                eds: {}
                lds: {}
                rds: {}
                total: {}
`))

			// when
			subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(t0.Add(2 * time.Second))
			subscription.Status.Lds.ResponsesSent += 1
			subscription.Status.Total.ResponsesSent += 1
			// and
			err = statusStore.Upsert(key, proto.Clone(subscription).(*mesh_proto.DiscoverySubscription))
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Eventually(func() bool {
				err := store.Get(context.Background(), dataplaneInsight, core_store.GetBy(key))
				if err != nil {
					return false
				}
				if dataplaneInsight.Meta.GetVersion() == lastSeenVersion {
					return false
				}
				lastSeenVersion = dataplaneInsight.Meta.GetVersion()
				return true
			}, "1s", "1ms").Should(BeTrue())
			// and
			Expect(util_proto.ToYAML(dataplaneInsight.GetSpec())).To(MatchYAML(`
            subscriptions:
            - connectTime: "2019-07-01T00:00:00Z"
              controlPlaneInstanceId: control-plane-01
              id: 3287995C-7E11-41FB-9479-7D39337F845D
              status:
                lastUpdateTime: "2019-07-01T00:00:02Z"
                cds: {}
                eds: {}
                lds:
                  responsesSent: "1"
                rds: {}
                total:
                  responsesSent: "1"
`))
		})
	})
})

var _ SubscriptionStatusAccessor = &SubscriptionStatusHolder{}

type SubscriptionStatusHolder struct {
	core_model.ResourceKey
	*mesh_proto.DiscoverySubscription
}

func (h *SubscriptionStatusHolder) GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription) {
	return h.ResourceKey, proto.Clone(h.DiscoverySubscription).(*mesh_proto.DiscoverySubscription)
}

var _ DataplaneInsightStore = &DataplaneInsightStoreRecorder{}

type DataplaneInsightUpsert struct {
	core_model.ResourceKey
	*mesh_proto.DiscoverySubscription
}

type DataplaneInsightStoreRecorder struct {
	Upserts chan DataplaneInsightUpsert
}

func (s *DataplaneInsightStoreRecorder) Upsert(dataplaneId core_model.ResourceKey, subscription *mesh_proto.DiscoverySubscription) error {
	s.Upserts <- DataplaneInsightUpsert{dataplaneId, subscription}
	return nil
}
