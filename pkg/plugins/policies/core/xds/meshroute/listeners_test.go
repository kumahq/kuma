package meshroute_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
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

	It("returns the real outbound tags without kuma.io/mesh for a legacy outbound", func() {
		// given
		ds := meshroute.DestinationService{Outbound: &xds_types.Outbound{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Tags: map[string]string{
					mesh_proto.ServiceTag: "backend",
					mesh_proto.MeshTag:    "default",
					"version":             "v1",
				},
			},
		}}

		// when
		tags := ds.OutboundListenerTags()

		// then
		Expect(tags).To(BeEquivalentTo(map[string]string{
			mesh_proto.ServiceTag: "backend",
			"version":             "v1",
		}))
		Expect(tags).NotTo(HaveKey(mesh_proto.MeshTag))
	})

	It("returns an empty map for a legacy outbound with nil tags", func() {
		// given
		ds := meshroute.DestinationService{Outbound: &xds_types.Outbound{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Tags: nil,
			},
		}}

		// when
		tags := ds.OutboundListenerTags()

		// then
		Expect(tags).To(BeEmpty())
	})

	It("returns an empty map for a legacy outbound with empty tags", func() {
		// given
		ds := meshroute.DestinationService{Outbound: &xds_types.Outbound{
			LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
				Tags: map[string]string{},
			},
		}}

		// when
		tags := ds.OutboundListenerTags()

		// then
		Expect(tags).To(BeEmpty())
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
