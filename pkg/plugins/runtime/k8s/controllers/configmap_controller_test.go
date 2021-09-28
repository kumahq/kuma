package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

var _ = Describe("DataplaneToMeshMapper", func() {
	It("should map ingress to list of meshes", func() {
		mapper := &DataplaneToMeshMapper{
			ResourceConverter: k8s.NewSimpleConverter(),
		}
		obj, err := mapper.ResourceConverter.ToKubernetesObject(&mesh.DataplaneResource{
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.20.1.2",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Tags: map[string]string{mesh_proto.ServiceTag: "ingress", mesh_proto.ZoneTag: "zone-2"},
							Port: 10001,
						},
					},
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						PublicAddress: "192.168.0.100",
						PublicPort:    12345,
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							{
								Instances: 2,
								Mesh:      "mesh-1",
								Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2"},
							},
							{
								Instances: 3,
								Mesh:      "mesh-1",
								Tags:      map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"},
							},
							{
								Instances: 3,
								Mesh:      "mesh-1",
								Tags:      map[string]string{mesh_proto.ServiceTag: "backend", "version": "v3"},
							},
							{
								Instances: 3,
								Mesh:      "mesh-2",
								Tags:      map[string]string{mesh_proto.ServiceTag: "db", "version": "v3"},
							},
							{
								Instances: 3,
								Mesh:      "mesh-2",
								Tags:      map[string]string{mesh_proto.ServiceTag: "web", "version": "v3"},
							},
							{
								Instances: 3,
								Mesh:      "mesh-3",
								Tags:      map[string]string{mesh_proto.ServiceTag: "frontend", "version": "v3"},
							},
						},
					},
				},
			},
		})
		Expect(err).ToNot(HaveOccurred())
		requests := mapper.Map(handler.MapObject{Object: obj})
		requestsStr := []string{}
		for _, r := range requests {
			requestsStr = append(requestsStr, r.Name)
		}
		Expect(requestsStr).To(HaveLen(3))
		Expect(requestsStr).To(ConsistOf("kuma-mesh-3-dns-vips", "kuma-mesh-2-dns-vips", "kuma-mesh-1-dns-vips"))
	})
})
