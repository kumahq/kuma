package global_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	sync_store_v2 "github.com/kumahq/kuma/pkg/kds/v2/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/grpc"
	kds_setup "github.com/kumahq/kuma/pkg/test/kds/setup"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Global Sync", func() {
	var zoneStores []store.ResourceStore
	var globalStore store.ResourceStore
	var closeFunc func()

	dataplaneFunc := func(zone, service string) *mesh_proto.Dataplane {
		return &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1212,
					Tags: map[string]string{
						mesh_proto.ZoneTag:    zone,
						mesh_proto.ServiceTag: service,
					},
				}},
				Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
					{
						Port: 10000,
						Tags: map[string]string{
							mesh_proto.ServiceTag:  "web",
							mesh_proto.ProtocolTag: "http",
						},
					},
				},
			},
		}
	}

	VerifyResourcesWereSynchronizedToGlobal := func() {
		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-1", fmt.Sprintf("service-1-%d", i))
			err := zoneStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-1-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}
		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(10))
		closeFunc()
	}

	VerifyResourcesWereSynchronizedIndependentlyForEachZone := func() {
		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-1", fmt.Sprintf("service-1-%d", i))
			err := zoneStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-1-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}

		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-2", fmt.Sprintf("service-2-%d", i))
			err := zoneStores[1].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-2-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "3s", "100ms").Should(Equal(20))

		err := zoneStores[0].Delete(context.Background(), &mesh.DataplaneResource{}, store.DeleteByKey("dp-1-0", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = zoneStores[0].Delete(context.Background(), &mesh.DataplaneResource{}, store.DeleteByKey("dp-1-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(18))

		closeFunc()
	}

	VerifySupportForTheSameNameOfDataplanesInDifferentClusters := func() {
		dp1 := dataplaneFunc("kuma-cluster-1", "backend")
		err := zoneStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp1}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := dataplaneFunc("kuma-cluster-2", "web")
		err = zoneStores[1].Create(context.Background(), &mesh.DataplaneResource{Spec: dp2}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "3s", "100ms").Should(Equal(2))
	}

	VerifyUpToDateListOfProvidedType := func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneType:         true,
			mesh.DataplaneOverviewType: true,
			mesh.ServiceOverviewType:   true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualProvidedTypes := registry.Global().ObjectTypes(model.HasScope(model.ScopeMesh), model.TypeFilterFn(func(descriptor model.ResourceTypeDescriptor) bool {
			return !excludeTypes[descriptor.Name]
		}))

		// plus the global-scope types
		extraTypes := []model.ResourceType{
			mesh.MeshType,
			mesh.ZoneIngressType,
			system.ConfigType,
			system.GlobalSecretType,
			hostnamegenerator_api.HostnameGeneratorType,
		}

		actualProvidedTypes = append(actualProvidedTypes, extraTypes...)
		Expect(actualProvidedTypes).To(ConsistOf(registry.Global().ObjectTypes(model.HasKDSFlag(model.GlobalToZoneSelector))))
	}

	Context("Delta KDS", func() {
		var globalSyncer sync_store_v2.ResourceSyncer

		BeforeEach(func() {
			const numOfZones = 2
			const zoneName = "zone-%d"

			// Start `numOfZones` Kuma CP Zone
			serverStreams := []*grpc.MockDeltaServerStream{}
			wg := &sync.WaitGroup{}
			zoneStores = []store.ResourceStore{}
			for i := 0; i < numOfZones; i++ {
				zoneStore := memory.NewStore()
				srv, err := kds_setup.NewKdsServerBuilder(zoneStore).AsZone(fmt.Sprintf(zoneName, i)).Delta()
				Expect(err).ToNot(HaveOccurred())
				serverStream := grpc.NewMockDeltaServerStream()
				wg.Add(1)
				go func() {
					defer func() {
						wg.Done()
						GinkgoRecover()
					}()
					Expect(srv.ZoneToGlobal(serverStream)).To(Succeed())
				}()

				serverStreams = append(serverStreams, serverStream)
				zoneStores = append(zoneStores, zoneStore)
			}

			// Start 1 Kuma CP Global
			globalStore = memory.NewStore()
			metrics, err := core_metrics.NewMetrics("")
			Expect(err).ToNot(HaveOccurred())
			globalSyncer, err = sync_store_v2.NewResourceSyncer(core.Log, globalStore, store.NoTransactions{}, metrics, context.Background())
			Expect(err).ToNot(HaveOccurred())
			stopCh := make(chan struct{})
			clientStreams := []*grpc.MockDeltaClientStream{}
			for _, ss := range serverStreams {
				clientStreams = append(clientStreams, ss.ClientStream(stopCh))
			}
			kds_setup.StartDeltaClient(clientStreams, []model.ResourceType{mesh.DataplaneType}, stopCh, sync_store_v2.GlobalSyncCallback(context.Background(), globalSyncer, false, nil, "kuma-system"))

			// Create Zone resources for each Kuma CP Zone
			for i := 0; i < numOfZones; i++ {
				zone := &system.ZoneResource{Spec: &system_proto.Zone{Enabled: util_proto.Bool(true)}}
				err := globalStore.Create(context.Background(), zone, store.CreateByKey(fmt.Sprintf(zoneName, i), model.NoMesh))
				Expect(err).ToNot(HaveOccurred())
			}

			closeFunc = func() {
				close(stopCh)
				wg.Wait()
			}
		})

		It("should add resource to global store after adding it to Zone", func() {
			VerifyResourcesWereSynchronizedToGlobal()
		})

		It("should not add resource to global store when resource is invalid", func() {
			dp := &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 1212,
						Tags: map[string]string{
							mesh_proto.ZoneTag:    "kuma-cluster-1",
							mesh_proto.ServiceTag: "service-1",
						},
					}},
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Tags: map[string]string{
								mesh_proto.ServiceTag:  "web",
								mesh_proto.ProtocolTag: "http",
							},
						},
					},
				},
			}
			err := zoneStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey("dp-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
			Consistently(func(g Gomega) {
				actual := mesh.DataplaneResourceList{}
				err := globalStore.List(context.Background(), &actual)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(actual.Items).To(BeEmpty())
			}, "1s", "100ms").Should(Succeed())
		})

		It("should sync resources independently for each Zone", func() {
			VerifyResourcesWereSynchronizedIndependentlyForEachZone()
		})

		It("should support same dataplane names through clusters", func() {
			VerifySupportForTheSameNameOfDataplanesInDifferentClusters()
		})

		It("should have up to date list of provided types", func() {
			VerifyUpToDateListOfProvidedType()
		})

		It("should sync policies from global store to the local after resource is valid", func() {
			// incorrct dp
			dp := &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 1234,
						Tags: map[string]string{
							mesh_proto.ZoneTag:    "kuma-cluster-1",
							mesh_proto.ServiceTag: "service-1",
						},
					}},
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Port: 1234,
							Tags: map[string]string{
								mesh_proto.ServiceTag:  "web",
								mesh_proto.ProtocolTag: "http",
							},
						},
					},
				},
			}
			err := zoneStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey("dp-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
			Consistently(func() int {
				actual := mesh.DataplaneResourceList{}
				err := globalStore.List(context.Background(), &actual)
				Expect(err).ToNot(HaveOccurred())
				return len(actual.Items)
			}, "5s", "100ms").Should(Equal(0))

			dp1 := mesh.NewDataplaneResource()
			err = zoneStores[0].Get(context.Background(), dp1, store.GetByKey("dp-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
			// when resource is fixed
			dp1.Spec.Networking.Address = "192.168.1.1"
			err = zoneStores[0].Update(context.Background(), dp1)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				actual := mesh.DataplaneResourceList{}
				err := globalStore.List(context.Background(), &actual)
				Expect(err).ToNot(HaveOccurred())
				return len(actual.Items)
			}, "5s", "100ms").Should(Equal(1))
		})
	})
})
