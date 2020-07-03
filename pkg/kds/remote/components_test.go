package remote_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime/component"
	kds_client "github.com/Kong/kuma/pkg/kds/client"
	"github.com/Kong/kuma/pkg/kds/global"
	"github.com/Kong/kuma/pkg/kds/remote"
	sync_store "github.com/Kong/kuma/pkg/kds/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/test/grpc"
	"github.com/Kong/kuma/pkg/test/kds/samples"
	"github.com/Kong/kuma/pkg/test/kds/setup"
)

var _ = Describe("Remote Sync", func() {

	remoteZone := "cluster-1"

	consumedTypes := []model.ResourceType{mesh.DataplaneType, mesh.MeshType, mesh.TrafficPermissionType}
	newPolicySink := func(zone string, resourceSyncer sync_store.ResourceSyncer, cs *grpc.MockClientStream) component.Component {
		return kds_client.NewKDSSink(core.Log, zone, consumedTypes, func() (client kds_client.KDSClient, err error) {
			return setup.NewMockKDSClient(kds_client.NewKDSStream(cs, remoteZone)), nil
		}, remote.Callbacks(resourceSyncer, false, zone))
	}
	start := func(comp component.Component, stop chan struct{}) {
		go func() {
			err := comp.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
	}
	ingressFunc := func(zone string) mesh_proto.Dataplane {
		return mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Ingress: &mesh_proto.Dataplane_Networking_Ingress{
					AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{{
						Tags: map[string]string{
							"service": "backend",
							"zone":    fmt.Sprintf("not-%s", zone),
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
		serverStream := setup.StartServer(globalStore, wg, "global", consumedTypes, global.ProvidedFilter)

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
		err := globalStore.Create(context.Background(), &mesh.MeshResource{Spec: samples.Mesh1}, store.CreateByKey("mesh-1", "mesh-1"))
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
		for _, item := range actual.Items {
			fmt.Println(item)
		}
		closeFunc()
	})
})
