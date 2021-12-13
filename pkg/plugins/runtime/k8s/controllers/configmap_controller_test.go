package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
	mesh_k8s "github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
)

var _ = Describe("DataplaneToMeshMapper", func() {
	It("should map ingress to list of meshes", func() {
		l := log.NewLogger(log.InfoLevel)
		mapper := DataplaneToMeshMapper(l, "ns", k8s.NewSimpleConverter())
		requests := mapper(&mesh_k8s.Dataplane{
			Mesh: "mesh-1",
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "10.20.1.2",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 10001,
							Tags: map[string]string{mesh_proto.ServiceTag: "redis"},
						},
					},
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						PublicAddress: "192.168.0.100",
						PublicPort:    12345,
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							{
								Instances: 2, Mesh: "mesh-1", Tags: map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2"},
							},
							{
								Instances: 3, Mesh: "mesh-1", Tags: map[string]string{mesh_proto.ServiceTag: "backend", "version": "v3"},
							},
							{
								Instances: 3, Mesh: "mesh-2", Tags: map[string]string{mesh_proto.ServiceTag: "db", "version": "v3"},
							},
							{
								Instances: 3, Mesh: "mesh-2", Tags: map[string]string{mesh_proto.ServiceTag: "web", "version": "v3"},
							},
							{
								Instances: 3, Mesh: "mesh-3", Tags: map[string]string{mesh_proto.ServiceTag: "frontend", "version": "v3"},
							},
						},
					},
				},
			},
		})
		requestsStr := []string{}
		for _, r := range requests {
			requestsStr = append(requestsStr, r.Name)
		}
		Expect(requestsStr).To(HaveLen(3))
		Expect(requestsStr).To(ConsistOf("kuma-mesh-3-dns-vips", "kuma-mesh-2-dns-vips", "kuma-mesh-1-dns-vips"))
	})
})
