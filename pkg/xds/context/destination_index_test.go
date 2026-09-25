package context_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("DestinationIndex", func() {
	Describe("GetReachableBackends", func() {
		It("should resolve labels and port to the MeshService port", func() {
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
				WithInboundOfTagsAndProtocol("http", "kuma.io/display-name", "web").
				Build()

			dp.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{}
			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind: "MeshService",
						Labels: map[string]string{
							mesh_proto.DisplayName:      "backend-svc",
							mesh_proto.KubeNamespaceTag: "other-ns",
						},
						Port: wrapperspb.UInt32(8080),
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

		It("should resolve display-name label derived from resource metadata", func() {
			ms := builders.MeshService().
				WithName("backend-svc").
				AddIntPort(9000, 9000, metadata.ProtocolHTTP).
				Build()

			dp := builders.Dataplane().
				WithName("dp-1").
				WithAddress("127.0.0.1").
				WithInboundOfTagsAndProtocol("http", "kuma.io/display-name", "web").
				Build()

			dp.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{}
			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind:   "MeshService",
						Labels: map[string]string{mesh_proto.DisplayName: "backend-svc"},
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

		It("should resolve user-defined outbound legacy MeshService name forms through derived labels", func() {
			ms := builders.MeshService().
				WithName("backend-svc-hash654").
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
				WithInboundOfTagsAndProtocol("http", "kuma.io/display-name", "web").
				AddOutbound(
					builders.Outbound().
						WithPort(10001).
						WithMeshService("backend-svc_other-ns_svc_8080", 8080),
				).
				Build()

			index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
			outbounds, matched := index.GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(HaveLen(1))
			Expect(outbounds).To(HaveKey(kri.WithSectionName(kri.From(ms), "8080")))
		})

		It("should resolve every MeshExternalService matching the labels", func() {
			var destinations []core_model.Resource
			for _, name := range []string{"api-a", "api-b", "api-c"} {
				destinations = append(destinations, builders.MeshExternalService().
					WithName(name).
					WithLabels(map[string]string{
						mesh_proto.KubeNamespaceTag: "kuma-system",
						"app":                       "api",
					}).
					Build())
			}

			dp := builders.Dataplane().
				WithName("dp-1").
				WithLabels(map[string]string{mesh_proto.KubeNamespaceTag: "beta"}).
				WithAddress("127.0.0.1").
				Build()
			dp.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{
				ReachableBackends: &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
					Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{{
						Kind:   "MeshExternalService",
						Labels: map[string]string{"app": "api"},
					}},
				},
			}

			outbounds, matched := xds_context.NewDestinationIndex(destinations).GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(HaveLen(3))
			for _, mes := range destinations {
				Expect(outbounds).To(HaveKey(kri.WithSectionName(kri.From(mes), "9000")))
			}
		})

		It("should resolve user-defined outbound MeshExternalService by name from another namespace", func() {
			mes := builders.MeshExternalService().
				WithName("api-cor-beta").
				WithLabels(map[string]string{mesh_proto.KubeNamespaceTag: "kuma-system"}).
				Build()

			dp := builders.Dataplane().
				WithName("dp-1").
				WithLabels(map[string]string{mesh_proto.KubeNamespaceTag: "beta"}).
				WithAddress("127.0.0.1").
				AddOutbound(builders.Outbound().WithPort(10001).WithMeshExternalService("api-cor-beta", 9000)).
				Build()

			outbounds, matched := xds_context.NewDestinationIndex([]core_model.Resource{mes}).GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(HaveKey(kri.WithSectionName(kri.From(mes), "9000")))
		})

		It("should not resolve when namespace label does not match", func() {
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
				WithInboundOfTagsAndProtocol("http", "kuma.io/display-name", "web").
				Build()

			dp.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{}
			dp.Spec.Networking.TransparentProxying.ReachableBackends = &mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackends{
				Refs: []*mesh_proto.Dataplane_Networking_TransparentProxying_ReachableBackendRef{
					{
						Kind: "MeshService",
						Labels: map[string]string{
							mesh_proto.DisplayName:      "backend-svc",
							mesh_proto.KubeNamespaceTag: "wrong-ns",
						},
					},
				},
			}

			index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
			outbounds, matched := index.GetReachableBackends(dp)

			Expect(matched).To(BeTrue())
			Expect(outbounds).To(BeEmpty())
		})

		Context("reachableBackends not set", func() {
			var ms core_model.Resource

			BeforeEach(func() {
				ms = builders.MeshService().
					WithName("backend-svc").
					AddIntPort(9000, 9000, metadata.ProtocolHTTP).
					Build()
			})

			DescribeTable("should resolve outbounds",
				func(withTPSection bool, allowAllOutbound bool, expectAll bool) {
					dp := builders.Dataplane().WithAddress("127.0.0.1").Build()
					if withTPSection {
						dp.Spec.Networking.TransparentProxying = &mesh_proto.Dataplane_Networking_TransparentProxying{}
					}

					index := xds_context.NewDestinationIndex([]core_model.Resource{ms}).WithAllowAllOutbound(allowAllOutbound)
					outbounds, matched := index.GetReachableBackends(dp)

					if expectAll {
						Expect(matched).To(BeFalse())
						Expect(outbounds).To(HaveKey(kri.WithSectionName(kri.From(ms), "9000")))
					} else {
						Expect(matched).To(BeTrue())
						Expect(outbounds).To(BeEmpty())
					}
				},
				Entry("deny by default", true, false, false),
				Entry("deny by default without transparentProxying section", false, false, false),
				Entry("allow all when allowAllOutbound is set", true, true, true),
				Entry("allow all without transparentProxying section when allowAllOutbound is set", false, true, true),
			)

			It("should return explicit outbounds", func() {
				dp := builders.Dataplane().
					WithAddress("127.0.0.1").
					AddOutbound(builders.Outbound().WithMeshService("backend-svc", 9000).WithPort(10001)).
					Build()

				index := xds_context.NewDestinationIndex([]core_model.Resource{ms})
				outbounds, matched := index.GetReachableBackends(dp)

				Expect(matched).To(BeTrue())
				Expect(outbounds).To(HaveLen(1))
				Expect(outbounds).To(HaveKey(kri.WithSectionName(kri.From(ms), "9000")))
			})
		})
	})
})
