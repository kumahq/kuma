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
			Spec: map[string]interface{}{
				"networking": map[string]interface{}{
					"address": "10.20.1.2",
					"inbound": []map[string]interface{}{
						{
							"tags": map[string]string{mesh_proto.ServiceTag: "ingress", mesh_proto.ZoneTag: "zone-2"},
							"port": 10001,
						},
					},
					"ingress": map[string]interface{}{
						"publicAddress": "192.168.0.100",
						"publicPort":    12345,
						"availableServices": []map[string]interface{}{
							{"instances": 2, "mesh": "mesh-1", "tags": map[string]string{mesh_proto.ServiceTag: "redis", "version": "v2"}},
							{"instances": 3, "mesh": "mesh-1", "tags": map[string]string{mesh_proto.ServiceTag: "redis", "version": "v3"}},
							{"instances": 3, "mesh": "mesh-1", "tags": map[string]string{mesh_proto.ServiceTag: "backend", "version": "v3"}},
							{"instances": 3, "mesh": "mesh-2", "tags": map[string]string{mesh_proto.ServiceTag: "db", "version": "v3"}},
							{"instances": 3, "mesh": "mesh-2", "tags": map[string]string{mesh_proto.ServiceTag: "web", "version": "v3"}},
							{"instances": 3, "mesh": "mesh-3", "tags": map[string]string{mesh_proto.ServiceTag: "frontend", "version": "v3"}},
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
