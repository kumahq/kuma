package ingress_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/kds/util"
	model "github.com/Kong/kuma/pkg/test/resources/model"
	"github.com/Kong/kuma/pkg/xds/ingress"
)

var _ = Describe("SplitIngressesByMeshAndFlatten", func() {

	availableService := func(service, mesh string) *mesh_proto.Dataplane_Networking_Ingress_AvailableService {
		return &mesh_proto.Dataplane_Networking_Ingress_AvailableService{
			Tags: map[string]string{
				"service": service,
				"mesh":    mesh,
				"zone":    "z1",
			},
			Instances: 1,
		}
	}

	It("should replace multi-mesh Ingress with multiple Ingresses for single mesh", func() {
		// given
		ingressRes := &mesh.DataplaneResource{
			Meta: &model.ResourceMeta{Mesh: "kuma-system", Name: "ingress"},
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							availableService("backend", "mesh1"),
							availableService("web", "mesh2"),
							availableService("app", "mesh2"),
							availableService("redis", "mesh3"),
						},
					},
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
						Port: 1212,
						Tags: map[string]string{
							mesh_proto.ZoneTag:    "z1",
							mesh_proto.ServiceTag: "ingress",
						},
					}},
				},
			},
		}
		list := &mesh.DataplaneResourceList{Items: []*mesh.DataplaneResource{ingressRes}}
		actual := ingress.SplitIngressesByMeshAndFlatten(list)

		expected := map[string]*mesh.DataplaneResource{
			"mesh1": {
				Meta: util.ResourceKeyToMeta("ingress", "mesh1"),
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Ingress: &mesh_proto.Dataplane_Networking_Ingress{
							AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
								availableService("backend", "mesh1"),
							},
						},
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
							Port: 1212,
							Tags: map[string]string{
								mesh_proto.ZoneTag:    "z1",
								mesh_proto.ServiceTag: "ingress",
							},
						}},
					},
				},
			},
			"mesh2": {
				Meta: util.ResourceKeyToMeta("ingress", "mesh2"),
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Ingress: &mesh_proto.Dataplane_Networking_Ingress{
							AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
								availableService("web", "mesh2"),
								availableService("app", "mesh2"),
							},
						},
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
							Port: 1212,
							Tags: map[string]string{
								mesh_proto.ZoneTag:    "z1",
								mesh_proto.ServiceTag: "ingress",
							},
						}},
					},
				},
			},
			"mesh3": {
				Meta: util.ResourceKeyToMeta("ingress", "mesh3"),
				Spec: mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Ingress: &mesh_proto.Dataplane_Networking_Ingress{
							AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
								availableService("redis", "mesh3"),
							},
						},
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
							Port: 1212,
							Tags: map[string]string{
								mesh_proto.ZoneTag:    "z1",
								mesh_proto.ServiceTag: "ingress",
							},
						}},
					},
				},
			},
		}

		Expect(actual.Items).To(HaveLen(3))
		actualMap := map[string]*mesh.DataplaneResource{}
		for _, item := range actual.Items {
			actualMap[item.GetMeta().GetMesh()] = item
		}
		Expect(actualMap).To(Equal(expected))
	})
})
