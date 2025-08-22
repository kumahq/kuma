package hooks_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/xds/hooks"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
)

var _ = Describe("ApiServerBypass", func() {
	It("should generate configuration for API Server bypass", func() {
		// given
		hook := hooks.NewApiServerBypass("1.1.1.1", 9090)
		rs := xds.NewResourceSet()
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Spec: &mesh_proto.Mesh{
						Networking: &mesh_proto.Networking{
							Outbound: &mesh_proto.Networking_Outbound{
								Passthrough: wrapperspb.Bool(false),
							},
						},
					},
				},
			},
		}
		proxy := &xds.Proxy{
			APIVersion: envoy.APIV3,
			Dataplane:  &mesh.DataplaneResource{},
		}

		// when
		Expect(hook.Modify(rs, ctx, proxy)).To(Succeed())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "api-server-bypass.yaml")))
	})

	It("should generate configuration for API Server bypass with unified naming", func() {
		// given
		hook := hooks.NewApiServerBypass("1.1.1.1", 9090)
		rs := xds.NewResourceSet()
		ctx := xds_context.Context{
			Mesh: xds_context.MeshContext{
				Resource: &mesh.MeshResource{
					Spec: &mesh_proto.Mesh{
						Networking: &mesh_proto.Networking{
							Outbound: &mesh_proto.Networking_Outbound{
								Passthrough: wrapperspb.Bool(false),
							},
						},
					},
				},
			},
		}
		proxy := &xds.Proxy{
			APIVersion: envoy.APIV3,
			Dataplane:  &mesh.DataplaneResource{},
			Metadata: &xds.DataplaneMetadata{
				Features: map[string]bool{
					types.FeatureUnifiedResourceNaming: true,
				},
			},
		}

		// when
		Expect(hook.Modify(rs, ctx, proxy)).To(Succeed())

		// then
		resp, err := rs.List().ToDeltaDiscoveryResponse()
		Expect(err).ToNot(HaveOccurred())

		actual, err := util_proto.ToYAML(resp)
		Expect(err).ToNot(HaveOccurred())

		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "api-server-bypass-unified-naming.yaml")))
	})
})
