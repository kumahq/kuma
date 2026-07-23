package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("InboundListenerTags", func() {
	It("keeps existing inbound tags untouched", func() {
		// given
		tags := map[string]string{mesh_proto.ServiceTag: "backend", "version": "v1"}

		// when
		out := generator.InboundListenerTags(tags, "self_inbound_dp_http")

		// then
		Expect(out).To(Equal(tags))
	})

	It("writes the contextual name under kuma.io/unified-name when tags are empty", func() {
		// when
		out := generator.InboundListenerTags(map[string]string{}, "self_inbound_dp_http")

		// then
		Expect(out).To(Equal(map[string]string{
			mesh_proto.UnifiedNameTag: "self_inbound_dp_http",
		}))
	})

	It("writes the contextual name when tags are nil", func() {
		// when
		out := generator.InboundListenerTags(nil, "self_inbound_dp_8080")

		// then
		Expect(out).To(Equal(map[string]string{
			mesh_proto.UnifiedNameTag: "self_inbound_dp_8080",
		}))
	})
})
