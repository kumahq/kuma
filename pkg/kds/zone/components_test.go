package zone_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/kds/zone"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/grpc"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/kds/setup"
	"github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var _ = Describe("Zone Sync", func() {

	zoneName := "zone-1"

	newPolicySink := func(zoneName string, resourceSyncer sync_store.ResourceSyncer, cs *grpc.MockClientStream, configs map[string]bool) kds_client.KDSSink {
		return kds_client.NewKDSSink(core.Log.WithName("kds-sink"), registry.Global().ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kds_client.NewKDSStream(cs, zoneName, ""), zone.Callbacks(configs, resourceSyncer, false, zoneName, nil))
	}
	ingressFunc := func(zone string) *mesh_proto.ZoneIngress {
		return &mesh_proto.ZoneIngress{
			Zone: zone,
			Networking: &mesh_proto.ZoneIngress_Networking{
				Address: "192.168.0.1",
				Port:    1212,
			},
			AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
				{
					Tags: map[string]string{
						mesh_proto.ServiceTag: "backend",
						mesh_proto.ZoneTag:    fmt.Sprintf("not-%s", zone),
					},
				},
			},
		}
	}

	var zoneStore store.ResourceStore
	var zoneSyncer sync_store.ResourceSyncer
	var globalStore store.ResourceStore
	var closeFunc func()

	BeforeEach(func() {
		globalStore = memory.NewStore()
		wg := &sync.WaitGroup{}

		kdsCtx := kds_context.DefaultContext(context.Background(), manager.NewResourceManager(globalStore), "global")
		wg.Add(1)
		serverStream := setup.StartServer(globalStore, wg, "global", registry.Global().ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kdsCtx.GlobalProvidedFilter, kdsCtx.GlobalResourceMapper)

		stop := make(chan struct{})
		clientStream := serverStream.ClientStream(stop)

		zoneStore = memory.NewStore()
		zoneSyncer = sync_store.NewResourceSyncer(core.Log.WithName("kds-syncer"), zoneStore)

		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = newPolicySink(zoneName, zoneSyncer, clientStream, kdsCtx.Configs).Receive()
		}()
		closeFunc = func() {
			Expect(clientStream.CloseSend()).To(Succeed())
			close(stop)
			wg.Wait()
		}
	})

	AfterEach(func() {
		closeFunc()
	})

	It("should sync policies from global store to the local", func() {
		err := globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.MeshResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(1))

		actual := mesh.MeshResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual.Items[0].Spec).To(Equal(samples.Mesh1))
	})

	It("should sync ingresses", func() {
		// create Ingress for current zone, shouldn't be synced
		err := globalStore.Create(context.Background(), &mesh.ZoneIngressResource{Spec: ingressFunc(zoneName)}, store.CreateByKey("dp-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.ZoneIngressResource{Spec: ingressFunc("another-zone-1")}, store.CreateByKey("dp-2", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.ZoneIngressResource{Spec: ingressFunc("another-zone-2")}, store.CreateByKey("dp-3", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.ZoneIngressResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(2))

		actual := mesh.ZoneIngressResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should have up to date list of consumed types", func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneOverviewType: true,
			mesh.ServiceOverviewType:   true,
			mesh.DataplaneType:         true,
			sample.TrafficRouteType:    true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualConsumedTypes := registry.Global().ObjectTypes(model.HasScope(model.ScopeMesh), model.TypeFilterFn(func(descriptor model.ResourceTypeDescriptor) bool {
			return !excludeTypes[descriptor.Name]
		}))

		// plus 4 global-scope types
		extraTypes := []model.ResourceType{
			mesh.MeshType,
			mesh.ZoneIngressType,
			system.ConfigType,
			system.GlobalSecretType,
		}

		actualConsumedTypes = append(actualConsumedTypes, extraTypes...)
		Expect(actualConsumedTypes).To(ConsistOf(registry.Global().ObjectTypes(model.HasKDSFlag(model.ConsumedByZone))))
	})

	It("should not delete predefined ConfigMaps in the Zone cluster", func() {
		// create kuma-cluster-id ConfigMap in Global
		err := globalStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "cluster-id"}},
			store.CreateByKey(config_manager.ClusterIdConfigKey, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// create kuma-cp-leader ConfigMap in Zone
		err = zoneStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "leader"}},
			store.CreateByKey("kuma-cp-leader", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// create kuma-control-plane-config ConfigMap in Zone
		err = zoneStore.Create(context.Background(), &system.ConfigResource{Spec: &v1alpha1.Config{Config: "kuma-cp config"}},
			store.CreateByKey("kuma-control-plane-config", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := system.ConfigResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(3))

		actual := system.ConfigResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())

		actualNames := []string{}
		for _, a := range actual.Items {
			actualNames = append(actualNames, a.GetMeta().GetName())
		}
		expectedNames := []string{
			"kuma-cp-leader",
			"kuma-control-plane-config",
			"kuma-cluster-id",
		}
		Expect(actualNames).To(ConsistOf(expectedNames))
	})
})
