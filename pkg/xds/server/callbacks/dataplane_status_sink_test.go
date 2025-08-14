package callbacks_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/events"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/server/callbacks"
)

var _ = Describe("DataplaneInsightSink", func() {
	t0, _ := time.Parse(time.RFC3339, "2019-07-01T00:00:00+00:00")

	Describe("DataplaneInsightSink", func() {
		var recorder *DataplaneInsightStoreRecorder
		var store callbacks.DataplaneInsightStore
		var stop chan struct{}

		BeforeEach(func() {
			recorder = &DataplaneInsightStoreRecorder{
				ResourceManager: manager.NewResourceManager(memory_resources.NewStore()),
				Creates:         make(chan DataplaneInsightOperation),
				Updates:         make(chan DataplaneInsightOperation),
			}
			Expect(
				recorder.ResourceManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("default", core_model.NoMesh)),
			).To(Succeed())
			store = callbacks.NewDataplaneInsightStore(recorder)
			stop = make(chan struct{})
		})

		AfterEach(func() {
			close(stop)
		})

		It("should periodically flush DataplaneInsight into a store", func() {
			// setup
			key := core_model.ResourceKey{Mesh: "default", Name: "example-001"}
			subscription := &mesh_proto.DiscoverySubscription{
				Id:                     "3287995C-7E11-41FB-9479-7D39337F845D",
				ControlPlaneInstanceId: "control-plane-01",
				ConnectTime:            util_proto.MustTimestampProto(t0),
				Status:                 mesh_proto.NewSubscriptionStatus(t0),
			}
			accessor := &SubscriptionStatusHolder{key, subscription}
			ticks := make(chan time.Time)
			ticker := &time.Ticker{
				C: ticks,
			}
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			eventBus, err := events.NewEventBus(10, metrics)
			Expect(err).ToNot(HaveOccurred())

			var latestOperation *DataplaneInsightOperation

			// given
			sink := callbacks.NewDataplaneInsightSink(
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						core_xds.FieldDataplaneProxyType: {
							Kind: &structpb.Value_StringValue{
								StringValue: string(mesh_proto.DataplaneProxyType),
							},
						},
					},
				},
				accessor,
				&xds.TestSecrets{},
				func() *time.Ticker { return ticker },
				func() *time.Ticker { return &time.Ticker{C: make(chan time.Time)} },
				1*time.Millisecond,
				store,
				eventBus,
				recorder.ResourceManager,
			)

			// when
			go sink.Start(stop)

			// then
			create, ok := <-recorder.Creates
			Expect(ok).To(BeTrue())
			latestOperation = &create

			// and
			Expect(util_proto.ToYAML(latestOperation.Subscriptions[len(latestOperation.Subscriptions)-1])).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: control-plane-01
            id: 3287995C-7E11-41FB-9479-7D39337F845D
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:00Z"
              lds: {}
              rds: {}
              total: {}
`))

			// and
			Expect(latestOperation.DataplaneInsight_MTLS.IssuedBackend).To(Equal(xds.TestSecretsInfo.IssuedBackend))
			Expect(latestOperation.DataplaneInsight_MTLS.SupportedBackends).To(Equal(xds.TestSecretsInfo.SupportedBackends))

			// when - time tick after changes
			subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(t0.Add(2 * time.Second))
			subscription.Status.Lds.ResponsesSent += 1
			subscription.Status.Total.ResponsesSent += 1
			// and
			ticks <- t0.Add(2 * time.Second)
			// then
			Eventually(func() bool {
				select {
				case update, ok := <-recorder.Updates:
					latestOperation = &update
					return ok
				default:
					return false
				}
			}, "1s", "1ms").Should(BeTrue())
			// and
			Expect(util_proto.ToYAML(latestOperation.Subscriptions[len(latestOperation.Subscriptions)-1])).To(MatchYAML(`
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
			case <-recorder.Creates:
				Fail("time tick should not lead to status update")
			case <-recorder.Updates:
				Fail("time tick should not lead to status update")
			case <-time.After(100 * time.Millisecond):
				// no update is good
			}
		})

		FIt("should periodically flush DataplaneInsight into a store with mTLS status based on events", func() {
			// setup
			key := core_model.ResourceKey{Mesh: "default", Name: "example-001"}
			subscription := &mesh_proto.DiscoverySubscription{
				Id:                     "3287995C-7E11-41FB-9479-7D39337F845D",
				ControlPlaneInstanceId: "control-plane-01",
				ConnectTime:            util_proto.MustTimestampProto(t0),
				Status:                 mesh_proto.NewSubscriptionStatus(t0),
			}
			accessor := &SubscriptionStatusHolder{key, subscription}
			ticks := make(chan time.Time)
			ticker := &time.Ticker{
				C: ticks,
			}
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			eventBus, err := events.NewEventBus(10, metrics)
			Expect(err).ToNot(HaveOccurred())

			var latestOperation *DataplaneInsightOperation

			// given
			sink := callbacks.NewDataplaneInsightSink(
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						core_xds.FieldDataplaneProxyType: {
							Kind: &structpb.Value_StringValue{
								StringValue: string(mesh_proto.DataplaneProxyType),
							},
						},
					},
				},
				accessor,
				&xds.TestSecrets{NoSecrets: true}, // let's use events
				func() *time.Ticker { return ticker },
				func() *time.Ticker { return &time.Ticker{C: make(chan time.Time)} },
				1*time.Millisecond,
				store,
				eventBus,
				recorder.ResourceManager,
			)

			// when
			go sink.Start(stop)

			// then
			create, ok := <-recorder.Creates
			Expect(ok).To(BeTrue())
			latestOperation = &create

			// and
			Expect(util_proto.ToYAML(latestOperation.Subscriptions[len(latestOperation.Subscriptions)-1])).To(MatchYAML(`
            connectTime: "2019-07-01T00:00:00Z"
            controlPlaneInstanceId: control-plane-01
            id: 3287995C-7E11-41FB-9479-7D39337F845D
            status:
              cds: {}
              eds: {}
              lastUpdateTime: "2019-07-01T00:00:00Z"
              lds: {}
              rds: {}
              total: {}
`))

			// and
			Expect(latestOperation.DataplaneInsight_MTLS).To(BeNil())

