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
			Spec: map[string]interface{}{
				"networking": map[string]interface{}{
					"address": "10.20.1.2",
					"inbound": []map[string]interface{}{
						{
							"tags": map[string]string{mesh_proto.ServiceTag: "redis"},
							"port": 10001,
						},
					},
				},
			},
		})
		requestsStr := []string{}
		for _, r := range requests {
			requestsStr = append(requestsStr, r.Name)
		}
		Expect(requestsStr).To(HaveLen(1))
		Expect(requestsStr).To(ConsistOf("kuma-mesh-1-dns-vips"))
	})
})
