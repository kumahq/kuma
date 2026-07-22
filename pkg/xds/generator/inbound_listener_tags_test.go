package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("InboundListenerTags", func() {
	dpWithLabels := func() *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Name: "backend-7f9c",
				Mesh: "default",
				Labels: map[string]string{
					mesh_proto.ZoneTag:          "east",
					mesh_proto.KubeNamespaceTag: "kuma-demo",
					mesh_proto.DisplayName:      "backend-7f9c",
				},
			},
			Spec: &mesh_proto.Dataplane{},
		}
	}

	It("keeps existing inbound tags untouched", func() {
		// given
		tags := map[string]string{mesh_proto.ServiceTag: "backend", "version": "v1"}

		// when
		out := generator.InboundListenerTags(dpWithLabels(), tags, "http")

		// then
		Expect(out).To(Equal(tags))
	})

	It("synthesizes a kuma.io/kri tag with the port as section name when tags are empty", func() {
		// when
		out := generator.InboundListenerTags(dpWithLabels(), map[string]string{}, "http")

		// then
		Expect(out).To(Equal(map[string]string{
			mesh_proto.KRITag: "kri_dp_default_east_kuma-demo_backend-7f9c_http",
		}))
	})

	It("returns the empty tags unchanged when the Dataplane KRI cannot be built", func() {
		// given a Dataplane with no meta, so no KRI can be synthesized
		dp := &core_mesh.DataplaneResource{Spec: &mesh_proto.Dataplane{}}
		tags := map[string]string{}

		// when
		out := generator.InboundListenerTags(dp, tags, "http")

		// then
		Expect(out).To(BeEmpty())
	})
})
