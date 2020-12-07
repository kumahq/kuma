package global_test

import (
	"context"
	"fmt"
	"sync"

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
		serverStreams := []*grpc.MockServerStream{}
		wg := &sync.WaitGroup{}
		remoteStores = []store.ResourceStore{}
		for i := 0; i < 2; i++ {
			wg.Add(1)
			remoteStore := memory.NewStore()
			serverStream := kds_setup.StartServer(remoteStore, wg, fmt.Sprintf("cluster-%d", i), kds.SupportedTypes, reconcile.Any)
			serverStreams = append(serverStreams, serverStream)
			remoteStores = append(remoteStores, remoteStore)
		}

		globalStore = memory.NewStore()
		globalSyncer = sync_store.NewResourceSyncer(core.Log, globalStore)
		stopCh := make(chan struct{})
		clientStreams := []*grpc.MockClientStream{}
		for _, ss := range serverStreams {
			clientStreams = append(clientStreams, ss.ClientStream(stopCh))
		}
		kds_setup.StartClient(clientStreams, []model.ResourceType{mesh.DataplaneType}, stopCh, global.Callbacks(globalSyncer, false, nil))

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

})
