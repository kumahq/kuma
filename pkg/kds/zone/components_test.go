package zone_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
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
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
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

type testRuntimeContext struct {
	core_runtime.Runtime
	kds *kds_context.Context
}

func (t *testRuntimeContext) KDSContext() *kds_context.Context {
	return t.kds
}

var _ = Describe("Zone Sync", func() {

	zoneName := "zone-1"

	newPolicySink := func(zoneName string, resourceSyncer sync_store.ResourceSyncer, cs *grpc.MockClientStream, rt core_runtime.Runtime) component.Component {
		return kds_client.NewKDSSink(core.Log, registry.Global().ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kds_client.NewKDSStream(cs, zoneName, ""), zone.Callbacks(rt, resourceSyncer, false, zoneName, nil))
	}
	start := func(comp component.Component, stop chan struct{}) {
		go func() {
			_ = comp.Start(stop)
		}()
	}
	ingressFunc := func(zone string) *mesh_proto.Dataplane {
		return &mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Ingress: &mesh_proto.Dataplane_Networking_Ingress{
					AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{{
						Tags: map[string]string{
							mesh_proto.ServiceTag: "backend",
							mesh_proto.ZoneTag:    fmt.Sprintf("not-%s", zone),
						},
					}},
				},
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
					Port: 1212,
					Tags: map[string]string{
						mesh_proto.ZoneTag:    zone,
						mesh_proto.ServiceTag: "ingress",
					},
				}},
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
		wg.Add(1)

		kdsCtx := kds_context.DefaultContext(manager.NewResourceManager(globalStore), "global")
		serverStream := setup.StartServer(globalStore, wg, "global", registry.Global().ObjectTypes(model.HasKDSFlag(model.ConsumedByZone)), kdsCtx.GlobalProvidedFilter)

		stop := make(chan struct{})
		clientStream := serverStream.ClientStream(stop)

		zoneStore = memory.NewStore()
		zoneSyncer = sync_store.NewResourceSyncer(core.Log, zoneStore)

		start(newPolicySink(zoneName, zoneSyncer, clientStream, &testRuntimeContext{kds: kdsCtx}), stop)
		closeFunc = func() {
			close(stop)
		}
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

		closeFunc()
	})

	It("should sync ingresses", func() {
		// create Ingress for current zone, shouldn't be synced
		err := globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc(zoneName)}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc("another-zone-1")}, store.CreateByKey("dp-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc("another-zone-2")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := zoneStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(2))

		actual := mesh.DataplaneResourceList{}
		err = zoneStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())
		closeFunc()
	})

	It("should have up to date list of consumed types", func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneOverviewType: true,
			mesh.GatewayType:           true, // Gateways are zone-local.
			mesh.ServiceInsightType:    true,
			mesh.ServiceOverviewType:   true,
			sample.TrafficRouteType:    true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualConsumedTypes := registry.Global().ObjectTypes(model.HasScope(model.ScopeMesh), model.TypeFilterFn(func(descriptor model.ResourceTypeDescriptor) bool {
			return !excludeTypes[descriptor.Name]
		}))

		// plus 2 global-scope types
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
		closeFunc()
	})
})
