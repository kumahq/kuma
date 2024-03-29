package ingress

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
)

var _ = Describe("Ingress BuildDestinationMap", func() {
	It("should generate destination map by ingress", func() {
		ingress := &mesh.ZoneIngressResource{
			Spec: &mesh_proto.ZoneIngress{
				AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
					{
						Mesh: "mesh-1",
						Tags: map[string]string{"kuma.io/service": "backend", "version": "v1", "region": "us"},
					},
					{
						Mesh: "mesh-1",
						Tags: map[string]string{"kuma.io/service": "backend"},
					},
					{
						Mesh: "mesh-1",
						Tags: map[string]string{"kuma.io/service": "web", "version": "v2", "region": "eu"},
					},
					{
						Mesh: "mesh-2",
						Tags: map[string]string{"kuma.io/service": "payments"},
					},
				},
			},
		}

		actual := BuildDestinationMap("mesh-1", ingress)
		expected := xds.DestinationMap{
			"backend": []mesh_proto.TagSelector{
				{
					"region":          "us",
					"version":         "v1",
					"kuma.io/service": "backend",
				},
				{
					"kuma.io/service": "backend",
				},
			},
			"web": []mesh_proto.TagSelector{
				{
					"region":          "eu",
					"kuma.io/service": "web",
					"version":         "v2",
				},
			},
		}
		Expect(actual).To(Equal(expected))
	})
})
