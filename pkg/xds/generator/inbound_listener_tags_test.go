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
	dataplaneWithLabels := func(labels map[string]string) *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{Name: "backend-01", Mesh: "default", Labels: labels},
			Spec: &mesh_proto.Dataplane{},
		}
	}

	It("keeps the Dataplane labels and adds the protocol", func() {
		// when
		out := generator.InboundListenerTags(
			dataplaneWithLabels(map[string]string{"k8s.kuma.io/namespace": "kuma-demo"}),
			"http",
			"self_inbound_dp_http",
		)

		// then
		Expect(out).To(Equal(map[string]string{
			"k8s.kuma.io/namespace": "kuma-demo",
			mesh_proto.ProtocolTag:  "http",
		}))
	})

	It("does not mutate the Dataplane labels", func() {
		// given
		labels := map[string]string{"k8s.kuma.io/namespace": "kuma-demo"}
		dataplane := dataplaneWithLabels(labels)

		// when
		generator.InboundListenerTags(dataplane, "http", "self_inbound_dp_http")

		// then
		Expect(labels).To(Equal(map[string]string{"k8s.kuma.io/namespace": "kuma-demo"}))
	})

	It("keeps the protocol alone when the Dataplane has no labels", func() {
		// when
		out := generator.InboundListenerTags(dataplaneWithLabels(nil), "grpc", "self_inbound_dp_grpc")

		// then
		Expect(out).To(Equal(map[string]string{mesh_proto.ProtocolTag: "grpc"}))
	})

	It("writes the contextual name under kuma.io/unified-name when there is nothing else", func() {
		// when
		out := generator.InboundListenerTags(dataplaneWithLabels(nil), "", "self_inbound_dp_8080")

		// then
		Expect(out).To(Equal(map[string]string{mesh_proto.UnifiedNameTag: "self_inbound_dp_8080"}))
	})
})
