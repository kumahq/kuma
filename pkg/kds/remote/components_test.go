package remote_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/apis/sample"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/registry"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	kds_context "github.com/kumahq/kuma/pkg/kds/context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	kds_client "github.com/kumahq/kuma/pkg/kds/client"
	"github.com/kumahq/kuma/pkg/kds/remote"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/grpc"
	"github.com/kumahq/kuma/pkg/test/kds/samples"
	"github.com/kumahq/kuma/pkg/test/kds/setup"
)

var _ = Describe("Remote Sync", func() {

	remoteZone := "zone-1"

	consumedTypes := []model.ResourceType{mesh.DataplaneType, mesh.MeshType, mesh.TrafficPermissionType}
	newPolicySink := func(zone string, resourceSyncer sync_store.ResourceSyncer, cs *grpc.MockClientStream) component.Component {
		return kds_client.NewKDSSink(core.Log, consumedTypes, kds_client.NewKDSStream(cs, remoteZone), remote.Callbacks(nil, resourceSyncer, false, zone, nil))
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

	var remoteStore store.ResourceStore
	var remoteSyncer sync_store.ResourceSyncer
	var globalStore store.ResourceStore
	var closeFunc func()

	BeforeEach(func() {
		globalStore = memory.NewStore()
		wg := &sync.WaitGroup{}
		wg.Add(1)
		serverStream := setup.StartServer(globalStore, wg, "global", consumedTypes, kds_context.GlobalProvidedFilter(manager.NewResourceManager(globalStore)))

		stop := make(chan struct{})
		clientStream := serverStream.ClientStream(stop)

		remoteStore = memory.NewStore()
		remoteSyncer = sync_store.NewResourceSyncer(core.Log, remoteStore)

		start(newPolicySink(remoteZone, remoteSyncer, clientStream), stop)
		closeFunc = func() {
			close(stop)
		}
	})

	It("should sync policies from global store to the local", func() {
		err := globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.MeshResourceList{}
			err := remoteStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(1))

		actual := mesh.MeshResourceList{}
		err = remoteStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual.Items[0].Spec).To(Equal(samples.Mesh1))

		closeFunc()
	})

	It("should sync ingresses", func() {
		// create Ingress for current remote zone, shouldn't be synced
		err := globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc(remoteZone)}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc("another-zone-1")}, store.CreateByKey("dp-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())
		err = globalStore.Create(context.Background(), &mesh.DataplaneResource{Spec: ingressFunc("another-zone-2")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := remoteStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(2))

		actual := mesh.DataplaneResourceList{}
		err = remoteStore.List(context.Background(), &actual)
		Expect(err).ToNot(HaveOccurred())
		closeFunc()
	})

	It("should have up to date list of consumed types", func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneOverviewType: true,
			mesh.ServiceInsightType:    true,
			mesh.ServiceOverviewType:   true,
			sample.TrafficRouteType:    true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualConsumedTypes := []model.ResourceType{}
		for _, typ := range registry.Global().ListTypes() {
			obj, err := registry.Global().NewObject(typ)
			Expect(err).ToNot(HaveOccurred())
			if obj.Scope() == model.ScopeMesh && !excludeTypes[typ] {
				actualConsumedTypes = append(actualConsumedTypes, typ)
			}
		}

		// plus 2 global-scope types
		extraTypes := []model.ResourceType{
			mesh.MeshType,
			system.ConfigType,
		}

		actualConsumedTypes = append(actualConsumedTypes, extraTypes...)
		Expect(actualConsumedTypes).To(HaveLen(len(remote.ConsumedTypes)))
		Expect(actualConsumedTypes).To(ConsistOf(remote.ConsumedTypes))
	})
})
