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
	It("should reject a create that marks a dataplane with inbounds as a gateway", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// given a gateway label that only the create options carry
		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{Port: 8080}},
				},
			},
		}

		// when
		err = manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"),
			store.CreateWithLabels(map[string]string{mesh_proto.GatewayLabel: mesh_proto.GatewayEnabled}))

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("inbound cannot be defined for delegated gateways"))
	})

	It("should reject an update that marks a dataplane with inbounds as a gateway", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{Port: 8080}},
				},
			},
		}
		Expect(manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"))).To(Succeed())

		actual := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), actual, store.GetByKey("dp1", "default"))).To(Succeed())

		// when the label is added while the inbounds stay
		err = manager.Update(context.Background(), actual,
			store.UpdateWithLabels(map[string]string{mesh_proto.GatewayLabel: mesh_proto.GatewayEnabled}))

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("inbound cannot be defined for delegated gateways"))
	})

	It("should accept an update that unmarks a gateway and adds inbounds", func() {
		// setup
		s := memory.NewStore()
		manager := dataplane.NewDataplaneManager(s, "zone-1", config_core.Zone, false, "", dataplane.NewMembershipValidator())
		err := s.Create(context.Background(), core_mesh.NewMeshResource(), store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		input := core_mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{Address: "10.0.0.1"},
			},
		}
		Expect(manager.Create(context.Background(), &input, store.CreateByKey("dp1", "default"),
			store.CreateWithLabels(map[string]string{mesh_proto.GatewayLabel: mesh_proto.GatewayEnabled}))).To(Succeed())

		actual := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), actual, store.GetByKey("dp1", "default"))).To(Succeed())
		Expect(actual.IsDelegatedGateway()).To(BeTrue())

		// when the label is dropped and inbounds are added in the same update
		actual.Spec.Networking.Inbound = []*mesh_proto.Dataplane_Networking_Inbound{{Port: 8080}}
		err = manager.Update(context.Background(), actual, store.UpdateWithLabels(map[string]string{}))

		// then
		Expect(err).ToNot(HaveOccurred())

		stored := core_mesh.NewDataplaneResource()
		Expect(s.Get(context.Background(), stored, store.GetByKey("dp1", "default"))).To(Succeed())
		Expect(stored.IsDelegatedGateway()).To(BeFalse())
	})

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
})
