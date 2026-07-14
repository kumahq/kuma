package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/xds/types"
)

var _ = Describe("Outbound", func() {
	Describe("ListenerTags", func() {
		It("returns legacy tags for legacy outbound", func() {
			o := &types.Outbound{
				LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Tags: map[string]string{
						mesh_proto.ServiceTag: "backend",
						"version":             "v1",
					},
				},
			}

			tags := o.ListenerTags()
			Expect(tags).To(Equal(map[string]string{
				mesh_proto.ServiceTag: "backend",
				"version":             "v1",
			}))
		})

		It("returns KRI tag for real resource outbound", func() {
			id := kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "zone-1",
				Namespace:    "ns",
				Name:         "backend",
				SectionName:  "http",
			}
			o := &types.Outbound{
				Resource: id,
			}

			tags := o.ListenerTags()
			Expect(tags).To(HaveKey(types.KRITag))
			Expect(tags[types.KRITag]).To(Equal(id.String()))
		})

		It("returns nil for empty outbound", func() {
			o := &types.Outbound{}

			tags := o.ListenerTags()
			Expect(tags).To(BeNil())
		})

		It("prefers legacy tags over KRI when both present", func() {
			id := kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         "default",
				Zone:         "zone-1",
				Namespace:    "ns",
				Name:         "backend",
				SectionName:  "http",
			}
			o := &types.Outbound{
				LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Tags: map[string]string{
						mesh_proto.ServiceTag: "backend",
					},
				},
				Resource: id,
			}

			tags := o.ListenerTags()
			Expect(tags).To(Equal(map[string]string{
				mesh_proto.ServiceTag: "backend",
			}))
		})
	})
})
