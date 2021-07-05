package insights_test

import (
	"context"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights"
	test_insights "github.com/kumahq/kuma/pkg/insights/test"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Insight Persistence", func() {
	var rm manager.ResourceManager
	nowMtx := &sync.RWMutex{}
	var now time.Time

	var eventCh chan events.Event
	var stopCh chan struct{}

	tickMtx := &sync.RWMutex{}
	var tickCh chan time.Time

	core.Now = func() time.Time {
		nowMtx.RLock()
		defer nowMtx.RUnlock()
		return now
	}

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())

		nowMtx.Lock()
		now = time.Now()
		nowMtx.Unlock()

		eventCh = make(chan events.Event)
		stopCh = make(chan struct{})

		tickMtx.Lock()
		tickCh = make(chan time.Time)
		tickMtx.Unlock()

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncTimeout:   5 * time.Second,
			MaxResyncTimeout:   1 * time.Minute,
			ResourceManager:    rm,
			EventReaderFactory: &test_insights.TestEventReaderFactory{Reader: &test_insights.TestEventReader{Ch: eventCh}},
			Tick: func(d time.Duration) (rv <-chan time.Time) {
				tickMtx.RLock()
				defer tickMtx.RUnlock()
				Expect(d).To(Equal(55 * time.Second)) // should be equal MaxResyncTimeout - MinResyncTimeout
				return tickCh
			},
		})
		go func(stopCh chan struct{}) {
			err := resyncer.Start(stopCh)
			Expect(err).ToNot(HaveOccurred())
		}(stopCh)
	})

	It("should sync more often than MaxResyncTimeout", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(61 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		insight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), insight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())
		Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(MatchProto(proto.MustTimestampProto(now)))
	})

	It("should count dataplanes by version", func() {
		// setup
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneInsightResource()
		dp1.Spec.Subscriptions = append(dp1.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(1),
			Version: &mesh_proto.Version{
				KumaDp: &mesh_proto.KumaDpVersion{
					Version: "1.0.3",
				},
				Envoy: &mesh_proto.EnvoyVersion{
					Version: "1.15.0",
				},
			},
		})
		err = rm.Create(context.Background(), dp1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := core_mesh.NewDataplaneInsightResource()
		dp2.Spec.Subscriptions = append(dp2.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(2),
			Version: &mesh_proto.Version{
				KumaDp: &mesh_proto.KumaDpVersion{
					Version: "1.0.4",
				},
				Envoy: &mesh_proto.EnvoyVersion{
					Version: "1.15.0",
				},
			},
		})
		err = rm.Create(context.Background(), dp2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp3 := core_mesh.NewDataplaneInsightResource()
		dp3.Spec.Subscriptions = append(dp3.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(3),
		})
		err = rm.Create(context.Background(), dp3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(60 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		// when
		meshInsight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())

		// then
		kumaDp := meshInsight.Spec.DpVersions.KumaDp
		Expect(kumaDp["unknown"].Total).To(Equal(uint32(1)))
		Expect(kumaDp["unknown"].Offline).To(Equal(uint32(1)))
		Expect(kumaDp["1.0.3"].Total).To(Equal(uint32(1)))
		Expect(kumaDp["1.0.3"].Offline).To(Equal(uint32(1)))
		Expect(kumaDp["1.0.4"].Total).To(Equal(uint32(1)))
		Expect(kumaDp["1.0.4"].Offline).To(Equal(uint32(1)))

		envoy := meshInsight.Spec.DpVersions.Envoy
		Expect(envoy["unknown"].Total).To(Equal(uint32(1)))
		Expect(envoy["unknown"].Offline).To(Equal(uint32(1)))
		Expect(envoy["1.15.0"].Total).To(Equal(uint32(2)))
		Expect(envoy["1.15.0"].Offline).To(Equal(uint32(2)))
	})

	It("should not count dataplane as a policy", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(61 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		insight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), insight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())
		Expect(insight.Spec.Policies[string(core_mesh.DataplaneType)]).To(BeNil())
		Expect(insight.Spec.Dataplanes.Total).To(Equal(uint32(1)))
		Expect(insight.Spec.LastSync).To(MatchProto(proto.MustTimestampProto(now)))
	})

	It("should return correct statuses in service insights", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneResource()
		dp1.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend-1",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi1 := core_mesh.NewDataplaneInsightResource()
		dpi1.Spec.Subscriptions = append(dpi1.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := core_mesh.NewDataplaneResource()
		dp2.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.2",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend-1",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi2 := core_mesh.NewDataplaneInsightResource()
		dpi2.Spec.Subscriptions = append(dpi2.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp3 := core_mesh.NewDataplaneResource()
		dp3.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.3",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: false,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend-1",
						},
					},
					{
						Port: 8888,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "db-1",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi3 := core_mesh.NewDataplaneInsightResource()
		dpi3.Spec.Subscriptions = append(dpi3.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp4 := core_mesh.NewDataplaneResource()
		dp4.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.4",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend-1",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp4, store.CreateByKey("dp4", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi4 := core_mesh.NewDataplaneInsightResource()
		dpi4.Spec.Subscriptions = append(dpi4.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
			DisconnectTime: &timestamppb.Timestamp{
				Seconds: 101,
				Nanos:   202,
			},
		})

		err = rm.Create(context.Background(), dpi4, store.CreateByKey("dp4", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(61 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		// when
		serviceInsight := core_mesh.NewServiceInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), serviceInsight, store.GetByKey("all-services-mesh-1", "mesh-1"))
		}, "10s", "100ms").Should(BeNil())

		service := serviceInsight.Spec.Services["backend-1"]

		// then
		Expect(service.Status).To(Equal(mesh_proto.ServiceInsight_Service_partially_degraded))
		Expect(service.Dataplanes.Total).To(Equal(uint32(4)))
		Expect(service.Dataplanes.Online).To(Equal(uint32(2)))
		Expect(service.Dataplanes.Offline).To(Equal(uint32(2)))
	})

	It("should return correct dataplanes statuses in mesh insights", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneResource()
		dp1.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi1 := core_mesh.NewDataplaneInsightResource()
		dpi1.Spec.Subscriptions = append(dpi1.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := core_mesh.NewDataplaneResource()
		dp2.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.2",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend",
						},
					},
					{
						Port: 8888,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "db",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi2 := core_mesh.NewDataplaneInsightResource()
		dpi2.Spec.Subscriptions = append(dpi2.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp3 := core_mesh.NewDataplaneResource()
		dp3.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.3",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend",
						},
					},
					{
						Port: 8888,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: false,
						},
						Tags: map[string]string{
							"kuma.io/service": "db",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi3 := core_mesh.NewDataplaneInsightResource()
		dpi3.Spec.Subscriptions = append(dpi3.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})

		err = rm.Create(context.Background(), dpi3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp4 := core_mesh.NewDataplaneResource()
		dp4.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.3",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 7777,
						Health: &mesh_proto.Dataplane_Networking_Inbound_Health{
							Ready: true,
						},
						Tags: map[string]string{
							"kuma.io/service": "backend",
						},
					},
				},
			},
		}

		err = rm.Create(context.Background(), dp4, store.CreateByKey("dp4", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpi4 := core_mesh.NewDataplaneInsightResource()
		dpi4.Spec.Subscriptions = append(dpi4.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
			DisconnectTime: &timestamppb.Timestamp{
				Seconds: 101,
				Nanos:   202,
			},
		})

		err = rm.Create(context.Background(), dpi4, store.CreateByKey("dp4", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		nowMtx.Lock()
		now = now.Add(61 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		// when
		meshInsight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())

		// then
		Expect(meshInsight.Spec.Dataplanes.Total).To(Equal(uint32(4)))
		Expect(meshInsight.Spec.Dataplanes.Online).To(Equal(uint32(2)))
		Expect(meshInsight.Spec.Dataplanes.PartiallyDegraded).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.Dataplanes.Offline).To(Equal(uint32(1)))
	})
})
