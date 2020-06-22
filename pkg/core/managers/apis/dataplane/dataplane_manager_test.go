package dataplane_test

import (
	"context"

	"github.com/Kong/kuma/pkg/core/managers/apis/dataplane"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Dataplane Manager", func() {

	It("should create a new resource with cluster tag", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "cluster-1")

		// given
		input := mesh_core.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:    3030,
							Address: "10.0.0.1",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "service-1",
							},
						},
					},
				},
			},
		}

		// when
		err := manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())

		actual := mesh_core.DataplaneResource{}
		err = s.Get(context.Background(), &actual, store.GetByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(actual.Spec.Networking.Inbound).To(HaveLen(1))
		Expect(actual.Spec.Networking.Inbound[0].Tags[mesh_proto.ClusterTag]).To(Equal("cluster-1"))
	})

	It("should update a resource with cluster tag", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "cluster-1")

		// given
		input := mesh_core.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:    3030,
							Address: "10.0.0.1",
							Tags: map[string]string{
								mesh_proto.ServiceTag: "service-1",
							},
						},
					},
				},
			},
		}

		err := s.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())

		actual := mesh_core.DataplaneResource{}
		err = s.Get(context.Background(), &actual, store.GetByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(len(actual.Spec.Networking.Inbound[0].Tags)).To(Equal(1))
		_, ok := actual.Spec.Networking.Inbound[0].Tags[mesh_proto.ClusterTag]
		Expect(ok).To(BeFalse())

		// when
		input.Spec.Networking.Address = "10.0.0.2"
		err = manager.Update(context.Background(), &input)
		Expect(err).ToNot(HaveOccurred())

		//then
		actual = mesh_core.DataplaneResource{}
		err = s.Get(context.Background(), &actual, store.GetByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Spec.Networking.Inbound).To(HaveLen(1))
		Expect(actual.Spec.Networking.Inbound[0].Tags[mesh_proto.ClusterTag]).To(Equal("cluster-1"))
	})

})
