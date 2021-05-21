package global_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/kumahq/kuma/pkg/test/resources/apis/sample"

	"github.com/kumahq/kuma/pkg/core/resources/registry"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/kds/reconcile"

	"github.com/kumahq/kuma/pkg/kds"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/kds/global"
	sync_store "github.com/kumahq/kuma/pkg/kds/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/grpc"
	kds_setup "github.com/kumahq/kuma/pkg/test/kds/setup"
)

var _ = Describe("Global Sync", func() {

	var remoteStores []store.ResourceStore
	var globalStore store.ResourceStore
	var globalSyncer sync_store.ResourceSyncer
	var closeFunc func()

	BeforeEach(func() {
		const numOfRemotes = 2
		const zoneName = "zone-%d"

		// Start `numOfRemotes` Kuma CP Remote
		serverStreams := []*grpc.MockServerStream{}
		wg := &sync.WaitGroup{}
		remoteStores = []store.ResourceStore{}
		for i := 0; i < numOfRemotes; i++ {
			wg.Add(1)
			remoteStore := memory.NewStore()
			serverStream := kds_setup.StartServer(remoteStore, wg, fmt.Sprintf(zoneName, i), kds.SupportedTypes, reconcile.Any)
			serverStreams = append(serverStreams, serverStream)
			remoteStores = append(remoteStores, remoteStore)
		}

		// Start 1 Kuma CP Global
		globalStore = memory.NewStore()
		globalSyncer = sync_store.NewResourceSyncer(core.Log, globalStore)
		stopCh := make(chan struct{})
		clientStreams := []*grpc.MockClientStream{}
		for _, ss := range serverStreams {
			clientStreams = append(clientStreams, ss.ClientStream(stopCh))
		}
		kds_setup.StartClient(clientStreams, []model.ResourceType{mesh.DataplaneType}, stopCh, global.Callbacks(globalSyncer, false, nil))

		// Create Zone resources for each Kuma CP Remote
		for i := 0; i < numOfRemotes; i++ {
			zone := &system.ZoneResource{Spec: &system_proto.Zone{Enabled: &wrappers.BoolValue{Value: true}}}
			err := globalStore.Create(context.Background(), zone, store.CreateByKey(fmt.Sprintf(zoneName, i), model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
		}

		closeFunc = func() {
			close(stopCh)
			wg.Wait()
		}
	})

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
						Tags: map[string]string{
							mesh_proto.ServiceTag:  "web",
							mesh_proto.ProtocolTag: "http",
						},
					},
				},
			},
		}
	}

	It("should add resource to global store after adding it to remote", func() {
		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-1", fmt.Sprintf("service-1-%d", i))
			err := remoteStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-1-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}
		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(10))

		closeFunc()
	})

	It("should sync resources independently for each Remote", func() {
		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-1", fmt.Sprintf("service-1-%d", i))
			err := remoteStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-1-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}

		for i := 0; i < 10; i++ {
			dp := dataplaneFunc("kuma-cluster-2", fmt.Sprintf("service-2-%d", i))
			err := remoteStores[1].Create(context.Background(), &mesh.DataplaneResource{Spec: dp}, store.CreateByKey(fmt.Sprintf("dp-2-%d", i), "mesh-1"))
			Expect(err).ToNot(HaveOccurred())
		}

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "3s", "100ms").Should(Equal(20))

		err := remoteStores[0].Delete(context.Background(), mesh.NewDataplaneResource(), store.DeleteByKey("dp-1-0", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = remoteStores[0].Delete(context.Background(), mesh.NewDataplaneResource(), store.DeleteByKey("dp-1-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "5s", "100ms").Should(Equal(18))

		closeFunc()
	})

	It("should support same dataplane names through clusters", func() {
		dp1 := dataplaneFunc("kuma-cluster-1", "backend")
		err := remoteStores[0].Create(context.Background(), &mesh.DataplaneResource{Spec: dp1}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		dp2 := dataplaneFunc("kuma-cluster-2", "web")
		err = remoteStores[1].Create(context.Background(), &mesh.DataplaneResource{Spec: dp2}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() int {
			actual := mesh.DataplaneResourceList{}
			err := globalStore.List(context.Background(), &actual)
			Expect(err).ToNot(HaveOccurred())
			return len(actual.Items)
		}, "3s", "100ms").Should(Equal(2))
	})

	It("should have up to date list of provided types", func() {
		excludeTypes := map[model.ResourceType]bool{
			mesh.DataplaneInsightType:  true,
			mesh.DataplaneOverviewType: true,
			mesh.ServiceInsightType:    true,
			mesh.ServiceOverviewType:   true,
			sample.TrafficRouteType:    true,
		}

		// take all mesh-scoped types and exclude types that won't be synced
		actualProvidedTypes := []model.ResourceType{}
		for _, typ := range registry.Global().ListTypes() {
			obj, err := registry.Global().NewObject(typ)
			Expect(err).ToNot(HaveOccurred())
			if obj.Scope() == model.ScopeMesh && !excludeTypes[typ] {
				actualProvidedTypes = append(actualProvidedTypes, typ)
			}
		}

		// plus 2 global-scope types
		extraTypes := []model.ResourceType{
			mesh.MeshType,
			system.ConfigType,
		}

		actualProvidedTypes = append(actualProvidedTypes, extraTypes...)
		Expect(actualProvidedTypes).To(HaveLen(len(global.ProvidedTypes)))
		Expect(actualProvidedTypes).To(ConsistOf(global.ProvidedTypes))
	})

})
