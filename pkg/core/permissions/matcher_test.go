package permissions

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/test/resources/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Matcher", func() {
	It("should match only proper permissions", func() {
		dataplane := mesh_proto.Dataplane{
			Networking: &mesh_proto.Dataplane_Networking{
				Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
					{
						Interface: "192.168.0.1:8080:80",
						Tags: map[string]string{
							"service": "backend",
						},
					},
					{
						Interface: "192.168.0.1:8090:90",
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
						Name:      "mobile-api-gateway",
						Mesh:      "default",
						Namespace: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Rules: []*mesh_proto.TrafficPermission_Rule{
							{
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "mobile",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "api-gateway",
										},
									},
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name:      "mobile",
						Mesh:      "default",
						Namespace: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Rules: []*mesh_proto.TrafficPermission_Rule{
							{ // relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "mobile",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "backend",
										},
									},
								},
							},
							{ // not relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "mobile",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "stats",
										},
									},
								},
							},
						},
					},
				},
				{
					Meta: &model.ResourceMeta{
						Name:      "load-balancer",
						Mesh:      "default",
						Namespace: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Rules: []*mesh_proto.TrafficPermission_Rule{
							{ // relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "load-balancer",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "web",
										},
									},
								},
							},
							{ // not relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "load-balancer",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "mobile",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// when
		matchedPerms := MatchDataplaneTrafficPermissions(&dataplane, &permissions)

		// then
		backendMatches := matchedPerms.Get("192.168.0.1:8080:80")
		expectedBackendMatches := core_mesh.TrafficPermissionResourceList{
			Items: []*core_mesh.TrafficPermissionResource{
				{
					Meta: &model.ResourceMeta{
						Name:      "mobile",
						Mesh:      "default",
						Namespace: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Rules: []*mesh_proto.TrafficPermission_Rule{
							{ // relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "mobile",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "backend",
										},
									},
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
						Name:      "load-balancer",
						Mesh:      "default",
						Namespace: "default",
					},
					Spec: mesh_proto.TrafficPermission{
						Rules: []*mesh_proto.TrafficPermission_Rule{
							{ // relevant rule
								Sources: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "load-balancer",
										},
									},
								},
								Destinations: []*mesh_proto.TrafficPermission_Rule_Selector{
									{
										Match: map[string]string{
											"service": "web",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		webMatches := matchedPerms.Get("192.168.0.1:8090:90")
		Expect(*webMatches).To(Equal(expectedWebMatches))
	})
})