// 			// when - time tick after changes
// 			subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(t0.Add(2 * time.Second))
// 			subscription.Status.Lds.ResponsesSent += 1
// 			subscription.Status.Total.ResponsesSent += 1
// 			// and
// 			ticks <- t0.Add(2 * time.Second)
// 			// then
// 			Eventually(func() bool {
// 				select {
// 				case update, ok := <-recorder.Updates:
// 					latestOperation = &update
// 					return ok
// 				default:
// 					return false
// 				}
// 			}, "1s", "1ms").Should(BeTrue())
// 			// and
// 			Expect(util_proto.ToYAML(latestOperation.Subscriptions[len(latestOperation.Subscriptions)-1])).To(MatchYAML(`
//             connectTime: "2019-07-01T00:00:00Z"
//             controlPlaneInstanceId: control-plane-01
//             id: 3287995C-7E11-41FB-9479-7D39337F845D
//             status:
//               lastUpdateTime: "2019-07-01T00:00:02Z"
//               cds: {}
//               eds: {}
//               lds:
//                 responsesSent: "1"
//               rds: {}
//               total:
//                 responsesSent: "1"
// `))

// 			// when - time tick without changes
// 			ticks <- t0.Add(3 * time.Second)
// 			// then
// 			select {
// 			case <-recorder.Creates:
// 				Fail("time tick should not lead to status update")
// 			case <-recorder.Updates:
// 				Fail("time tick should not lead to status update")
// 			case <-time.After(100 * time.Millisecond):
// 				// no update is good
// 			}
		})
	})

	Describe("DataplaneInsightStore", func() {
		var store core_store.ResourceStore

		BeforeEach(func() {
			store = core_store.NewPaginationStore(memory_resources.NewStore())
			err := store.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create/update DataplaneInsight resource", func() {
			ctx := context.Background()

			// setup
			key := core_model.ResourceKey{Mesh: "default", Name: "example-001"}
			subscription := &mesh_proto.DiscoverySubscription{
				Id:                     "3287995C-7E11-41FB-9479-7D39337F845D",
				ControlPlaneInstanceId: "control-plane-01",
				ConnectTime:            util_proto.MustTimestampProto(t0),
				Status:                 mesh_proto.NewSubscriptionStatus(t0),
			}
			dataplaneType := core_mesh.DataplaneType
			dataplaneInsight := core_mesh.NewDataplaneInsightResource()
			lastSeenVersion := ""

			// given
			statusStore := callbacks.NewDataplaneInsightStore(manager.NewResourceManager(store))

			// when
			err := statusStore.Upsert(ctx, nil, dataplaneType, key, proto.Clone(subscription).(*mesh_proto.DiscoverySubscription), nil)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Eventually(func() bool {
				err := store.Get(ctx, dataplaneInsight, core_store.GetBy(key))
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
			Expect(util_proto.ToYAML(dataplaneInsight.GetSpec().(proto.Message))).To(MatchYAML(`
            subscriptions:
            - connectTime: "2019-07-01T00:00:00Z"
              controlPlaneInstanceId: control-plane-01
              id: 3287995C-7E11-41FB-9479-7D39337F845D
              status:
                cds: {}
                eds: {}
                lastUpdateTime: "2019-07-01T00:00:00Z"
                lds: {}
                rds: {}
                total: {}
`))

			// when
			subscription.Status.LastUpdateTime = util_proto.MustTimestampProto(t0.Add(2 * time.Second))
			subscription.Status.Lds.ResponsesSent += 1
			subscription.Status.Total.ResponsesSent += 1
			// and
			err = statusStore.Upsert(ctx, nil, dataplaneType, key, proto.Clone(subscription).(*mesh_proto.DiscoverySubscription), nil)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Eventually(func() bool {
				err := store.Get(ctx, dataplaneInsight, core_store.GetBy(key))
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
			Expect(util_proto.ToYAML(dataplaneInsight.GetSpec().(proto.Message))).To(MatchYAML(`
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

var _ callbacks.SubscriptionStatusAccessor = &SubscriptionStatusHolder{}

type SubscriptionStatusHolder struct {
	core_model.ResourceKey
	*mesh_proto.DiscoverySubscription
}

func (h *SubscriptionStatusHolder) GetStatus() (core_model.ResourceKey, *mesh_proto.DiscoverySubscription) {
	return h.ResourceKey, proto.Clone(h.DiscoverySubscription).(*mesh_proto.DiscoverySubscription)
}

var _ manager.ResourceManager = &DataplaneInsightStoreRecorder{}

type DataplaneInsightOperation struct {
	core_model.ResourceKey
	*mesh_proto.DataplaneInsight_MTLS
	Subscriptions []*mesh_proto.DiscoverySubscription
}

type DataplaneInsightStoreRecorder struct {
	manager.ResourceManager
	Creates chan DataplaneInsightOperation
	Updates chan DataplaneInsightOperation
}

func (d *DataplaneInsightStoreRecorder) Create(ctx context.Context, resource core_model.Resource, optionsFunc ...core_store.CreateOptionsFunc) error {
	if err := d.ResourceManager.Create(ctx, resource, optionsFunc...); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(optionsFunc...)
	d.Creates <- DataplaneInsightOperation{
		ResourceKey:           core_model.ResourceKey{Mesh: opts.Mesh, Name: opts.Name},
		Subscriptions:         resource.GetSpec().(*mesh_proto.DataplaneInsight).Subscriptions,
		DataplaneInsight_MTLS: resource.GetSpec().(*mesh_proto.DataplaneInsight).MTLS,
	}
	return nil
}

func (d *DataplaneInsightStoreRecorder) Update(ctx context.Context, resource core_model.Resource, optionsFunc ...core_store.UpdateOptionsFunc) error {
	if err := d.ResourceManager.Update(ctx, resource, optionsFunc...); err != nil {
		return err
	}
	d.Updates <- DataplaneInsightOperation{
		ResourceKey:           core_model.ResourceKey{Mesh: resource.GetMeta().GetMesh(), Name: resource.GetMeta().GetName()},
		Subscriptions:         resource.GetSpec().(*mesh_proto.DataplaneInsight).Subscriptions,
		DataplaneInsight_MTLS: resource.GetSpec().(*mesh_proto.DataplaneInsight).MTLS,
	}
	return nil
}
