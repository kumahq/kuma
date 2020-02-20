package topology_test

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	plugins_memory "github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_model "github.com/Kong/kuma/pkg/test/resources/model"
	"github.com/Kong/kuma/pkg/xds/topology"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetTrafficLog", func() {

	dataplane := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Name: "dp1",
			Mesh: "default",
		},
		Spec: mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port: 8080,
						Tags: map[string]string{
							"service": "backend",
						},
					},
				},
			},
		},
	}

	It("should return matched TrafficLog", func() {
		// given
		store := plugins_memory.NewStore()
		manager := resources_manager.NewResourceManager(store)

		trafficLog1 := core_mesh.TrafficLogResource{
			Meta: &test_model.ResourceMeta{
				Name: "tl1",
				Mesh: "default",
			},
			Spec: mesh_proto.TrafficLog{
				Selectors: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "web",
						},
					},
				},
			},
		}

		trafficLog2 := core_mesh.TrafficLogResource{
			Meta: &test_model.ResourceMeta{
				Name: "tl2",
				Mesh: "default",
			},
			Spec: mesh_proto.TrafficLog{
				Selectors: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
			},
		}

		err := store.Create(context.Background(), &trafficLog1, core_store.CreateBy(core_model.MetaToResourceKey(trafficLog1.GetMeta())))
		Expect(err).ToNot(HaveOccurred())

		err = store.Create(context.Background(), &trafficLog2, core_store.CreateBy(core_model.MetaToResourceKey(trafficLog2.GetMeta())))
		Expect(err).ToNot(HaveOccurred())

		// when
		picked, err := topology.GetTrafficLog(context.Background(), &dataplane, manager)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(picked).To(Equal(&trafficLog2))
	})

	It("should return nil when there are no matching traffic logs", func() {
		// given
		store := plugins_memory.NewStore()
		manager := resources_manager.NewResourceManager(store)

		// when
		picked, err := topology.GetTrafficLog(context.Background(), &dataplane, manager)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(picked).To(BeNil())
	})
})
