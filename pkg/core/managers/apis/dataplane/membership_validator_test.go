package dataplane_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

var _ = Describe("Membership validator", func() {
	type testCase struct {
		mesh *mesh_proto.Mesh
		dp   *mesh_proto.Dataplane
	}

	DescribeTable("should pass validation",
		func(given testCase) {
			// given
			mesh := core_mesh.NewMeshResource()
			Expect(mesh.SetSpec(given.mesh)).To(Succeed())

			dpKey := core_model.ResourceKey{
				Mesh: core_model.DefaultMesh,
				Name: "dp-1",
			}
			dp := core_mesh.NewDataplaneResource()
			Expect(dp.SetSpec(given.dp)).To(Succeed())

			// when
			err := dataplane.NewMembershipValidator().ValidateCreate(context.Background(), dpKey, dp, mesh)

			// then
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("when membership is nil", testCase{
			mesh: &mesh_proto.Mesh{},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}),
		Entry("when membership lists are empty", testCase{
			mesh: &mesh_proto.Mesh{},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
					},
				},
			},
		}),
	)
})
