package insights_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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
	start := time.Now()

	var stopCh chan struct{}
	var tickCh chan time.Time

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())

		stopCh = make(chan struct{})
		tickCh = make(chan time.Time)

		resyncer := insights.NewResyncer(&insights.Config{
			MinResyncTimeout:   5 * time.Second,
			MaxResyncTimeout:   1 * time.Minute,
			ResourceManager:    rm,
			EventReaderFactory: &test_insights.TestEventReaderFactory{Reader: &test_insights.TestEventReader{Ch: make(chan events.Event)}},
			Tick: func(d time.Duration) (rv <-chan time.Time) {
				Expect(d).To(Equal(55 * time.Second)) // should be equal MaxResyncTimeout - MinResyncTimeout
				return tickCh
			},
			Now: func() time.Time {
				Fail("This shouldn't be used as not tests use the event factory")
				return start
			},
			Registry: registry.Global(),
			AddressPortGenerator: func(s string) string {
				return fmt.Sprintf("%s.mesh:80", s)
			},
		})
		go func() {
			err := resyncer.Start(stopCh)
			Expect(err).ToNot(HaveOccurred())
		}()
	})
	AfterEach(func() {
		close(stopCh)
	})

	It("should sync more often than MaxResyncTimeout", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.TrafficPermissionResource{Spec: samples.TrafficPermission}, store.CreateByKey("tp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		tickCh <- start.Add(61 * time.Second)

		Eventually(func(g Gomega) {
			insight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), insight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(insight.Spec.Policies[string(core_mesh.TrafficPermissionType)].Total).To(Equal(uint32(1)))
		}).Should(Succeed())
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

		tickCh <- start.Add(60 * time.Second)

		// when
		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())

			// then
			kumaDp := meshInsight.Spec.DpVersions.KumaDp
			g.Expect(kumaDp["unknown"].Total).To(Equal(uint32(1)))
			g.Expect(kumaDp["unknown"].Offline).To(Equal(uint32(1)))
			g.Expect(kumaDp["1.0.3"].Total).To(Equal(uint32(1)))
			g.Expect(kumaDp["1.0.3"].Offline).To(Equal(uint32(1)))
			g.Expect(kumaDp["1.0.4"].Total).To(Equal(uint32(1)))
			g.Expect(kumaDp["1.0.4"].Offline).To(Equal(uint32(1)))

			envoy := meshInsight.Spec.DpVersions.Envoy
			g.Expect(envoy["unknown"].Total).To(Equal(uint32(1)))
			g.Expect(envoy["unknown"].Offline).To(Equal(uint32(1)))
			g.Expect(envoy["1.15.0"].Total).To(Equal(uint32(2)))
			g.Expect(envoy["1.15.0"].Offline).To(Equal(uint32(2)))
		}).Should(Succeed())
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

		tickCh <- start.Add(time.Minute)

		// when
		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())

			// then
			standardDP := meshInsight.Spec.GetDataplanesByType().GetStandard()
			g.Expect(standardDP.GetTotal()).To(Equal(uint32(3)))
			g.Expect(standardDP.GetOnline()).To(Equal(uint32(1)))
			g.Expect(standardDP.GetOffline()).To(Equal(uint32(2)))

			gatewayDP := meshInsight.Spec.GetDataplanesByType().GetGateway()
			g.Expect(gatewayDP.GetTotal()).To(Equal(uint32(1)))
			g.Expect(gatewayDP.GetOffline()).To(Equal(uint32(0)))
			g.Expect(gatewayDP.GetOnline()).To(Equal(uint32(1)))
		}).Should(Succeed())
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
		tickCh <- start.Add(time.Minute)

		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())

			// then
			g.Expect(meshInsight.Spec.MTLS.IssuedBackends).To(HaveLen(2))
			g.Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-1"].Total).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-1"].Offline).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-2"].Total).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.IssuedBackends["ca-2"].Online).To(Equal(uint32(1)))

			g.Expect(meshInsight.Spec.MTLS.SupportedBackends).To(HaveLen(2))
			g.Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Total).To(Equal(uint32(2)))
			g.Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Offline).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-1"].Online).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-2"].Total).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.MTLS.SupportedBackends["ca-2"].Online).To(Equal(uint32(1)))
		}).Should(Succeed())
	})

	It("should not count dataplane as a policy", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &core_mesh.DataplaneResource{Spec: samples.Dataplane}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		tickCh <- start.Add(time.Minute + time.Second)

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

		tickCh <- start.Add(time.Minute + time.Second)

		// when
		Eventually(func(g Gomega) {
			serviceInsight := core_mesh.NewServiceInsightResource()
			err := rm.Get(context.Background(), serviceInsight, store.GetByKey("all-services-mesh-1", "mesh-1"))
			g.Expect(err).ToNot(HaveOccurred())

			service := serviceInsight.Spec.Services["backend-1"]
			// then
			g.Expect(service.Status).To(Equal(mesh_proto.ServiceInsight_Service_partially_degraded))
			g.Expect(service.Dataplanes.Total).To(Equal(uint32(4)))
			g.Expect(service.Dataplanes.Online).To(Equal(uint32(2)))
			g.Expect(service.Dataplanes.Offline).To(Equal(uint32(2)))
		}).Should(Succeed())
	})

	It("should return correct service types in serviceInsights", func() {
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		dp1 := core_mesh.NewDataplaneResource()
		dp1.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "10.0.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 5000,
						Tags: map[string]string{
							"kuma.io/service": "internal",
						},
					},
				},
			},
		}
		builtinGw := core_mesh.NewDataplaneResource()
		builtinGw.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "10.0.0.1",
				Gateway: &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{
						"kuma.io/service": "gw1",
					},
					Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
				},
			},
		}
		delegatedGw := core_mesh.NewDataplaneResource()
		delegatedGw.Spec = &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "10.0.0.1",
				Gateway: &mesh_proto.Dataplane_Networking_Gateway{
					Tags: map[string]string{
						"kuma.io/service": "gw2",
					},
					Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
				},
			},
		}
		externalService := core_mesh.NewExternalServiceResource()
		externalService.Spec = &mesh_proto.ExternalService{
			Tags: map[string]string{"kuma.io/service": "external-service"},
			Networking: &mesh_proto.ExternalService_Networking{
				Address: "foobar:8080",
			},
		}

		for n, entity := range map[string]model.Resource{"dp1": dp1, "dgw": delegatedGw, "bgw": builtinGw, "externalService": externalService} {
			err = rm.Create(context.Background(), entity, store.CreateByKey(n, "mesh-1"))
			Expect(err).ToNot(HaveOccurred(), n)
		}

		tickCh <- start.Add(time.Minute + time.Second)

		// when
		Eventually(func(g Gomega) {
			serviceInsight := core_mesh.NewServiceInsightResource()
			err := rm.Get(context.Background(), serviceInsight, store.GetBy(insights.ServiceInsightKey("mesh-1")))
			g.Expect(err).ToNot(HaveOccurred())
			internal := serviceInsight.Spec.Services["internal"]
			g.Expect(internal).ToNot(BeNil())
			g.Expect(internal.ServiceType).To(Equal(mesh_proto.ServiceInsight_Service_internal))
			g.Expect(internal.AddressPort).To(Equal("internal.mesh:80"))

			gw1 := serviceInsight.Spec.Services["gw1"]
			g.Expect(gw1).ToNot(BeNil())
			g.Expect(gw1.ServiceType).To(Equal(mesh_proto.ServiceInsight_Service_gateway_builtin))
			g.Expect(gw1.AddressPort).To(Equal(""))

			gw2 := serviceInsight.Spec.Services["gw2"]
			g.Expect(gw2).ToNot(BeNil())
			g.Expect(gw2.ServiceType).To(Equal(mesh_proto.ServiceInsight_Service_gateway_delegated))
			g.Expect(gw2.AddressPort).To(Equal(""))

			ext := serviceInsight.Spec.Services["external-service"]
			g.Expect(ext).ToNot(BeNil())
			g.Expect(ext.ServiceType).To(Equal(mesh_proto.ServiceInsight_Service_external))
			g.Expect(ext.AddressPort).To(Equal("foobar:8080"))
		}).Should(Succeed())
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

		tickCh <- start.Add(time.Minute + time.Second)

		// when
		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(meshInsight.Spec.Dataplanes.Total).To(Equal(uint32(4)))
			g.Expect(meshInsight.Spec.Dataplanes.Online).To(Equal(uint32(2)))
			g.Expect(meshInsight.Spec.Dataplanes.PartiallyDegraded).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.Dataplanes.Offline).To(Equal(uint32(1)))
		}).Should(Succeed())

		// then
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

		tickCh <- start.Add(time.Minute + time.Second)

		// when
		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(meshInsight.Spec.Services.Total).To(Equal(uint32(4)))
			g.Expect(meshInsight.Spec.Services.Internal).To(Equal(uint32(2)))
			g.Expect(meshInsight.Spec.Services.External).To(Equal(uint32(2)))
		}).Should(Succeed())
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

		tickCh <- start.Add(61 * time.Second)

		// when
		Eventually(func(g Gomega) {
			meshInsight := core_mesh.NewMeshInsightResource()
			err := rm.Get(context.Background(), meshInsight, store.GetByKey("mesh-1", model.NoMesh))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(meshInsight.Spec.Dataplanes.Total).To(Equal(uint32(3)))
			g.Expect(meshInsight.Spec.Dataplanes.Online).To(Equal(uint32(1)))
			g.Expect(meshInsight.Spec.Dataplanes.PartiallyDegraded).To(Equal(uint32(0)))
			g.Expect(meshInsight.Spec.Dataplanes.Offline).To(Equal(uint32(2)))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			serviceInsight := core_mesh.NewServiceInsightResource()
			err := rm.Get(context.Background(), serviceInsight, store.GetBy(insights.ServiceInsightKey("mesh-1")))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(serviceInsight.Spec.Services).To(HaveKey("gateway"))
			g.Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Total).To(Equal(uint32(3)))
			g.Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Online).To(Equal(uint32(1)))
			g.Expect(serviceInsight.Spec.Services["gateway"].Dataplanes.Offline).To(Equal(uint32(2)))
			g.Expect(serviceInsight.Spec.Services["gateway"].Status).To(Equal(mesh_proto.ServiceInsight_Service_partially_degraded))
		}).Should(Succeed())
	})
})
