package topology_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	plugins_memory "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("GetTrafficTrace", func() {

	dataplane := core_mesh.DataplaneResource{
		Meta: &test_model.ResourceMeta{
			Name: "dp1",
			Mesh: "default",
		},
		Spec: &mesh_proto.Dataplane{
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

	It("should return matched TrafficTrace", func() {
		// given
		store := plugins_memory.NewStore()
		manager := resources_manager.NewResourceManager(store)

		trafficTrace1 := core_mesh.TrafficTraceResource{
			Meta: &test_model.ResourceMeta{
				Name: "tt1",
				Mesh: "default",
			},
			Spec: &mesh_proto.TrafficTrace{
				Selectors: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "web",
						},
					},
				},
			},
		}

		trafficTrace2 := core_mesh.TrafficTraceResource{
			Meta: &test_model.ResourceMeta{
				Name: "tt2",
				Mesh: "default",
			},
			Spec: &mesh_proto.TrafficTrace{
				Selectors: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
			},
		}

		err := store.Create(context.Background(), &trafficTrace1, core_store.CreateBy(core_model.MetaToResourceKey(trafficTrace1.GetMeta())))
		Expect(err).ToNot(HaveOccurred())

		err = store.Create(context.Background(), &trafficTrace2, core_store.CreateBy(core_model.MetaToResourceKey(trafficTrace2.GetMeta())))
		Expect(err).ToNot(HaveOccurred())

		// when
		picked, err := topology.GetTrafficTrace(context.Background(), &dataplane, manager)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(picked).To(Equal(&trafficTrace2))
	})

	It("should return nil when there are no matching traffic traces", func() {
		// given
		store := plugins_memory.NewStore()
		manager := resources_manager.NewResourceManager(store)

		// when
		picked, err := topology.GetTrafficTrace(context.Background(), &dataplane, manager)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(picked).To(BeNil())
	})
})
