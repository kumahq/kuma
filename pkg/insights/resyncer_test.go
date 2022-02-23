package insights_test

import (
	"context"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/insights"
	test_insights "github.com/kumahq/kuma/pkg/insights/test"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
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
			Registry: registry.Global(),
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

	It("should count dataplanes by type", func() {
		// setup
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneInsightResource()
		dp1.Spec.Subscriptions = append(dp1.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(1),
		})
		err = rm.Create(context.Background(), dp1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := core_mesh.NewDataplaneInsightResource()
		dp2.Spec.Subscriptions = append(dp2.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(2),
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
			DisconnectTime: &timestamppb.Timestamp{
				Seconds: 101,
				Nanos:   202,
			},
		})
		err = rm.Create(context.Background(), dp2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp3 := core_mesh.NewDataplaneInsightResource()
		dp3.Spec.Subscriptions = append(dp3.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(3),
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})
		err = rm.Create(context.Background(), dp3, store.CreateByKey("dp3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.GatewayDataplane}, store.CreateByKey("dp4", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp4 := core_mesh.NewDataplaneInsightResource()
		dp4.Spec.Subscriptions = append(dp4.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			Id: strconv.Itoa(4),
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})
		err = rm.Create(context.Background(), dp4, store.CreateByKey("dp4", "mesh-1"))
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
		standardDP := meshInsight.Spec.GetDataplanesByType().GetStandard()
		Expect(standardDP.GetTotal()).To(Equal(uint32(3)))
		Expect(standardDP.GetOnline()).To(Equal(uint32(1)))
		Expect(standardDP.GetOffline()).To(Equal(uint32(2)))

		gatewayDP := meshInsight.Spec.GetDataplanesByType().GetGateway()
		Expect(gatewayDP.GetTotal()).To(Equal(uint32(1)))
		Expect(gatewayDP.GetOffline()).To(Equal(uint32(0)))
		Expect(gatewayDP.GetOnline()).To(Equal(uint32(1)))
	})

	It("should count dataplanes by mTLS backends", func() {
		// given mesh
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// and dp1 with ca-1 backend
		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneInsightResource()
		dp1.Spec.MTLS = &mesh_proto.DataplaneInsight_MTLS{
			IssuedBackend:     "ca-1",
			SupportedBackends: []string{"ca-1"},
		}
		err = rm.Create(context.Background(), dp1, store.CreateByKey("dp1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		// and dp2 with ca-2 backend
		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := core_mesh.NewDataplaneInsightResource()
		dp2.Spec.MTLS = &mesh_proto.DataplaneInsight_MTLS{
			IssuedBackend:     "ca-2",
			SupportedBackends: []string{"ca-1", "ca-2"},
		}
		dp2.Spec.Subscriptions = append(dp2.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
			},
		})
		err = rm.Create(context.Background(), dp2, store.CreateByKey("dp2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		// when resyncer generates insight
		nowMtx.Lock()
		now = now.Add(60 * time.Second)
		nowMtx.Unlock()
		tickCh <- now

		meshInsight := core_mesh.NewMeshInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
		}, "10s", "100ms").Should(BeNil())

		// then
		Expect(meshInsight.Spec.MTLS.IssuedBackends).To(HaveLen(2))
		Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-1"].Total).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-1"].Offline).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-2"].Total).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-2"].Online).To(Equal(uint32(1)))

		Expect(meshInsight.Spec.MTLS.SupportedBackends).To(HaveLen(2))
		Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Total).To(Equal(uint32(2)))
		Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Offline).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Online).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-2"].Total).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-2"].Online).To(Equal(uint32(1)))
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

	It("should return correct amounts of internal/external services in mesh insights", func() {
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

		es1 := core_mesh.NewExternalServiceResource()
		es1.Spec = &mesh_proto.ExternalService{
			Networking: &mesh_proto.ExternalService_Networking{
				Address: "example.com:80",
			},
			Tags: map[string]string{
				"kuma.io/service": "externalService1",
			},
		}

		err = rm.Create(context.Background(), es1, store.CreateByKey("es1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		es2 := core_mesh.NewExternalServiceResource()
		es2.Spec = &mesh_proto.ExternalService{
			Networking: &mesh_proto.ExternalService_Networking{
				Address: "kuma.io:80",
			},
			Tags: map[string]string{
				"kuma.io/service": "externalService2",
			},
		}

		err = rm.Create(context.Background(), es2, store.CreateByKey("es2", "mesh-1"))
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
		Expect(meshInsight.Spec.Services.Total).To(Equal(uint32(4)))
		Expect(meshInsight.Spec.Services.Internal).To(Equal(uint32(2)))
		Expect(meshInsight.Spec.Services.External).To(Equal(uint32(2)))
	})

	It("should return gateway in services", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		dpOnline := core_mesh.NewDataplaneResource()
		dpOnline.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.1",
				Gateway: &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{"kuma.io/service": "gateway"},
				},
			},
		}
		err = rm.Create(context.Background(), dpOnline, store.CreateByKey("dpOnline", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpOnlineI := core_mesh.NewDataplaneInsightResource()
		dpOnlineI.Spec.Subscriptions = append(dpOnlineI.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
		})
		err = rm.Create(context.Background(), dpOnlineI, store.CreateByKey("dpOnline", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpOffline := core_mesh.NewDataplaneResource()
		dpOffline.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.1",
				Gateway: &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{"kuma.io/service": "gateway"},
				},
			},
		}
		err = rm.Create(context.Background(), dpOffline, store.CreateByKey("dpOffline", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpOfflineI := core_mesh.NewDataplaneInsightResource()
		dpOfflineI.Spec.Subscriptions = append(dpOfflineI.Spec.Subscriptions, &mesh_proto.DiscoverySubscription{
			ConnectTime: &timestamppb.Timestamp{
				Seconds: 100,
				Nanos:   200,
			},
			DisconnectTime: &timestamppb.Timestamp{
				Seconds: 101,
			},
		})
		err = rm.Create(context.Background(), dpOfflineI, store.CreateByKey("dpOffline", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dpNoInsights := core_mesh.NewDataplaneResource()
		dpNoInsights.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.0.0.1",
				Gateway: &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{"kuma.io/service": "gateway"},
				},
			},
		}
		err = rm.Create(context.Background(), dpNoInsights, store.CreateByKey("dpNoInsights", "mesh-1"))
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
		Expect(meshInsight.Spec.Dataplanes.Total).To(Equal(uint32(3)))
		Expect(meshInsight.Spec.Dataplanes.Online).To(Equal(uint32(1)))
		Expect(meshInsight.Spec.Dataplanes.PartiallyDegraded).To(Equal(uint32(0)))
		Expect(meshInsight.Spec.Dataplanes.Offline).To(Equal(uint32(2)))

		serviceInsight := core_mesh.NewServiceInsightResource()
		Eventually(func() error {
			return rm.Get(context.Background(), serviceInsight, store.GetBy(insights.ServiceInsightKey("mesh-1")))
		}, "10s", "100ms").Should(BeNil())

		Expect(serviceInsight.Spec.Services).To(HaveKey("gateway"))
		Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Total).To(Equal(uint32(3)))
		Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Online).To(Equal(uint32(1)))
		Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Offline).To(Equal(uint32(2)))
		Expect(serviceInsight.Spec.Services["gateway"].Status).To(Equal(mesh_proto.ServiceInsight_Service_partially_degraded))
	})
})
