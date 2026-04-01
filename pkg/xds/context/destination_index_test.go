package context_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	"github.com/kumahq/kuma/v2/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
)

var _ = Describe("DestinationIndex", func() {
	Describe("GetReachableBackends", func() {
		It("should resolve name/namespace format to correct MeshService", func() {
			ms := builders.MeshService().
				WithName("backend-svc-hash123").
				WithLabels(map[string]string{
					mesh_proto.DisplayName:      "backend-svc",
					mesh_proto.KubeNamespaceTag: "other-ns",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				AddIntPort(8080, 8080, metadata.ProtocolHTTP).
				Build()

			dp := builders.Dataplane().
				WithName("dp-1").
				WithLabels(map[string]string{
					mesh_proto.KubeNamespaceTag: "default",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				WithAddress("127.0.0.1").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				WithTransparentProxying(15001, 15006, "").
				Build()

			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind:      "MeshService",
						Name:      "backend-svc",
						Namespace: "other-ns",
						Port:      wrapperspb.UInt32(8080),
					},
				},
			}

			index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
			outbounds, matched := index.GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(HaveLen(1))

			expectedKRI := kri.WithSectionName(kri.From(ms), "8080")
			Expect(outbounds).To(HaveKey(expectedKRI))
		})

		It("should fallback to dataplane namespace when namespace not specified", func() {
			ms := builders.MeshService().
				WithName("backend-svc-hash456").
				WithLabels(map[string]string{
					mesh_proto.DisplayName:      "backend-svc",
					mesh_proto.KubeNamespaceTag: "default",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				AddIntPort(9000, 9000, metadata.ProtocolHTTP).
				Build()

			dp := builders.Dataplane().
				WithName("dp-1").
				WithLabels(map[string]string{
					mesh_proto.KubeNamespaceTag: "default",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				WithAddress("127.0.0.1").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				WithTransparentProxying(15001, 15006, "").
				Build()

			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind: "MeshService",
						Name: "backend-svc",
					},
				},
			}

			index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
			outbounds, matched := index.GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(HaveLen(1))

			expectedKRI := kri.WithSectionName(kri.From(ms), "9000")
			Expect(outbounds).To(HaveKey(expectedKRI))
		})

		It("should not resolve when namespace does not match", func() {
			ms := builders.MeshService().
				WithName("backend-svc-hash789").
				WithLabels(map[string]string{
					mesh_proto.DisplayName:      "backend-svc",
					mesh_proto.KubeNamespaceTag: "other-ns",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				AddIntPort(8080, 8080, metadata.ProtocolHTTP).
				Build()

			dp := builders.Dataplane().
				WithName("dp-1").
				WithLabels(map[string]string{
					mesh_proto.KubeNamespaceTag: "default",
					mesh_proto.ZoneTag:          "zone-1",
				}).
				WithAddress("127.0.0.1").
				WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http").
				WithTransparentProxying(15001, 15006, "").
				Build()

			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind:      "MeshService",
						Name:      "backend-svc",
						Namespace: "wrong-ns",
					},
				},
			}

			index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
			outbounds, matched := index.GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(BeEmpty())
		})
	})
})
