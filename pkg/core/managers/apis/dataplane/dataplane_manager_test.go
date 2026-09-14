package dataplane_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
)

var _ = Describe("Dataplane Manager", func() {
	It("should set health.ready to false if serviceProbe is provided and health is nil", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// given
		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:    3030,
							Address: "10.0.0.1",
							ServiceProbe: &mesh_proto.Dataplane_Networking_Inbound_ServiceProbe{
								Tcp: &mesh_proto.Dataplane_Networking_Inbound_ServiceProbe_Tcp{},
							},
						},
					},
				},
			},
		}

		err = manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())

		actual := core_mesh.NewDataplaneResource()
		err = s.Get(context.Background(), actual, store.GetByKey("dp1", "default"))
		Expect(err).ToNot(HaveOccurred())
		Expect(actual.Spec.Networking.Inbound[0].State).To(Equal(mesh_proto.Dataplane_Networking_Inbound_NotReady))
	})
	It("should default the outbound address on create and keep an explicit one", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// given an outbound without an address and one that picks its own
		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Port: 10001,
							BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
								Kind: "MeshService",
								Name: "backend",
								Port: 80,
							},
						},
						{
							Port:    10002,
							Address: "240.0.0.1",
							BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
								Kind: "MeshService",
								Name: "redis",
								Port: 6379,
							},
						},
					},
				},
			},
		}

		// when
		err = manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))

		// then
		Expect(err).ToNot(HaveOccurred())
		actual := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), actual, store.GetByKey("dp1", "default"))).To(Succeed())
		Expect(actual.Spec.Networking.Outbound[0].Address).To(Equal("127.0.0.1"))
		Expect(actual.Spec.Networking.Outbound[1].Address).To(Equal("240.0.0.1"))
	})

	It("should default the outbound address on update", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
				},
			},
		}
		Expect(manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))).To(Succeed())

		// given an outbound added without an address
		actual := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), actual, store.GetByKey("dp1", "default"))).To(Succeed())
		actual.Spec.Networking.Outbound = []*mesh_proto.Dataplane_Networking_Outbound{
			{
				Port: 10001,
				BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
					Kind: "MeshService",
					Name: "backend",
					Port: 80,
				},
			},
		}

		// when
		Expect(manager.Update(context.Background(), actual)).To(Succeed())

		// then
		updated := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), updated, store.GetByKey("dp1", "default"))).To(Succeed())
		Expect(updated.Spec.Networking.Outbound[0].Address).To(Equal("127.0.0.1"))
	})
})
