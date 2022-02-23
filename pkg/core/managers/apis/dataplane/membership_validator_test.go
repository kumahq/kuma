package dataplane_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/managers/apis/dataplane"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var _ = Describe("Membership validator", func() {

	type testCase struct {
		mesh *mesh_proto.Mesh
		dp   *mesh_proto.Dataplane
		err  error
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
		Entry("when dp fulfills allowed mesh tag requirements", testCase{
			mesh: &mesh_proto.Mesh{
				Constraints: &mesh_proto.Mesh_Constraints{
					DataplaneProxy: &mesh_proto.Mesh_DataplaneProxyConstraints{
						Requirements: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
									"version":         "*",
								},
							},
							{
								Tags: map[string]string{
									"kuma.io/service": "backend-api",
									"version":         "*",
								},
							},
						},
						Restrictions: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "web",
								},
							},
						},
					},
				},
			},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
								"kuma.io/zone":    "east",
								"version":         "v1",
							},
						},
						{
							Tags: map[string]string{
								"kuma.io/service": "backend-api",
								"kuma.io/zone":    "east",
								"version":         "v2",
							},
						},
					},
				},
			},
		}),
	)

	DescribeTable("should fail validation",
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
			Expect(err).To(MatchError(given.err))
		},
		Entry("when dp fails to fulfill requirements", testCase{
			mesh: &mesh_proto.Mesh{
				Constraints: &mesh_proto.Mesh_Constraints{
					DataplaneProxy: &mesh_proto.Mesh_DataplaneProxyConstraints{
						Requirements: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
					},
				},
			},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
							},
						},
						{
							Tags: map[string]string{
								"kuma.io/service": "backend-api",
							},
						},
					},
				},
			},
			err: &dataplane.NotAllowedErr{
				Mesh: core_model.DefaultMesh,
				TagSet: map[string]string{
					"kuma.io/service": "backend-api",
				},
			},
		}),
		Entry("when dp uses tag value that is denied", testCase{
			mesh: &mesh_proto.Mesh{
				Constraints: &mesh_proto.Mesh_Constraints{
					DataplaneProxy: &mesh_proto.Mesh_DataplaneProxyConstraints{
						Requirements: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
						Restrictions: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/zone": "east",
								},
							},
						},
					},
				},
			},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
								"kuma.io/zone":    "east",
							},
						},
					},
				},
			},
			err: &dataplane.DeniedErr{
				Mesh: core_model.DefaultMesh,
				DpTagSet: map[string]string{
					"kuma.io/service": "backend",
					"kuma.io/zone":    "east",
				},
				DeniedTagSet: map[string]string{
					"kuma.io/zone": "east",
				},
			},
		}),
		Entry("when dp do not provide required tag", testCase{
			mesh: &mesh_proto.Mesh{
				Constraints: &mesh_proto.Mesh_Constraints{
					DataplaneProxy: &mesh_proto.Mesh_DataplaneProxyConstraints{
						Requirements: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
									"team":            "*",
								},
							},
						},
					},
				},
			},
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
			err: &dataplane.NotAllowedErr{
				Mesh: core_model.DefaultMesh,
				TagSet: map[string]string{
					"kuma.io/service": "backend",
				},
			},
		}),
		Entry("when dp uses tag key that is denied", testCase{
			mesh: &mesh_proto.Mesh{
				Constraints: &mesh_proto.Mesh_Constraints{
					DataplaneProxy: &mesh_proto.Mesh_DataplaneProxyConstraints{
						Requirements: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"kuma.io/service": "backend",
								},
							},
						},
						Restrictions: []*mesh_proto.Mesh_DataplaneProxyConstraints_Rules{
							{
								Tags: map[string]string{
									"version": "*",
								},
							},
						},
					},
				},
			},
			dp: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{
								"kuma.io/service": "backend",
								"version":         "v1",
							},
						},
					},
				},
			},
			err: &dataplane.DeniedErr{
				Mesh: core_model.DefaultMesh,
				DpTagSet: map[string]string{
					"kuma.io/service": "backend",
					"version":         "v1",
				},
				DeniedTagSet: map[string]string{
					"version": "*",
				},
			},
		}),
	)
})
