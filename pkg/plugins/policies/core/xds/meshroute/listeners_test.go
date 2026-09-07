package meshroute_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
)

var _ = Describe("OutboundListenerTags", func() {
	It("returns a synthesized kuma.io/unified-name tag for a real-resource outbound", func() {
		// given
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			AddIntPortWithName(8080, 8080, core_meta.ProtocolHTTP, "http").
			Build()
		id := kri.WithSectionName(kri.From(ms), "http")
		ds := meshroute.DestinationService{Outbound: &xds_types.Outbound{Resource: id}}

		// when
		tags := ds.OutboundListenerTags()

		// then
		Expect(tags).To(BeEquivalentTo(map[string]string{mesh_proto.UnifiedNameTag: id.String()}))
	})

	It("returns nil when Outbound is nil", func() {
		// given
		ds := meshroute.DestinationService{Outbound: nil}

		// when
		tags := ds.OutboundListenerTags()

		// then
		Expect(tags).To(BeNil())
	})
})

var _ = Describe("DestinationPortFromRef", func() {
	It("returns the only port when a real backendRef has no section name", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			WithLabels(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			Build()

		meshCtx := xds_builders.Context().
			WithMeshLocalResources([]core_model.Resource{ms}).
			Build().
			Mesh

		dest, port, ok := meshroute.DestinationPortFromRef(meshCtx, &resolve.RealResourceBackendRef{
			Resource: kri.From(ms),
		})

		Expect(ok).To(BeTrue())
		Expect(dest).To(Equal(ms))
		Expect(port.GetValue()).To(Equal(int32(8080)))
	})

	It("does not guess a port when a real backendRef has multiple ports and no section name", func() {
		ms := builders.MeshService().
			WithName("backend").
			WithMesh("default").
			WithLabels(map[string]string{
				mesh_proto.DisplayName:      "backend",
				mesh_proto.KubeNamespaceTag: "kuma-demo",
			}).
			AddIntPort(8080, 8080, core_meta.ProtocolHTTP).
			AddIntPort(9090, 9090, core_meta.ProtocolHTTP).
			Build()

		meshCtx := xds_builders.Context().
			WithMeshLocalResources([]core_model.Resource{ms}).
			Build().
			Mesh

		_, _, ok := meshroute.DestinationPortFromRef(meshCtx, &resolve.RealResourceBackendRef{
			Resource: kri.From(ms),
		})

		Expect(ok).To(BeFalse())
	})
})
