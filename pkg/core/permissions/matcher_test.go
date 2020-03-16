package permissions

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/test/resources/model"
)

var _ = Describe("Matcher", func() {
	It("should match only proper permissions", func() {
		dataplane := mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Address: "192.168.0.1",
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Port:        8080,
						ServicePort: 80,
						Tags: map[string]string{
							"service": "backend",
						},
					},
					{
						Port:        8090,
						ServicePort: 90,
						Tags: map[string]string{
							"service": "web",
						},
					},
				},
			},
		}

		permissions := core_mesh.TrafficPermissionResourceList{
			Items: []*core_mesh.TrafficPermissionResource{
				{ // not relevant resource
					Meta: &model.ResourceMeta{
						Name: "mobile-api-gateway",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "mobile",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "api-gateway",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "mobile",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{ // relevant rule
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "mobile",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "backend",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "mobile-2",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{ // not relevant rule
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "mobile",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "stats",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "load-balancer",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{ // relevant rule
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "load-balancer",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web",
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name: "load-balancer-2",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{ // not relevant rule
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "load-balancer",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "mobile",
								},
							},
						},
					},
				},
			},
		}

		// when
		matchedPerms, err := MatchDataplaneTrafficPermissions(&dataplane, &permissions)

		// then
		Expect(err).ToNot(HaveOccurred())
		backendMatches := matchedPerms.Get(mesh_proto.InboundInterface{
			DataplaneIP:   "192.168.0.1",
			DataplanePort: 8080,
			WorkloadPort:  80,
		})
		expectedBackendMatches := core_mesh.TrafficPermissionResourceList{
			Items: []*core_mesh.TrafficPermissionResource{
				{
					Meta: &model.ResourceMeta{
						Name: "mobile",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "mobile",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "backend",
								},
							},
						},
					},
				},
			},
		}
		Expect(*backendMatches).To(Equal(expectedBackendMatches))

		// and
		expectedWebMatches := core_mesh.TrafficPermissionResourceList{
			Items: []*core_mesh.TrafficPermissionResource{
				{
					Meta: &model.ResourceMeta{
						Name: "load-balancer",
						Mesh: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "load-balancer",
								},
							},
						},
						Destinations: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"service": "web",
								},
							},
						},
					},
				},
			},
		}
		webMatches := matchedPerms.Get(mesh_proto.InboundInterface{
			DataplaneIP:   "192.168.0.1",
			DataplanePort: 8090,
			WorkloadPort:  90,
		})
		Expect(*webMatches).To(Equal(expectedWebMatches))
	})
})
